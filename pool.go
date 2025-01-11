package jar

import (
	"github.com/vulcand/oxy/v2/buffer"
	"github.com/vulcand/oxy/v2/forward"
	"github.com/vulcand/oxy/v2/roundrobin"

	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// PoolMaterializer is a function responsible for materializing a Pool
type PoolMaterializer func(*Pool) (http.Handler, error)

// MemberBuilder is a function responsible for building a Pool member if *Member is nil, or modifying the existing one
type MemberBuilder func(*PoolConfig, *url.URL, *Member) *Member

const (
	// ErrPoolsConfigdefaultmembererrorstatusInvalid is returned when the pools.defaultmembererrorstatus is set improperly
	ErrPoolsConfigdefaultmembererrorstatusInvalid = Error("pools.defaultmembererrorstatus is set to an invalid HealthCheckStatus")

	// ErrPoolsConfigdefaultmembererrorstatusEmpty is returned when the pools.defaultmembererrorstatus is empty
	ErrPoolsConfigdefaultmembererrorstatusEmpty = Error("pools.defaultmembererrorstatus is empty")

	// ErrPoolStickyAESNoKey is returned when materializing a Pool with StickyCookieType set to 'aes' but 'keys.stickycookie' is not set
	ErrPoolStickyAESNoKey = Error("Pool.StickyCookieType set to 'aes' but no keys.stickycookie set")

	// ErrPoolAddMemberNotSupported is returned when Pool.AddMember is called on a Pool that doesn't support the operation
	ErrPoolAddMemberNotSupported = Error("this Pool does not support dynamic adding of members")

	// ErrPoolDeleteMemberNotSupported is returned when Pool.DeleteMember is called on a Pool that doesn't support the operation
	ErrPoolDeleteMemberNotSupported = Error("this Pool does not support dynamic deletion of members")

	// ErrPoolRemoveMemberNotSupported is returned when Pool.RemoveMember is called on a Pool that doesn't support the operation
	ErrPoolRemoveMemberNotSupported = Error("this Pool does not support dynamic removing of members")

	// ErrPoolNoMembersConfigured is returned when a non-dynamic Pool type (e.g. S3) has no members configured
	ErrPoolNoMembersConfigured = Error("no members configured for a non-dynamic Pool")

	// ErrPoolConfigConsistentAndSticky is returned when a Pool has both Sticky and ConsistentHashing set
	ErrPoolConfigConsistentAndSticky = Error("a Pool cannot have Sticky and ConsistentHashing set")

	// ErrPoolConfigMissing is returned when an operation on a Pool is requested, but no config is set
	ErrPoolConfigMissing = Error("no Config present for Pool")
)

// Constants for configuration key strings
const (
	ConfigKeysStickyCookie           = ConfigKey("keys.stickycookie")
	ConfigStickyCookieAESTTL         = ConfigKey("stickycookie.aes.ttl")
	ConfigStickyCookieHTTPOnly       = ConfigKey("stickycookie.httponly")
	ConfigStickyCookieSecure         = ConfigKey("stickycookie.secure")
	ConfigConsistentHashPartitions   = ConfigKey("consistenthash.partitions")
	ConfigConsistentHashReplications = ConfigKey("consistenthash.replfactor")
	ConfigConsistentHashLoad         = ConfigKey("consistenthash.load")
)

var (
	// DefaultTrip should be used instead of the http.DefaultTransport, for pools/etc.
	DefaultTrip http.RoundTripper

	// DefaultClient should be used instead of using http.DefaultClient, for pools/etc.
	DefaultClient *http.Client

	// DefaultMemberWeight is the weight added to each member by default
	DefaultMemberWeight int
	// LocalMemberWeight is the weight assigned to each member that is AZ-local
	LocalMemberWeight int

	// ResponseModifierChain is a ProxyResponseModifierChain to handle sequences of modifications
	// use ``ResponseModifierChain.Add()`` to add your own
	ResponseModifierChain ProxyResponseModifierChain

	// Materializers is a map of available PoolMaterializers
	Materializers = make(map[string]PoolMaterializer)

	// MemberBuilders is a map of available MemberBuilders (should be one per Materializer)
	MemberBuilders = make(map[string][]MemberBuilder)
)

func init() {
	// Set up the Materializers
	Materializers["http"] = materializeHTTP
	Materializers["https"] = materializeHTTP
	Materializers["ws"] = materializeHTTP

	InitFuncs.Add(func() {
		// Defaults for any subrequests
		DefaultTransport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   Conf.GetDuration(ConfigTimeout),
				KeepAlive: Conf.GetDuration(ConfigKeepaliveTimeout),
				DualStack: false,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       3 * Conf.GetDuration(ConfigTimeout),
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		// Register with StrainFuncs, so we can free up resources if needed
		StrainFuncs.Add(DefaultTransport.CloseIdleConnections)

		DefaultTrip = DefaultTransport

		DefaultClient = &http.Client{
			Timeout:   time.Second * 30,
			Transport: DefaultTrip,
		}

		DefaultMemberWeight = Conf.GetInt(ConfigPoolsDefaultMemberWeight)
		LocalMemberWeight = Conf.GetInt(ConfigPoolsLocalMemberWeight)

	})

	ConfigValidations[ConfigPoolsDefaultMemberErrorStatus] = func() error {
		dmes := Conf.GetString(ConfigPoolsDefaultMemberErrorStatus)
		if dmes == "" {
			//DebugOut.Printf("pools.defaultmembererrorstatus is empty!\n")
			return ErrPoolsConfigdefaultmembererrorstatusEmpty
		}

		_, err := StringToHealthCheckStatus(dmes)
		if err != nil {
			//DebugOut.Printf("pools.defaultmembererrorstatus is set to an invalid HealthCheckStatus!\n")
			return ErrPoolsConfigdefaultmembererrorstatusInvalid
		}
		return nil
	}

	ConfigAdditions[ConfigPoolsDefaultMemberWeight] = 1
	ConfigAdditions[ConfigPoolsLocalMemberWeight] = 1000
	ConfigAdditions[ConfigPoolsDefaultMemberErrorStatus] = "Warning"
}

// Member is an attribute struct to describe a Pool Member
type Member struct {
	URL     *url.URL
	Address string
	AZ      string
	Weight  roundrobin.ServerOption
}

// NewMember returns a default Member
func NewMember(u *url.URL) *Member {
	return &Member{
		URL:     u,
		Address: u.Hostname(),
		Weight:  roundrobin.Weight(DefaultMemberWeight),
	}
}

// Pool is a list of like-minded destinations
type Pool struct {
	Config *PoolConfig

	members                sync.Map
	poolMaterializer       PoolMaterializer
	healthCheckErrorStatus HealthCheckStatus

	// AddMember adds a URI to the loadbalancer. An error is returned if the URI doesn't parse properly
	AddMember func(string) error
	// RemoveMember removes a URI from the loadbalancer, but not from the member cache.
	// ErrNoSuchMemberError is returned if the requested member doesn't exist,
	// or another error if the URI provided doesn't parse properly.
	RemoveMember func(string) error
	// DeleteMember removes a URI from the entire Pool construct,
	// ErrNoSuchMemberError is returned if the requested member doesn't exist,
	// or another error if the URI provided doesn't parse properly.
	DeleteMember func(string) error
	// ListMembers returns a list of URIs for existing members
	ListMembers func() []*url.URL

	// Materialized pool
	poollock sync.RWMutex
	pool     http.Handler
}

// NewPool returns a new, minimal, unmaterialized pool with the attached config
func NewPool(conf *PoolConfig) *Pool {
	p := Pool{
		Config: conf,
	}

	// Sets p.healthCheckErrorStatus to a valid HealthCheckStatus.
	// ConfigPoolsDefaultMemberErrorStatus can be assumed validated at this point
	if !p.Config.HealthCheckDisabled && p.Config.HealthCheckErrorStatus != "" {
		DebugOut.Printf("\t\tHealthCheckErrorStatus: %s\n", p.Config.HealthCheckErrorStatus)
		if hcs, serr := StringToHealthCheckStatus(p.Config.HealthCheckErrorStatus); serr != nil {
			ErrorOut.Printf("Non-fatal error materializing %s.HealthCheckErrorStatus of '%s', setting to default (%s)\n", p.Config.Name, p.Config.HealthCheckErrorStatus, Conf.GetString(ConfigPoolsDefaultMemberErrorStatus))

			// Default
			hcs, _ = StringToHealthCheckStatus(Conf.GetString(ConfigPoolsDefaultMemberErrorStatus))
			p.healthCheckErrorStatus = hcs
		} else {
			p.healthCheckErrorStatus = hcs
		}
	} else if !p.Config.HealthCheckDisabled {
		// Default
		hcs, _ := StringToHealthCheckStatus(Conf.GetString(ConfigPoolsDefaultMemberErrorStatus))
		p.healthCheckErrorStatus = hcs
	}

	return &p
}

// IsMaterialized return boolean on whether the pool has been materialized or not
func (p *Pool) IsMaterialized() bool {
	return p.pool != nil
}

// GetPool returns the materialized pool or an error. If the Pool has not been
// materialized, it does that.
func (p *Pool) GetPool() (http.Handler, error) {
	p.poollock.RLock()
	// We do not defer here, because if we have to call Materialize, it needs a WLock
	if p.pool == nil {
		p.poollock.RUnlock()
		return p.Materialize()
	}
	defer p.poollock.RUnlock()
	return p.pool, nil
}

// GetMember interacts with an internal cache, returning a Member from the cache or crafting a new one (and adding it to the cache)
func (p *Pool) GetMember(u *url.URL) *Member {

	if v, ok := p.members.Load(*u); ok {
		// We already have one
		return v.(*Member)
	}

	// We need to craft a Member
	m := NewMember(u)

	if v, ok := MemberBuilders[u.Scheme]; ok {
		for _, b := range v {
			m = b(p.Config, u, m)
		}
	} // else we just use the default

	p.members.Store(*u, m)
	return m
}

// Materialize returns a Handler that can represent the Pool.
//
// Generally, you should call Pool.GetPool instead, so you can receive
// a pointer to the exist materialized pool if it exists, or it will
// Materialize it for you.
func (p *Pool) Materialize() (http.Handler, error) {
	if p.poolMaterializer == nil {
		// No poolMaterializer was set, detect and set poolTypeID and poolMaterializer

		if p.Config == nil {
			return nil, fmt.Errorf("no Config present for Pool. Cannot materialize")
		}

		if len(p.Config.Members) < 1 {
			return nil, ErrPoolNoMembersConfigured
		}

		// We only take the first.
		member := p.Config.Members[0]

		memberURL, err := url.Parse(member)
		if err != nil {
			return nil, err
		}

		// Grab the URL scheme, and switch on it
		if v, ok := Materializers[memberURL.Scheme]; ok {
			p.poolMaterializer = v
		} else {
			// Um... no supported scheme?
			ErrorOut.Printf("FATAL: Materialization of Pool failed, Member scheme was %s", memberURL.Scheme)
			panic(fmt.Errorf("materialization of Pool failed, Member scheme was %s", memberURL.Scheme))
		}
	}

	h, err := p.poolMaterializer(p)
	if err != nil {
		return nil, err
	}

	// Write lock the pool briefly to set it
	p.poollock.Lock()
	p.pool = h
	p.poollock.Unlock()
	return h, nil
}

func materializeHTTP(p *Pool) (http.Handler, error) {
	var (
		fwd  *httputil.ReverseProxy
		err  error
		pool http.Handler
	)

	if p.Config.ReplacePath != "" {
		DebugOut.Printf("\t\tReplacePath: %s\n", p.Config.ReplacePath)
	}

	// Make a copy of the global request headers to strip
	pheaders := Conf.GetStringSlice(ConfigStripRequestHeaders)
	if len(p.Config.RemoveHeaders) > 0 {
		// Remove moar headers
		pheaders = append(pheaders, p.Config.RemoveHeaders...)
	}
	rw := reqRewriter{Headers: pheaders, To: p.Config.ReplacePath, StripPrefix: p.Config.StripPrefix}

	fwd = forward.New(true)
	fwd.ErrorLog = ErrorOut
	fwd.ModifyResponse = ResponseModifierChain.ToProxyResponseModifier()
	fwd.Transport = DefaultTrip

	urlcapture := URLCaptureHandler(rw.Handler(fwd))

	if p.Config.Sticky && p.Config.ConsistentHashing {
		// Mutually exclusive
		return nil, ErrPoolConfigConsistentAndSticky
	}

	// Build a PoolManager
	var (
		pm    PoolManager
		pmErr error
	)
	if p.Config.Sticky {
		// Pool is doing sticky load-balancing
		pm, pmErr = p.materializeSticky(urlcapture, roundrobin.Logger(&oxyLogger))
	} else if p.Config.ConsistentHashing {
		// Pool is using a consistent hash to direct traffics
		pm, pmErr = p.materializeConsistent(urlcapture)
	} else {
		// Pool is not not sticky nor consistent, so standard rrlb
		pm, pmErr = roundrobin.New(urlcapture, roundrobin.Logger(&oxyLogger))
	}
	if pmErr != nil {
		return nil, pmErr
	}

	// Define ListMembers
	p.ListMembers = func() []*url.URL {
		return pm.Servers()
	}

	// Define AddMember
	p.AddMember = func(member string) error {
		u, uerr := url.Parse(member)
		if uerr != nil {
			return uerr
		}

		m := p.GetMember(u)
		uerr = pm.UpsertServer(u, m.Weight)
		if uerr != nil {
			return uerr
		}
		return nil
	}

	// Define DeleteMember
	p.DeleteMember = func(member string) error {
		u, uerr := url.Parse(member)
		if uerr != nil {
			return uerr
		}

		// If the member has been materialized, remove it from the cache
		p.members.Delete(*u)

		uerr = pm.RemoveServer(u)
		if uerr != nil {
			if uerr.Error() == "server not found" { // Bad, M@. BAD. M@.
				return ErrNoSuchMemberError
			}
			return uerr
		}

		return nil
	}

	// Define RemoveMember
	p.RemoveMember = func(member string) error {
		u, uerr := url.Parse(member)
		if uerr != nil {
			return uerr
		}

		uerr = pm.RemoveServer(u)
		if uerr != nil {
			if uerr.Error() == "server not found" { // Bad, M@. BAD. M@.
				return ErrNoSuchMemberError
			}
			return uerr
		}

		return nil
	}

	// Add members
	for _, member := range p.Config.Members {
		DebugOut.Printf("\t\tAdding member '%s'\n", member)
		err = p.AddMember(member)
		if err != nil {
			return nil, err
		}
	}

	// Buffer all the requests
	if p.Config.Buffered {
		DebugOut.Printf("\t\tBuffering with %d retries.\n", p.Config.BufferedFails)
		buff, err := buffer.New(pm, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", p.Config.BufferedFails)), buffer.Logger(&oxyLogger))
		if err != nil {
			return nil, err
		}
		pool = buff
	} else {
		pool = pm
	}

	return pool, nil
}

// reqRewriter is a forward.ReqRewriter, that removes headers and/or mangles the request URI
type reqRewriter struct {
	// Headers is a list of headers to remove from the request
	Headers []string
	// To sets the request URI path.
	// Mutually exclusive with StripPrefix
	To string
	// StripPrefix removes the prefix from the request URI if present.
	// Mutually exclusive with To
	StripPrefix string
}

// Rewrite remove headers from a request
func (h *reqRewriter) Rewrite(r *http.Request) {

	DebugOut.Print(ErrRequestError{r, fmt.Sprintf("reqRewriter firing! %+v\n", h)}.String())
	hr := forward.HeaderRewriter{TrustForwardHeader: true, Hostname: Hostname}
	hr.Rewrite(r)

	for _, header := range h.Headers {
		r.Header.Del(header)
		if header == "Host" {
			// Additionally, swap out the Request.Host
			r.Host = r.URL.Host
		}
	}
	//DebugOut.Printf(ErrRequestError{r, fmt.Sprintf("reqRewriter Headers: %v URL.Host: %s Request.Host %s\n", r.Header, r.URL.Host, r.Host)}.String())

	if h.StripPrefix != "" {
		TrimPrefixURI(r, h.StripPrefix)
	} else if h.To != "" {
		ReplaceURI(r, h.To, h.To)
	}

}

// Handler is an http.Handler to wrap the request rewriter.
func (h *reqRewriter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.Rewrite(r)
		next.ServeHTTP(w, r)
	})
}
