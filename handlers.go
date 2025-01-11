package jar

import (
	"github.com/cognusion/go-prw"
	"github.com/cognusion/go-timings"
	"github.com/davecgh/go-spew/spew"
	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	gerrors "github.com/go-errors/errors"
	"github.com/justinas/alice"

	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	requestIDKey commonIDKey = iota
	sessionIDKey
	timestampKey
	pathIDKey
	poolIDKey
	// PathOptionsKey is a keyid for setting/getting PathOptions to/from a Context
	PathOptionsKey

	// ErrAborted is only used during panic recovery, if http.ErrAbortHandler was called
	ErrAborted = Error("client aborted connection, or connection closed")
)

// Constants for configuration key strings
const (
	ConfigHeaders                 = ConfigKey("headers")
	ConfigRecovererLogStackTraces = ConfigKey("Recoverer.logstacktraces")
)

type commonIDKey int

var (
	// Handlers is a map of available Handlers (middlewares)
	Handlers = make(HandlerMap)
)

func init() {
	// Set up the static handlers
	Handlers["recoverer"] = Recoverer
	Handlers["routeidinspectionhandler"] = RouteIDInspectionHandler

	// Set up the configs
	ConfigAdditions[ConfigRecovererLogStackTraces] = false // logstacktraces controls whether the Recoverer handler logs stacktraces, or just the error
}

// HandlerMap maps handler names to their funcs
type HandlerMap map[string]func(http.Handler) http.Handler

// List returns the names of all of the Handlers
func (h *HandlerMap) List() []string {
	l := make([]string, len(*h))
	i := 0
	for k := range *h {
		l[i] = k
		i++
	}
	return l
}

// HandleHandler takes a handler name, and an existing chain, and returns a new chain or an error
func HandleHandler(handler string, hchain alice.Chain) (alice.Chain, error) {

	if h, ok := Handlers[strings.ToLower(handler)]; ok {
		hchain = hchain.Append(h)
	} else {
		return alice.Chain{}, ErrConfigurationError{fmt.Sprintf("handler '%s' is not a listed handler", handler)}
	}

	return hchain, nil
}

// AuthoritativeDomainsHandler declines to handle requests that are not listed in "authoritativedomains" config
func AuthoritativeDomainsHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		if autho := CheckAuthoritative(r); !autho {
			ErrorOut.Print(ErrRequestError{r, fmt.Sprintf("Declining HTTP traffic for '%s' because not a valid authoritative domain\n", r.Host)})
			http.Error(w, ErrRequestError{r, "Not authoritative"}.Error(), 400)
			RequestErrorResponse(r, w, "Not authoritative for this domain", http.StatusBadRequest)
			return
		}
		TimingOut.Printf("AuthoritativeDomainsHandler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// URLCaptureHandler is an unchainable handler that captures the Hostname of the Pool Member servicing a request
func URLCaptureHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		r.Header.Set(Conf.GetString(ConfigPoolMemberHeaderName), r.URL.Hostname())
		DebugOut.Println(RequestErrorString(r, fmt.Sprintf("Request: %+v URL %+v", r, *r.URL)))
		TimingOut.Printf("URLCaptureHandler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// Timeout is a middleware that causes a 503 Service Unavailable message to be handed back if the timeout trips
type Timeout struct {
	Duration time.Duration
	Message  string
}

// Handler is the handler for Timeout
func (t *Timeout) Handler(next http.Handler) http.Handler {
	if t.Message == "" {
		t.Message = fmt.Sprintf("Timeout of %s exceeded\n", t.Duration.String())
	}
	return http.TimeoutHandler(next, t.Duration, t.Message)
}

// RealAddr is a special handler to grab the most probable "real" client address
func RealAddr(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			rad := strings.TrimSpace(ips[len(ips)-1]) // the last item in an XFF list is probably what we want
			//DebugOut.Printf("RealAddr: %s\n", rad)
			r.RemoteAddr = rad
		}
		TimingOut.Printf("RealAddr handler took %s\n", t.Since().String())
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// RouteIDInspectionHandler checks the Query params for a ROUTEID and shoves it into a cookie
func RouteIDInspectionHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		// Check to see if ROUTEID is specified on the request line, and if so, push it into the request cookie
		if strings.Contains(r.URL.RawQuery, "ROUTEID") {
			if rkey := r.URL.Query().Get("ROUTEID"); rkey != "" {
				r.AddCookie(&http.Cookie{
					Name:  "ROUTEID",
					Value: rkey,
					Path:  "/",
				})
			}
		}
		TimingOut.Printf("RouteIDInspectionHandler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// SetupHandler adds the RequestID and various other informatives to a request context
func SetupHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		// Hit the request counter
		Counter()

		// Track connections, roughly
		ConnectionCounterAdd()
		defer ConnectionCounterRemove()

		var requestID string

		if Conf.GetBool(ConfigTrustRequestIDHeader) {
			requestID = r.Header.Get(Conf.GetString(ConfigRequestIDHeaderName))
		}

		if requestID == "" {
			// Set the requestID ourselves
			requestID = Seq.NextHashID()
		}

		if Conf.GetBool(ConfigDebug) && Conf.GetBool(ConfigDebugRequests) {
			// dump the request, yo
			DebugOut.Printf("Request {%s}:\n%s\n/Request {%s}\n", requestID, spew.Sdump(*r), requestID)
		}

		DebugOut.Printf("{%s} Proto %s\n", requestID, r.Proto)

		if Conf.GetBool(ConfigDebug) && r.TLS != nil {
			DebugOut.Printf("{%s} TLS %s, CipherSuite %s\n", requestID, SslVersions.Suite(r.TLS.Version), Ciphers.Suite(r.TLS.CipherSuite))
		}

		r = r.WithContext(WithRqID(context.WithValue(r.Context(), timestampKey, time.Now()), requestID))

		r.Header.Set(Conf.GetString(ConfigRequestIDHeaderName), requestID)

		var pathID string
		if pid := r.Context().Value(pathIDKey); pid != nil {
			pathID = pid.(string)
		}
		DebugOut.Printf("{%s} Executing Path: %s\n", requestID, pathID)

		TimingOut.Printf("Setup handler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
		// CRITICAL: SetupHandler must never ever change the response. Do not write below this line.
		DebugOut.Printf("{%s} Done with Path: %s\n", requestID, pathID)
	}

	return http.HandlerFunc(fn)
}

// Recoverer is a wrapping handler to make panic-capable handlers safer
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if rec := recover(); rec != nil {
				var err error

				switch t := rec.(type) {
				case string:
					err = Error(t)
				case error:
					err = t
				default:
					err = ErrUnknownError
				}
				if errors.Is(err, http.ErrAbortHandler) {
					// ErrAbortHandler is called when the client closes the connection or the connection is closed
					// so we don't need to lose our poop, just clean it up and move on
					ErrorOut.Printf("%s\n", ErrRequestError{r, ErrAborted.Error()})
					DebugOut.Printf("ErrAbortHandler: %s\n", ErrRequestError{r, fmt.Sprintf("Panic occurred: %s", gerrors.Wrap(err, 2).ErrorStack())})
					http.Error(w, ErrRequestError{r, StatusClientClosedRequestText}.Error(), StatusClientClosedRequest) // Machine-readable
					return
				} else if Conf.GetBool(ConfigRecovererLogStackTraces) {
					ErrorOut.Printf("%s\n", ErrRequestError{r, fmt.Sprintf("Panic occurred: %s", gerrors.Wrap(err, 2).ErrorStack())})
				} else {
					ErrorOut.Printf("%s\n", ErrRequestError{r, fmt.Sprintf("Panic occurred: %s", err)})
				}
				//http.Error(w, ErrRequestError{r, "an internal error occurred"}.Error(), http.StatusInternalServerError)
				RequestErrorResponse(r, w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimiter is a wrapper around limiter.Limiter
type RateLimiter struct {
	*limiter.Limiter
	collectOnly bool
}

// NewRateLimiter returns a RateLimiter based on the specified max rps and purgeDuration
func NewRateLimiter(max float64, purgeDuration time.Duration) RateLimiter {
	rl := RateLimiter{}

	if purgeDuration == 0 {
		rl = RateLimiter{
			tollbooth.NewLimiter(max, nil),
			false,
		}
	} else {
		rl = RateLimiter{
			tollbooth.NewLimiter(max, &limiter.ExpirableOptions{DefaultExpirationTTL: purgeDuration}),
			false,
		}
	}

	rl.Limiter.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})

	return rl
}

// NewRateLimiterCollector returns a RateLimiter based on the specified max rps and purgeDuration
func NewRateLimiterCollector(max float64, purgeDuration time.Duration) RateLimiter {
	rl := NewRateLimiter(max, purgeDuration)
	rl.collectOnly = true

	return rl
}

// Handler is the middleware for the RateLimiter
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Make a new PluggableResponseWriter if we need to
		DebugOut.Printf("RateLimiter Pluggable ResponseWriter...\n")
		rw, _ := prw.NewPluggableResponseWriterIfNot(w)
		defer rw.Flush()

		httpError := tollbooth.LimitByRequest(rl.Limiter, rw, r)
		if httpError != nil {
			DebugOut.Printf("%s: %s\n", ErrRequestError{r, "Request tripped limiter"}.String(), r.RequestURI)
			if !rl.collectOnly {
				rl.ExecOnLimitReached(rw, r)
				rw.Header().Add("Content-Type", rl.GetMessageContentType())
				sec := 1
				max := rl.GetMax()
				if max < 1.0 && max > 0.0 {
					sec = int(math.Ceil(1.0 / max))
				}
				rw.Header().Add("Retry-After", fmt.Sprintf("%d", sec))
				rw.WriteHeader(httpError.StatusCode)
				rw.Write([]byte(httpError.Message))
				return
			}
		}

		next.ServeHTTP(rw, r)
	}
	return http.HandlerFunc(fn)
}

// ResponseHeaders is a simple piece of middleware that sets configured headers
func ResponseHeaders(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		// Make a new PluggableResponseWriter if we need to
		DebugOut.Printf("ResponseHeaders Pluggable ResponseWriter...\n")
		rw, _ := prw.NewPluggableResponseWriterIfNot(w)
		defer rw.Flush()

		// NOTE: we do this early in the event a Flush() is called during
		// the pre-response-end phase
		rmHeaders := make([]string, 0)
		addHeaders := make(map[string]string)
		for _, header := range Conf.GetStringSlice(ConfigHeaders) {
			if strings.Contains(header, " ") {
				// Set it
				hparts := strings.SplitN(header, " ", 2)
				hvalue := hparts[1]
				if strings.Contains(hvalue, "%%") {
					hvalue = MacroDictionary.Replacer(hvalue)
				}
				addHeaders[hparts[0]] = hvalue
			} else {
				// Queue it for removal
				rmHeaders = append(rmHeaders, header)
			}
		}

		// Pass along the headers to remove or add to PRW (flush-safe)
		rw.SetHeadersToRemove(rmHeaders)
		rw.SetHeadersToAdd(addHeaders)

		TimingOut.Printf("ResponseHeaders handler took %s\n", t.Since().String())
		next.ServeHTTP(rw, r)
	}
	return http.HandlerFunc(fn)
}

// ForbiddenPaths is a struct to assist in the expedient resolution of determining if a Request is destined to a forbidden path
type ForbiddenPaths struct {
	// Paths is a list of compiled Regexps, because speed
	Paths []*regexp.Regexp
}

// NewForbiddenPaths takes a list of regexp-compatible strings, and returns the analogous ForbiddenPaths with compiled regexps,
// or an error if a regexp could not be compiled
func NewForbiddenPaths(paths []string) (*ForbiddenPaths, error) {

	fp := ForbiddenPaths{
		Paths: make([]*regexp.Regexp, len(paths)),
	}

	for i, f := range paths {
		re, err := regexp.Compile("(?i)" + f) // case-insensitive
		if err != nil {
			return nil, ErrConfigurationError{fmt.Sprintf("could not compile forbidden path %d regexp: '%s': %s\n", i, f, err)}
		}
		fp.Paths[i] = re
	}

	return &fp, nil
}

func (f *ForbiddenPaths) match(path string) (int, bool) {
	for i, fp := range f.Paths {
		if fp.MatchString(path) {
			return i, true
		}
	}
	return -1, false
}

func (f *ForbiddenPaths) remove(index int) {
	if len(f.Paths) == 0 || index+1 > len(f.Paths) {
		// WTF?!
		return
	} else if index == 0 {
		// Base
		f.Paths = make([]*regexp.Regexp, 0)
		return
	}
	f.Paths = append(f.Paths[:index], f.Paths[index+1:]...)
}

func (f *ForbiddenPaths) strings() []string {
	slist := make([]string, len(f.Paths))
	for i, p := range f.Paths {
		slist[i] = p.String()
	}
	return slist
}

// Handler is a middleware that checks the request URI against regexps and 403's if match
func (f *ForbiddenPaths) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		for i, f := range f.Paths {
			if f.MatchString(r.URL.Path) {
				DebugOut.Printf("ForbiddenPath '%s' matched #%d\n", r.URL.Path, i)
				RequestErrorResponse(r, w, ErrForbiddenError.Error(), http.StatusForbidden)
				TimingOut.Printf("ForbiddenPathsHandler handler (matched) took %s\n", t.Since().String())
				return
			}
		}
		TimingOut.Printf("ForbiddenPathsHandler handler (no match) took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// PathStripper is a wrapping struct to remove the prefix from the Request path
type PathStripper struct {
	Prefix string
}

// Handler is a middleware that replaces the Request path
func (p *PathStripper) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		TrimPrefixURI(r, p.Prefix)

		TimingOut.Printf("PathStripper handler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// PathReplacer is a wrapping struct to replace the Request path
type PathReplacer struct {
	From string
	To   string
}

// Handler is a middleware that replaces the Request path
func (p *PathReplacer) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		ReplaceURI(r, strings.Replace(r.URL.Path, p.From, p.To, 1), strings.Replace(r.RequestURI, p.From, p.To, 1))

		TimingOut.Printf("PathReplacer handler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// PathHandler is a wrapping struct to inject the Path name, and any PathOptions into the Context
type PathHandler struct {
	Path    string
	Options PathOptions
}

// Handler is a middleware that injects the Path name, and any PathOptions into the Context
func (p *PathHandler) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		r = r.WithContext(context.WithValue(r.Context(), pathIDKey, p.Path))
		if len(p.Options) > 0 {
			r = r.WithContext(context.WithValue(r.Context(), PathOptionsKey, p.Options))
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// PoolID is a wrapping struct to inject the Pool name into the Context
type PoolID struct {
	Pool string
}

// Handler injects the Pool name into the Context
func (p *PoolID) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		r = r.WithContext(context.WithValue(r.Context(), poolIDKey, p.Pool))
		r.Header.Set(Conf.GetString(ConfigPoolHeaderName), p.Pool)

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// BodyByteLimit is a Request.Body size limiter
type BodyByteLimit struct {
	byteLimit int64
}

// NewBodyByteLimit returns an initialized BodyByteLimit
func NewBodyByteLimit(limit int64) BodyByteLimit {
	return BodyByteLimit{limit}
}

// Handler limits the size of Request.Body
func (b *BodyByteLimit) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			// Read at most 1 more byte than we allow, replacing
			// the request body with a new body

			buff := RecyclableBufferPool.Get()
			if err := buff.ResetFromLimitedReader(r.Body, b.byteLimit); err != nil {
				r.Body.Close()
				buff.Close()
				DebugOut.Print(ErrRequestError{r, "Request body too large"}.Error())
				http.Error(w, ErrRequestError{r, http.StatusText(http.StatusRequestEntityTooLarge)}.Error(), http.StatusRequestEntityTooLarge) // Machine-readable
				return
			}
			r.Body.Close()
			// Replace the Body with our own
			r.Body = buff
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// WithRqID returns a context which knows its request ID
func WithRqID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithSessionID returns a context which knows its session ID
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}
