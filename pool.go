package jar

import (
	"github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/buffer"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"

	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

const (
	httpPool poolType = iota + 10
	s3Pool
	wsPool
)

type poolType int

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

	// ResponseModiferChain is a ProxyResponseModifierChain to handle sequences of modifications
	// use ``ResponseModiferChain.Add()`` to add your own
	ResponseModiferChain ProxyResponseModifierChain
)

func init() {

	InitFuncs.Add(func() {
		// TODO: Tighten up these Defaults!
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
	weight  roundrobin.ServerOption
}

// Pool is a list of like-minded destinations
type Pool struct {
	Config *PoolConfig

	members                sync.Map
	poolTypeID             poolType
	poolMaterializer       func() (http.Handler, error)
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
	m := p.buildMember(u)
	p.members.Store(*u, m)
	return m
}

// buildMember does the heavy-lifting of [re]building a Member from a URL. This should never be called directly, and GetMember should be called instead
func (p *Pool) buildMember(u *url.URL) *Member {

	m := Member{
		URL:     u,
		Address: u.Hostname(),
		weight:  roundrobin.Weight(DefaultMemberWeight),
	}

	// If we're EC2-aware, and the member is for an S3 bucket
	if AWSSession != nil && u.Scheme == "s3" {
		return &m
	}

	// If we're EC2-aware, and this Pool is using EC2Affinity, let's find out what the AZ is
	if AWSSession != nil && Conf.GetBool(ConfigEC2) && p.Config.EC2Affinity {

		if !p.Config.Prune {
			// Not forcing this, but chances are you don't know what you're doing.
			ErrorOut.Printf("WARNING!!! Pool %s is using EC2Affinity but not Prune. This may delay or prevent expected failover to non-local members in the event of a member failure.\n", p.Config.Name)
			DebugOut.Printf("WARNING!!! Pool %s is using EC2Affinity but not Prune. This may delay or prevent expected failover to non-local members in the event of a member failure.\n", p.Config.Name)
		}

		// If the Hostname isn't just digits and dots, it's a name and not a number, make it a number
		if ok, err := regexp.MatchString(`[^\d\.]`, u.Hostname()); err == nil && ok {
			// hostname is probably not an address
			addrs, err := net.LookupHost(u.Hostname())
			if err != nil {
				DebugOut.Printf("Error resolving hostname '%s': %s\n", u.Hostname(), err)
			} else if len(addrs) > 0 {
				// Take the first address
				m.Address = addrs[0]
			}
		}

		if az, azerr := AWSSession.GetInstanceAZByIP(m.Address); azerr != nil {
			ErrorOut.Printf("Error adding EC2-aware pool-member '%s' to %s: %s\n", m.Address, p.Config.Name, azerr)
		} else if az == "" {
			DebugOut.Printf("\t\t\tPool %s has member %s that has no AZ\n", p.Config.Name, m.Address)
		} else if az == AWSSession.Me.AvailabilityZone {
			DebugOut.Printf("\t\t\tPool %s has member %s that is AZ-local!\n", p.Config.Name, m.Address)
			m.weight = roundrobin.Weight(LocalMemberWeight)
			m.AZ = az
		} else {
			DebugOut.Printf("\t\t\tPool %s has member %s that is not AZ-local (%s)\n", p.Config.Name, m.Address, az)
			m.AZ = az
		}
	}
	return &m
}

// Materialize returns a Handler that can represent the Pool.
//
// Generally, you should call Pool.GetPool instead, so you can receive
// a pointer to the exist materialized pool if it exists, or it will
// Materialize it for you.
func (p *Pool) Materialize() (http.Handler, error) {
	if p.poolMaterializer == nil {
		// No poolMaterializer was set, detect and set poolTypeID and poolMaterializer
		if len(p.Config.Members) < 1 {
			return nil, fmt.Errorf("no Members configured for Pool")
		}

		// We only take the first.
		member := p.Config.Members[0]

		memberURL, err := url.Parse(member)
		if err != nil {
			return nil, err
		}

		// Grab the URL scheme, and switch on it
		switch memberURL.Scheme {
		case "http":
			fallthrough
		case "https":
			p.poolTypeID = httpPool
			p.poolMaterializer = p.materializeHTTP
		case "s3":
			p.poolTypeID = s3Pool
			p.poolMaterializer = p.materializeS3
		case "ws":
			p.poolTypeID = wsPool
			p.poolMaterializer = p.materializeHTTP
		default:
			// Um... no supported scheme?
			ErrorOut.Printf("FATAL: Materialization of Pool failed, Member scheme was %s", memberURL.Scheme)
			panic(fmt.Errorf("materialization of Pool failed, Member scheme was %s", memberURL.Scheme))
		}
	}

	return p.poolMaterializer()
}

func (p *Pool) materializeS3() (http.Handler, error) {

	// Define ListMembers
	p.ListMembers = func() []*url.URL {
		return nil
	}

	// Define AddMember
	p.AddMember = func(member string) error {
		return ErrPoolAddMemberNotSupported
	}

	// Define DeleteMember
	p.DeleteMember = func(member string) error {
		return ErrPoolDeleteMemberNotSupported
	}

	// Define RemoveMember
	p.RemoveMember = func(member string) error {
		return ErrPoolRemoveMemberNotSupported
	}

	// Add members
	if len(p.Config.Members) < 1 {
		return nil, fmt.Errorf("no Members configured for Pool")
	}

	// We only take the first.
	member := p.Config.Members[0]

	memberURL, err := url.Parse(member)
	if err != nil {
		return nil, err
	}

	DebugOut.Printf("\t\tAdding member '%s'\n", member)
	p.GetMember(memberURL)

	// Add it to
	pool, err := NewS3Pool(member)
	if err != nil {
		return nil, err
	}

	// Write lock the pool briefly to set it
	p.poollock.Lock()
	p.pool = pool
	p.poollock.Unlock()
	return pool, nil
}

func (p *Pool) materializeHTTP() (http.Handler, error) {
	var (
		fwd  *forward.Forwarder
		err  error
		pool http.Handler
	)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	if Conf.GetBool(ConfigDebug) {
		hook := loggerHook{
			Log:  DebugOut,
			Name: p.Config.Name,
		}
		hook.AddLevels(logrus.AllLevels)
		logrusLogger.AddHook(&hook)
		logrusLogger.SetLevel(logrus.DebugLevel)
	}

	if p.Config.ReplacePath != "" {
		DebugOut.Printf("\t\tReplacePath: %s\n", p.Config.ReplacePath)
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

	// Make a copy of the global request headers to strip
	pheaders := Conf.GetStringSlice(ConfigStripRequestHeaders)
	if len(p.Config.RemoveHeaders) > 0 {
		// Remove moar headers
		pheaders = append(pheaders, p.Config.RemoveHeaders...)
	}
	fwd, err = forward.New(forward.Logger(logrusLogger), forward.PassHostHeader(true), forward.Rewriter(&reqRewriter{Headers: pheaders, To: p.Config.ReplacePath, StripPrefix: p.Config.StripPrefix}), forward.ResponseModifier(ResponseModiferChain.ToProxyResponseModifier()), forward.RoundTripper(DefaultTrip))
	if err != nil {
		return nil, err
	}

	urlcapture := URLCaptureHandler(fwd)

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
		pm, pmErr = p.materializeSticky(urlcapture, roundrobin.RoundRobinLogger(logrusLogger))
	} else if p.Config.ConsistentHashing {
		// Pool is using a consistent hash to direct traffics
		pm, pmErr = p.materializeConsistent(urlcapture)
	} else {
		// Pool is not not sticky nor consistent, so standard rrlb
		pm, pmErr = roundrobin.New(urlcapture, roundrobin.RoundRobinLogger(logrusLogger))
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
		uerr = pm.UpsertServer(u, m.weight)
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
		buff, err := buffer.New(pm, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", p.Config.BufferedFails)), buffer.Logger(logrusLogger))
		if err != nil {
			return nil, err
		}
		pool = buff
	} else {
		pool = pm
	}

	// Write lock the pool briefly to set it
	p.poollock.Lock()
	p.pool = pool
	p.poollock.Unlock()
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

	DebugOut.Printf(ErrRequestError{r, fmt.Sprintf("reqRewriter firing! %+v\n", h)}.String())
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
