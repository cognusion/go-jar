package jar

import (
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/spf13/cast"

	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Constants for configuration key strings
const (
	ConfigCompression     = ConfigKey("compression")
	ConfigDisableRealAddr = ConfigKey("disablerealaddr")
	ConfigForbiddenPaths  = ConfigKey("forbiddenpaths")
)

// Path is an extensible struct, detailing its configuration
type Path struct {
	// Name is an optional "name" for the path. Will be output in some logs. If not set, will use an index number
	Name string
	// Path is a URI prefix to match
	Path string
	// Absolute declares if Path should be absolute instead of as a prefix
	Absolute bool
	// Allow
	Allow string
	// Deny
	Deny string
	// Host is a hostname or hostname-pattern to restrict this Path too
	Host string
	// Hosts is a list of hostnames or hostname-patterns to restrict this Path too.
	// Will result in one actual Path per entry, which is almost always fine.
	Hosts []string
	// Methods is a list of HTTP methods to restrict this path to
	Methods []string
	// Headers is a list of HTTP Request headers to restrict this path to
	Headers []string
	// Handlers is an ordered list of http.Handlers to apply
	Handlers []string
	// Pool is an actual Pool to handle the proxying. Mutually exclusive with Finisher
	Pool string
	// Finisher is the final handler. Mutually exclusive with Pool
	Finisher string
	// RateLimit each IP to these many requests/second. Also must have the "RateLimiter" handler, or it will be appended to the chain
	RateLimit float64
	// RateLimitPurge is a duration where a limit gets dumped
	RateLimitPurge time.Duration
	// RateLimitCollectOnly sets if the ratelimiter should only collect and log, versus enforce
	RateLimitCollectOnly bool
	// BodyByteLimit is the maximum number of bytes a Request.Body is allowed to be. It is poor form to set this unless the Path is terminated by
	// a finisher that will otherwise consume the Request.Body and possibly OOM and/or overuse disk space.
	BodyByteLimit int64
	// Redirect is a special Finisher. "%1" may be used to optionally denote the request path.
	// e.g. Redirect http://somewhereelse.com%1
	Redirect string
	// RedirectCode is an optional code to send as the redirect status
	RedirectCode int
	// RedirectHostMatch is a Perl-Compatible Regular Expression with grouping to apply to the Hostname, replacing $1,$2, etc. in ``Redirect``
	RedirectHostMatch string
	// ReplacePath is used to replace the requested path with the target path
	ReplacePath string
	// StripPrefix is used to replace the requested path with one sans prefix
	StripPrefix string
	// BrowserExclusions is a list of browsers disallowed down this path, based on best-effort analysis of request headers
	BrowserExclusions []string
	// ForbiddenPaths is a list of path prefixes that will result in a 403, while traversing this path
	ForbiddenPaths []string
	// Timeout is a path-specific override of how long a request and response may take on this path
	Timeout time.Duration
	// BasicAuthRealm is the name of the HTTP Auth Realm on this Path. Need not be unique. Should not be empty.
	BasicAuthRealm string
	// BasicAuthSource is a URL to specify where HTTP Basic Auth information should come from (file://). Setting this forces auth
	BasicAuthSource string
	// BasicAuthUsers is a list of usernames allowed on this Path. Default is "all"
	BasicAuthUsers []string
	// ErrorMessage is a static message to respond with, if this path is executed
	ErrorMessage string
	// ErrorCode is the HTTP response code that will be returned with ErrorMessage, IFF ErrorMessage is set. Defaults to StatusOK
	ErrorCode int
	// Options is a horrible, brittle map[string]interface{} that some handlers or finishers
	// use for per-path configuration. Avoid if possible.
	Options PathOptions
}

// PathOptions is an MSI with a case-agnostic getter
type PathOptions map[string]interface{}

// Get returns an interface{} if *key* matches, otherwise nil
func (p *PathOptions) Get(key string) interface{} {
	if p == nil {
		return nil
	}
	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return v
		}
	}
	return nil
}

// GetString returns a string if *key* matches, otherwise empty string
func (p *PathOptions) GetString(key string) string {
	if p == nil {
		return ""
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToString(v)
		}
	}
	return ""
}

// GetBool returns a bool value if *key* matches, otherwise false
func (p *PathOptions) GetBool(key string) bool {
	if p == nil {
		return false
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToBool(v)
		}
	}
	return false
}

// GetStringSlice returns a []string if *key* matches, otherwise an empty []string
func (p *PathOptions) GetStringSlice(key string) []string {
	if p == nil {
		return make([]string, 0)
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToStringSlice(v)
		}
	}
	return make([]string, 0)
}

// BuildPaths unmarshalls the paths config, creates handler chains, and updates the mux
func BuildPaths(router *mux.Router) error {
	if ipaths := Conf.Get(ConfigPaths); ipaths != nil {
		paths := make([]Path, len(ipaths.([]interface{})))
		Conf.UnmarshalKey(ConfigPaths, &paths)

		// Range over the paths
		pcount := 0
		for _, path := range paths {
			lastindex, err := BuildPath(path, pcount, router)
			if err != nil {
				return err
			}
			if lastindex > 0 {
				// if a listindex was returned, then probable some
				// other paths were created since our last iteration.
				pcount = lastindex
			}
			pcount++
		}
	}
	return nil
}

// BuildPath does the heavy lifting to build a single path (which may result in multiple paths, but that's just bookkeeping)
func BuildPath(path Path, index int, router *mux.Router) (int, error) {

	if len(path.Hosts) > 0 {
		DebugOut.Printf("Multihost Path %s: %+v\n", path.Path, path.Hosts)
		for _, host := range path.Hosts {

			tpath := path
			tpath.Host = host
			tpath.Hosts = make([]string, 0)

			lastindex, err := BuildPath(tpath, index, router)
			if err != nil {
				return 0, err
			}
			if lastindex > 0 {
				index = lastindex
			}
			index++
		}
		return index - 1, nil
	}

	DebugOut.Printf("Path %s (index %d)\n", path.Path, index)

	var (
		hchain alice.Chain
		b      PathHandler
	)
	if path.Name == "" {
		b.Path = strconv.Itoa(index)
	} else {
		b.Path = path.Name
	}
	if len(path.Options) > 0 {
		b.Options = path.Options
	}
	hchain = hchain.Append(b.Handler)

	if path.Pool != "" {
		p := PoolID{path.Pool}
		hchain = hchain.Append(p.Handler)
	}

	// Let's build a Route for this Path
	var pathRouter *mux.Route
	if path.Absolute {
		DebugOut.Print("\tAbsolute\n")
		pathRouter = router.Path(path.Path)
	} else {
		pathRouter = router.PathPrefix(path.Path)
	}

	// Load Host restrictions
	if path.Host != "" {
		DebugOut.Printf("\tHost: %s\n", path.Host)
		pathRouter.Host(path.Host)
	}

	// Load Method restrictions
	if len(path.Methods) > 0 {
		DebugOut.Printf("\tMethods: %+v\n", path.Methods)
		pathRouter.Methods(path.Methods...)
	}

	// Load Header restrictions
	if len(path.Headers) > 0 {
		headers := make([]string, len(path.Headers)*2)

		for i, header := range path.Headers {
			DebugOut.Printf("\tHeader: %s\n", header)
			hparts := strings.SplitN(header, " ", 2)
			if len(hparts) != 2 {
				return 0, ErrConfigurationError{"Path Header appears malformed. No space?"}
			}
			headers[i+i] = hparts[0]
			headers[i+i+1] = hparts[1]
		}
		pathRouter.HeadersRegexp(headers...)
	}

	var (
		timeoutterFound bool // false
	)

	// Automatically load SetupHandler
	DebugOut.Printf("\tAdding %s\n", "SetupHandler")
	hchain = hchain.Append(SetupHandler)

	// Load Compression handler, maybe
	if c := Conf.GetStringSlice(ConfigCompression); len(c) > 0 {
		DebugOut.Printf("\tAdding Compression\n")
		ch := NewCompression(c)
		hchain = hchain.Append(ch.Handler)
	}

	// Automatically load RealAddr and ResponseHeaders, maybe
	if ok := Conf.GetBool(ConfigDisableRealAddr); !ok {
		DebugOut.Printf("\tAdding %s\n", "RealAddr")
		hchain = hchain.Append(RealAddr)
	}
	if h := Conf.GetStringSlice(ConfigHeaders); len(h) > 0 {
		DebugOut.Printf("\tAdding %s\n", "ResponseHeaders")
		hchain = hchain.Append(ResponseHeaders)
	}

	// Automatically load AccessLogHandler, always
	DebugOut.Printf("\tAdding %s\n", "AccessLogHandler")
	hchain = hchain.Append(AccessLogHandler)

	// Automatically load AuthoritativeDomainsHandler
	DebugOut.Printf("\tAdding %s\n", "AuthoritativeDomainsHandler")
	hchain = hchain.Append(AuthoritativeDomainsHandler)

	// Automatically load Access handler, maybe
	if path.Allow != "" || path.Deny != "" {
		DebugOut.Printf("\tAdding Access: Allow '%s', Deny '%s'\n", path.Allow, path.Deny)
		a, err := NewAccessFromStrings(path.Allow, path.Deny)
		if err != nil {
			return 0, err
		}
		hchain = hchain.Append(a.AccessHandler)
	}

	// Automatically load the RateLimit handler, maybe
	if path.RateLimit > 0 {
		if path.RateLimitPurge == 0 {
			path.RateLimitPurge = time.Hour
		}
		DebugOut.Printf("\t\tRate Limit: %f\n", path.RateLimit)

		var rl RateLimiter
		if path.RateLimitCollectOnly {
			rl = NewRateLimiterCollector(path.RateLimit, path.RateLimitPurge)
		} else {
			rl = NewRateLimiter(path.RateLimit, path.RateLimitPurge)
		}
		hchain = hchain.Append(rl.Handler)
	}

	// Automatically load the BasicAuth handler, maybe
	if path.BasicAuthSource != "" {
		DebugOut.Printf("\tAdding %s (%s) @ %s\n", "BasicAuth.Handler", path.BasicAuthRealm, path.BasicAuthSource)
		for _, u := range path.BasicAuthUsers {
			DebugOut.Printf("\t\tUser '%s'\n", u)
		}

		b, err := NewVerifiedBasicAuth(path.BasicAuthSource, path.BasicAuthRealm, path.BasicAuthUsers)
		if err != nil {
			return 0, ErrConfigurationError{"BasicAuth source could not be verified"}
		}

		hchain = hchain.Append(b.handler)
	}

	// Automatically load the BodyByteLimit handler, maybe
	if path.BodyByteLimit > 0 {
		DebugOut.Printf("\tAdding BodyByteLimit(%d) handler\n", path.BodyByteLimit)
		bbl := NewBodyByteLimit(path.BodyByteLimit)
		hchain = hchain.Append(bbl.Handler)
	}

	// Load global handlers
	for _, handler := range Conf.GetStringSlice(ConfigHandlers) {
		DebugOut.Printf("\tAdding %s\n", handler)

		lchandler := strings.ToLower(handler)

		if lchandler == ConfigTimeout {
			timeoutterFound = true

			if path.Timeout != 0 {
				// Local timeout
				t := Timeout{
					Duration: path.Timeout,
				}
				DebugOut.Printf("\t\tTimeout (Path): %s\n", path.Timeout.String())
				hchain = hchain.Append(t.Handler)
			} else if gt := Conf.GetDuration(ConfigTimeout); gt != 0 {
				// Global timeout
				t := Timeout{
					Duration: gt,
				}
				DebugOut.Printf("\t\tTimeout (Global): %s\n", gt.String())
				hchain = hchain.Append(t.Handler)
			} else {
				return 0, ErrConfigurationError{"timeout handler inline, but no timelimit set globally or on path!"}
			}

			continue
		}

		var err error
		hchain, err = HandleHandler(handler, hchain)
		if err != nil {
			DebugOut.Printf("\t\tFailed: %s\n", err)
			return 0, err
		}

	}

	// Load CORS handler, maybe
	if c := Conf.GetStringSlice(ConfigCORSOrigins); len(c) > 0 {
		DebugOut.Printf("\tAdding CORS\n")
		hchain = hchain.Append(CorsHandler)
	}

	// Load ForbiddenPaths, maybe
	{
		fpaths := Conf.GetStringSlice(ConfigForbiddenPaths)
		if len(path.ForbiddenPaths) > 0 {
			fpaths = append(fpaths, path.ForbiddenPaths...)
		}
		if len(fpaths) > 0 {
			DebugOut.Printf("\tAdding ForbiddenPaths\n")
			fp, err := NewForbiddenPaths(fpaths)
			if err != nil {
				return 0, err
			}
			hchain = hchain.Append(fp.Handler)
		}
	}

	// Load path-specific handlers
	for _, handler := range path.Handlers {
		DebugOut.Printf("\tAdding %s\n", handler)

		lchandler := strings.ToLower(handler)

		if lchandler == ConfigTimeout {
			timeoutterFound = true

			if path.Timeout != 0 {
				// Local timeout
				t := Timeout{
					Duration: path.Timeout,
				}
				DebugOut.Printf("\t\tTimeout (Path): %s\n", path.Timeout.String())
				hchain = hchain.Append(t.Handler)
			} else if gt := Conf.GetDuration(ConfigTimeout); gt != 0 {
				// Global timeout
				t := Timeout{
					Duration: gt,
				}
				DebugOut.Printf("\t\tTimeout (Global): %s\n", gt.String())
				hchain = hchain.Append(t.Handler)
			} else {
				return 0, ErrConfigurationError{"timeout handler inline, but no timelimit set globally or on path!"}
			}

			continue
		}

		var err error
		hchain, err = HandleHandler(handler, hchain)
		if err != nil {
			DebugOut.Printf("\t\tFailed: %s\n", err)
			return 0, err
		}
	}

	// Load PathReplacer maybe
	if path.ReplacePath != "" {
		DebugOut.Printf("\tAdding PathReplacer to '%s'\n", path.ReplacePath)
		pr := PathReplacer{
			From: path.Path,
			To:   path.ReplacePath,
		}
		hchain = hchain.Append(pr.Handler)
	}

	// Load PathStripper maybe
	if path.StripPrefix != "" {
		DebugOut.Printf("\tAdding PathStripper to '%s'\n", path.ReplacePath)
		pr := PathStripper{
			Prefix: path.StripPrefix,
		}
		hchain = hchain.Append(pr.Handler)
	}

	// Safety check - if timeout is set, but the handler wasn't declared, append it
	if !timeoutterFound && (path.Timeout != 0 || Conf.GetDuration(ConfigTimeout) != 0) {
		if path.Timeout != 0 {
			// Local timeout
			t := Timeout{
				Duration: path.Timeout,
			}
			DebugOut.Printf("\tAppending Timeout.Handler: Timeout (Path): %s\n", path.Timeout.String())
			hchain = hchain.Append(t.Handler)
		} else if gt := Conf.GetDuration(ConfigTimeout); gt != 0 {
			// Global timeout
			t := Timeout{
				Duration: gt,
			}
			DebugOut.Printf("\tAppending Timeout.Handler:Timeout (Global): %s\n", gt.String())
			hchain = hchain.Append(t.Handler)
		}
	}

	// Load endpoint handlers
	var pathHandler http.Handler

	switch {
	case path.Redirect != "":
		// path will be redirected
		DebugOut.Printf("\tAdding Redirect %s\n", path.Redirect)

		p := Redirect{URL: path.Redirect, Code: path.RedirectCode, PCRE: nil}
		if path.RedirectHostMatch != "" {
			DebugOut.Printf("\t\tPCRE: %s\n", path.RedirectHostMatch)
			// Test the PCRE
			re, rerr := pcre.Compile(path.RedirectHostMatch, pcre.CASELESS)
			if rerr != nil {
				DebugOut.Printf("\t\tFailed: %s\n", rerr)
				return 0, ErrConfigurationError{rerr.String()}
			}
			if re.Groups() < 1 {
				return 0, ErrConfigurationError{"RedirectHostMatch has no groups"}
			}
			gcount := 0
			for i := 0; i < re.Groups(); i++ {
				s := fmt.Sprintf("$%d", i+1) // $1 $2 $3 etc
				if strings.Contains(path.Redirect, s) {
					gcount++
				}
			}
			if gcount != re.Groups() {
				return 0, ErrConfigurationError{fmt.Sprintf("RedirectHostMatch has %d groups, but Redirect has %d", re.Groups(), gcount)}
			}

			p = Redirect{URL: path.Redirect, Code: path.RedirectCode, PCRE: &re}
		}

		pathHandler = hchain.ThenFunc(p.Finisher)

	case path.ErrorMessage != "":
		// path will have a static error
		DebugOut.Printf("\tAdding ErrorMessage '%s' and ErrorCode '%d'\n", path.ErrorMessage, path.ErrorCode)
		p := GenericResponse{path.ErrorMessage, path.ErrorCode}
		pathHandler = hchain.ThenFunc(p.Finisher)

	case path.Pool != "":
		// path will be proxied, with the Pool
		if p, ok := LoadBalancers.Get(path.Pool); ok {
			DebugOut.Printf("\tAdding Pool %s\n", p.Config.Name)
			pool, err := p.GetPool()
			if err != nil {
				return 0, ErrConfigurationError{fmt.Sprintf("pool '%s' had an error materializing: %s", path.Pool, err)}
			}
			pathHandler = hchain.Then(pool)
		} else {
			// Pool doesn't exist
			return 0, ErrConfigurationError{fmt.Sprintf("pool '%s' is not a listed pool", path.Pool)}
		}

	case path.Finisher != "":
		// path will be handled by Finisher
		if l, err := HandleFinisher(path.Finisher); err == nil {
			DebugOut.Printf("\tAdding Finisher %s\n", path.Finisher)
			pathHandler = hchain.Then(l)
		} else if err == ErrFinisher404 {
			// Finisher doesn't exist
			return 0, ErrConfigurationError{fmt.Sprintf("finisher '%s' is not a listed finisher handler", path.Finisher)}
		} else {
			// Other finisher error
			return 0, ErrConfigurationError{fmt.Sprintf("finisher '%s' exists but retuned error: %v", path.Finisher, err)}
		}

	default:
		// No Pool, no Finisher? Problem.
		return 0, ErrConfigurationError{"path must have Redirect, ErrorMessage, Pool, or Finisher defined"}
	}

	pathRouter.Handler(pathHandler)

	return index, nil
}
