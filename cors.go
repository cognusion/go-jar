package jar

import (
	"github.com/cognusion/go-prw"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"

	"fmt"
	"net/http"
)

// Constants for configuration key strings
const (
	ConfigCORSAllowCredentials = ConfigKey("CORS.allowcredentials")
	ConfigCORSAllowHeaders     = ConfigKey("CORS.allowheaders")
	ConfigCORSAllowMethods     = ConfigKey("CORS.allowmethods")
	ConfigCORSOrigins          = ConfigKey("CORS.origins")
	ConfigCORSMaxAge           = ConfigKey("CORS.maxage")

	CORSAllowOrigin      = CorsString("Access-Control-Allow-Origin")
	CORSAllowCredentials = CorsString("Access-Control-Allow-Credentials")
	CORSExposeHeaders    = CorsString("Access-Control-Expose-Headers")
	CORSAllowMethods     = CorsString("Access-Control-Allow-Methods")
	CORSAllowHeaders     = CorsString("Access-Control-Allow-Headers")
	CORSMaxAge           = CorsString("Access-Control-Max-Age")
)

var (
	// CorsHandler is the global handler for CORS
	CorsHandler func(next http.Handler) http.Handler
)

// CorsString is a string type for static string consistency
type CorsString = string

func init() {

	// Set up the configs
	ConfigAdditions[ConfigCORSOrigins] = make([]string, 0) // A list of acceptable origins/patterns
	ConfigAdditions[ConfigCORSAllowHeaders] = ""           // A verbatim string to put in Expose/Allow headers
	ConfigAdditions[ConfigCORSAllowMethods] = ""           // A verbatim string to put in Allow methods
	ConfigAdditions[ConfigCORSAllowCredentials] = "false"  // A verbatim string to put in Allow credentials
	ConfigAdditions[ConfigCORSMaxAge] = "60"               // A verbatim string to put in Max age

	InitFuncs.Add(func() {
		// Compile the CORS.origins expressions, for speed
		if len(Conf.GetStringSlice(ConfigCORSOrigins)) > 0 {
			cmap := map[string]string{
				"allowheaders":     Conf.GetString(ConfigCORSAllowHeaders),
				"allowmethods":     Conf.GetString(ConfigCORSAllowMethods),
				"allowcredentials": Conf.GetString(ConfigCORSAllowCredentials),
				"maxage":           Conf.GetString(ConfigCORSMaxAge),
			}
			c, err := NewCORSFromConfig(Conf.GetStringSlice(ConfigCORSOrigins), cmap)
			if err != nil {
				ErrorOut.Fatalln(err)
			}
			CorsHandler = c.Handler
			ResponseModiferChain.Add(c.ResponseModifier) // wire in our ResponseModifier
		}
	})
}

// CORS is an abstraction to handle CORS header nonsense.
// In order to keep origin comparisons as fast as possible, the expressions are pre-compiled,
// and thus need to either be added via AddOrigins() or supplied to NewCORSFromConfig().
type CORS struct {
	AllowCredentials string
	AllowMethods     string
	AllowHeaders     string
	MaxAge           string

	// originRes is a list of compiled Regexps, because faster
	originRes []*pcre.Regexp
}

// NewCORS returns an initialized CORS struct.
func NewCORS() *CORS {
	c := CORS{
		originRes: make([]*pcre.Regexp, 0),
	}
	return &c
}

// NewCORSFromConfig returns an initialized CORS struct from a list of origins and a config map
func NewCORSFromConfig(origins []string, conf map[string]string) (*CORS, error) {
	c := NewCORS()
	c.AllowCredentials = conf["allowcredentials"]
	c.AllowMethods = conf["allowmethods"]
	c.AllowHeaders = conf["allowheaders"]
	c.MaxAge = conf["maxage"]

	if err := c.AddOrigin(origins); err != nil {
		return nil, err
	}
	return c, nil
}

// AddOrigin adds an origin expression to the CORS struct
func (c *CORS) AddOrigin(origins []string) error {

	// Compile the CORS.origins expressions, for speed
	for _, o := range origins {
		re, err := pcre.Compile(o, pcre.CASELESS)
		if err != nil {
			return fmt.Errorf("could not compile CORS.origins regexp: '%s': %s", o, err)
		}
		c.originRes = append(c.originRes, &re)
	}

	return nil
}

// Handler is a middleware that validates Origin request headers against
// a whitelist of expressions, and may change the response headers accordingly
func (c *CORS) Handler(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Make a new PluggableResponseWriter if we need to
		DebugOut.Printf("CorsHandler Pluggable ResponseWriter...\n")
		rw, _ := prw.NewPluggableResponseWriterIfNot(w)
		defer rw.Flush()

		var (
			origin    string
			requestID string
			matched   bool
		)

		if origin = r.Header.Get("Origin"); origin != "" {
			// Do it early, because Flush
			if rid := r.Context().Value(requestIDKey); rid != nil {
				requestID = rid.(string)
			}

			DebugOut.Printf("{%s} CORS check for %s\n", requestID, origin)

			matched = c.handle(rw, origin, r.Method, requestID, false, false)
			if !matched {
				DebugOut.Printf("{%s} CORS whitelisting failed for '%s'\n", requestID, origin)
			}
		}

		// Pass on, and we'll handle the response headers again at the end, tyvm
		next.ServeHTTP(rw, r)

		if origin != "" {
			// Do it again, because response mangling. If we didn't match previously, we only want to delete
			// the headers if they've been added
			c.handle(rw, origin, r.Method, requestID, matched, !matched)
		}

	}

	return http.HandlerFunc(fn)
}

func (c *CORS) String() string {
	return fmt.Sprintf("%s: %s\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n",
		CORSAllowCredentials, c.AllowCredentials,
		CORSExposeHeaders, c.AllowHeaders,
		CORSAllowMethods, c.AllowMethods,
		CORSAllowHeaders, c.AllowHeaders,
		CORSMaxAge, c.MaxAge)
}

// ResponseModifier is an oxy/forward opsetter to remove CORS headers from responses
func (c *CORS) ResponseModifier(resp *http.Response) error {
	// Trash any of these if they exist
	delete(resp.Header, CORSAllowOrigin)
	delete(resp.Header, CORSAllowCredentials)
	delete(resp.Header, CORSExposeHeaders)
	delete(resp.Header, CORSAllowMethods)
	delete(resp.Header, CORSAllowHeaders)
	delete(resp.Header, CORSMaxAge)

	return nil
}

func (c *CORS) handle(rw *prw.PluggableResponseWriter, origin, Method, requestID string, matched, deleteOnly bool) bool {

	// Trash any of these if they exist
	rw.Header().Del(CORSAllowOrigin)
	rw.Header().Del(CORSAllowCredentials)
	rw.Header().Del(CORSExposeHeaders)
	rw.Header().Del(CORSAllowMethods)
	rw.Header().Del(CORSAllowHeaders)
	rw.Header().Del(CORSMaxAge)

	if deleteOnly {
		// We already know we're not going to match, so
		// no sense in rerunning all that jazz
		return false
	}

	if !matched {
		// Check the origin
		for _, o := range c.originRes {
			if o.MatcherString(origin, 0).Matches() {
				DebugOut.Printf("{%s} Origin '%s' matched whitelist\n", requestID, origin)
				matched = true
				break
			}
		}
	}

	if matched {
		// Sweeeeet
		rw.Header().Set(CORSAllowOrigin, origin)
		rw.Header().Set(CORSAllowCredentials, c.AllowCredentials)
		rw.Header().Set(CORSExposeHeaders, c.AllowHeaders)

		if Method == http.MethodOptions {
			rw.Header().Set(CORSAllowMethods, c.AllowMethods)
			rw.Header().Set(CORSAllowHeaders, c.AllowHeaders)
			rw.Header().Set(CORSMaxAge, c.MaxAge)
		}
	}
	return matched
}
