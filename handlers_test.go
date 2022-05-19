package jar

import (
	"github.com/cognusion/go-jar/utils"

	"github.com/justinas/alice"
	. "github.com/smartystreets/goconvey/convey"

	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	//ErrorOut = log.New(os.Stderr, "[ERROR] ", OutFormat)
	ErrorOut = log.New(ioutil.Discard, "", 0) // Silence error output, explicitly
}

func HandlersInit() {
	Conf = InitConfig()

	// Run the InitFuncs
	InitFuncs.Call()

	// Setup for ResponseHeaders testing
	Conf.Set(ConfigHeaders, []string{"Crazy-Header %%VERSION", "Dumb"})
}

func TestHandleHandlerLowerCase(t *testing.T) {

	Convey("When a request for a known-handler is made, and the name is lower-cased, it is found", t, func() {
		chain, err := HandleHandler("recoverer", alice.Chain{})
		So(err, ShouldBeNil)
		So(chain, ShouldNotBeNil)
	})
}

func TestHandleHandlerMixedCase(t *testing.T) {
	Convey("When a request for a known-handler is made, and the name is mix-cased, it is found", t, func() {
		chain, err := HandleHandler("RecoVereR", alice.Chain{})
		So(err, ShouldBeNil)
		So(chain, ShouldNotBeNil)
	})
}

// TestRecovererPanic tests Recoverer when there is a panic
func TestRecovererPanic(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and panics, it is recovered", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("OH MY GOSH!")
		})

		rr := httptest.NewRecorder()

		handler := Recoverer(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusInternalServerError)
		So(rr.Body.String(), ShouldContainSubstring, http.StatusText(http.StatusInternalServerError))
	})
}

// TestRecovererPanicErrAbortHandler tests Recoverer when there is a panic, that throws ErrAbortHandler
func TestRecovererPanicErrAbortHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and panics with http.ErrAbortHandler, it is recovered", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(http.ErrAbortHandler)
		})

		rr := httptest.NewRecorder()

		handler := Recoverer(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusInternalServerError)
		So(rr.Body.String(), ShouldContainSubstring, ErrAborted.Error())
	})
}

// TestRecovererPanic tests Recoverer when there is a panic
func TestRecovererPanicWithStack(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and panics, it is recovered", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("OH MY GOSH!")
		})

		Conf.Set(ConfigRecovererLogStackTraces, true)

		rr := httptest.NewRecorder()

		handler := Recoverer(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusInternalServerError)
		So(rr.Body.String(), ShouldContainSubstring, http.StatusText(http.StatusInternalServerError))
	})
}

// TestRecovererNoPanic tests Recoverer when there is no panic
func TestRecovererNoPanic(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and it doesn't panic, it is left alone", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := Recoverer(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestRealAddrNoXFF tests RealAdd when there is no X-Forwarded-For set
func TestRealAddrNoXFF(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and there is no X-Forwarded-For header set, RemoteAddr isn't set", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.RemoteAddr, ShouldBeEmpty)
		})

		rr := httptest.NewRecorder()

		handler := RealAddr(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestRealAddrXFF tests RealAdd when there is a simple X-Forwarded-For set
func TestRealAddrXFF(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-FORWARDED-FOR", "zip.zap.zot")

	Convey("When a request is made, and there is a X-Forwarded-For header set, RemoteAddr is set properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.RemoteAddr, ShouldEqual, "zip.zap.zot")
		})

		rr := httptest.NewRecorder()

		handler := RealAddr(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestRealAddXFFChain tests RealAdd when there is an X-Forwarded-For set with a chain
func TestRealAddrXFFChain(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-FORWARDED-FOR", "yippy.skippy,hippy.dippy.do,zip.zap.zot")

	Convey("When a request is made, and there is a multi-item X-Forwarded-For header set, RemoteAddr is set properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.RemoteAddr, ShouldEqual, "zip.zap.zot")
		})

		rr := httptest.NewRecorder()

		handler := RealAddr(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestPathReplacer tests PathReplacer.Handler
func TestPathReplacer(t *testing.T) {
	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and a replacement is requested, it is done correctly", t, func() {
		p := PathReplacer{
			From: "/stuff",
			To:   "/other",
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.RequestURI, ShouldBeEmpty) // not always set
			So(r.URL.Path, ShouldEqual, "/other")
		})

		rr := httptest.NewRecorder()

		handler := p.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestRateLimiterClear tests RateLimiter.Handler slow enough to not trip
func TestRateLimiterClear(t *testing.T) {
	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.RemoteAddr = "zip.zap.zot"
	rl := NewRateLimiter(1, 0)

	Convey("When a request is made, and there is a ratelimiter, if the request rate is under the limit, the request is ok", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := rl.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestRateLimiterBroke tests RateLimiter.Handler too fast
func TestRateLimiterBroke(t *testing.T) {
	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.RemoteAddr = "zip.zap.zot"
	rl := NewRateLimiter(1, time.Minute)

	Convey("When many requests are made, and there is a ratelimiter, the appropriate number of requests are allowed and denied", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		handler := rl.Handler(testHandler)

		var (
			fourtwentynines = 0
			twohundreds     = 0
		)

		for i := 0; i < 100; i++ {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code == 429 {
				fourtwentynines++
			} else if rr.Code == 200 {
				twohundreds++
			} else {
				t.Errorf("handler returned wrong status code: got %v want 200 or 429\n",
					rr.Code)
			}
		}
		// Check the status code is what we expect.
		So(twohundreds, ShouldBeLessThanOrEqualTo, 1)
	})
}

func TestRateLimiterFractionClear(t *testing.T) {

	t.Skip("Slow")

	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.RemoteAddr = "zip.zap.zot"
	rl := NewRateLimiter(0.5, 0)

	Convey("When a request is made, and there is a ratelimiter set to a fraction, if the request rate is under the limit, the request is ok", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		var (
			fourtwentynines = 0
			twohundreds     = 0
		)

		//Println()
		for i := 0; i < 10; i++ {
			rr := httptest.NewRecorder()

			handler := rl.Handler(testHandler)

			handler.ServeHTTP(rr, req)

			if rr.Code == 429 {
				fourtwentynines++
			} else if rr.Code == 200 {
				twohundreds++
			} else {
				t.Errorf("handler returned wrong status code: got %v want 200 or 429\n",
					rr.Code)
			}

			//Printf("%+v\n", rr.Header())

			time.Sleep(2 * time.Second)
		}
		// Check the status code is what we expect.
		So(fourtwentynines, ShouldBeZeroValue)

	})
}

// TestResponseHeadersHandler tests ResponseHeaders to ensure they are set or deleted
func TestResponseHeadersHandler(t *testing.T) {

	HandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and header adds/removes are requested, the response headers look correct", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				So(rec.Header().Get("Crazy-Header"), ShouldEqual, VERSION)
				So(rec.Header().Get("Dumb"), ShouldBeEmpty)
			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Dumb", "dumdum")
		})

		rr := httptest.NewRecorder()

		handler := testHandler(ResponseHeaders(dumbHandler))

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestResponseHeadersHandlerFlush tests ResponseHeaders to ensure they are set or deleted properly if Flush is called downstream
func TestResponseHeadersHandlerFlush(t *testing.T) {

	HandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and header adds/removes are requested, the response headers look correct", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				So(rec.Header().Get("Crazy-Header"), ShouldEqual, VERSION)
				So(rec.Header().Get("Dumb"), ShouldBeEmpty)
			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			w.Header().Set("Dumb", "dumdum")
		})

		rr := httptest.NewRecorder()

		handler := testHandler(ResponseHeaders(dumbHandler))

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestAuthoritativeDomainsHandlerDefault tests AuthoritativeDomainsHandler when it is unconfigured (off)
func TestAuthoritativeDomainsHandlerDefault(t *testing.T) {
	req, err := http.NewRequest("GET", "/notbanned", nil)
	if err != nil {
		t.Fatal(err)
	}

	// authoritativeDomains = []string{"whatevs"}
	CheckAuthoritative = func(*http.Request) bool { return true } // checkAuthoritative

	Convey("When a request is made and no authoritative domains are configured, it succeeds", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := AuthoritativeDomainsHandler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestAuthoritativeDomainsHandlerNot tests AuthoritativeDomainsHandler when it is configured and the request is not to one
func TestAuthoritativeDomainsHandlerNot(t *testing.T) {
	req, err := http.NewRequest("GET", "/notbanned", nil)
	if err != nil {
		t.Fatal(err)
	}

	authoritativeDomains = []string{"whatevs"}
	CheckAuthoritative = checkAuthoritative

	Convey("When a request is made and authoritative domains are configured, a request elsewhere fails", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := AuthoritativeDomainsHandler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusBadRequest)
	})
}

// TestAuthoritativeDomainsHandlerOk tests AuthoritativeDomainsHandler when it is configured and the request is to one
func TestAuthoritativeDomainsHandlerOk(t *testing.T) {
	req, err := http.NewRequest("GET", "http://whatevs/notbanned", nil)
	if err != nil {
		t.Fatal(err)
	}

	authoritativeDomains = []string{"whatevs"}
	CheckAuthoritative = checkAuthoritative

	Convey("When a request is made and authoritative domains are configured, a request to a listed domain succeeds", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := AuthoritativeDomainsHandler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestAuthoritativeDomainsHandlerOkCaseJacked(t *testing.T) {
	req, err := http.NewRequest("GET", "http://whaTEvs/notbanned", nil)
	if err != nil {
		t.Fatal(err)
	}

	authoritativeDomains = []string{"whatevs"}
	CheckAuthoritative = checkAuthoritative

	Convey("When a request is made with mixedcase, and authoritative domains are configured, a request to a listed domain succeeds", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := AuthoritativeDomainsHandler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestForbiddenPathsHandlerNot tests ForbiddenPathsHandler when the path is not forbidden
func TestForbiddenPathsHandlerNot(t *testing.T) {
	req, err := http.NewRequest("GET", "/notbanned", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, to a not-banned path, it succeeds", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		fp, err := NewForbiddenPaths([]string{"/banned"})
		So(err, ShouldBeNil)

		handler := fp.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestForbiddenPathsHandlerForbidden tests ForbiddenPathsHandler when the path is forbidden
func TestForbiddenPathsHandlerForbidden(t *testing.T) {
	req, err := http.NewRequest("GET", "/banned", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, to a banned path, it fails", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Next handler fired! Fail!\n")
		})

		rr := httptest.NewRecorder()

		fp, err := NewForbiddenPaths([]string{"/banned"})
		if err != nil {
			t.Fatal(err)
		}

		handler := fp.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)
	})
}

func TestUrlCapture(t *testing.T) {

	uri := "http://somewheresomwhere.com/some/path/somewhere"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to a specific URL, the requested uri is put into the context correctly", t, func() {
		rr := httptest.NewRecorder()

		afterHandler := func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				So(r.Header.Get("X-URLPOOLMEMBER"), ShouldBeEmpty)

				next.ServeHTTP(w, r)

				So(r.Header.Get("X-URLPOOLMEMBER"), ShouldEqual, "somewheresomwhere.com")

			}
			return http.HandlerFunc(fn)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		handler := afterHandler(URLCaptureHandler(testHandler))

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

	})
}

func TestTimeoutHandlerMessage(t *testing.T) {

	uri := "http://somewheresomwhere.com/some/path/somewhere"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to a specific URL, the requested uri is put into the context correctly", t, func() {
		rr := httptest.NewRecorder()

		to := Timeout{
			Duration: 100 * time.Millisecond,
			Message:  "too long",
		}

		delayHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Minute)

				next.ServeHTTP(w, r)

			}
			return http.HandlerFunc(fn)
		}
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Next handler fired! Fail!\n")
		})

		handler := to.Handler(delayHandler(testHandler))

		start := time.Now()
		handler.ServeHTTP(rr, req)
		total := time.Since(start)

		So(rr.Code, ShouldEqual, http.StatusServiceUnavailable)
		So(total, ShouldBeLessThan, 1*time.Second)
		So(rr.Body.String(), ShouldEqual, "too long")

	})
}

func TestTimeoutHandlerNoMessage(t *testing.T) {

	uri := "http://somewheresomwhere.com/some/path/somewhere"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to a specific URL, the requested uri is put into the context correctly", t, func() {
		rr := httptest.NewRecorder()

		to := Timeout{
			Duration: 100 * time.Millisecond,
		}

		delayHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Minute)

				next.ServeHTTP(w, r)

			}
			return http.HandlerFunc(fn)
		}
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Next handler fired! Fail!\n")
		})

		handler := to.Handler(delayHandler(testHandler))

		start := time.Now()
		handler.ServeHTTP(rr, req)
		total := time.Since(start)

		So(rr.Code, ShouldEqual, http.StatusServiceUnavailable)
		So(total, ShouldBeLessThan, 10*time.Second)
		So(rr.Body.String(), ShouldEqual, "Timeout of 100ms exceeded\n")

	})
}

func TestBodyByteLimit(t *testing.T) {
	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	rd := strings.NewReader(randString(5 + 1))
	rc := ioutil.NopCloser(rd)
	req.Body = rc

	bl := NewBodyByteLimit(5)

	Convey("When a request is made, and there a body byte limit, and the body is too large, the handler returns properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not be reached")
		})

		rr := httptest.NewRecorder()

		handler := bl.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusRequestEntityTooLarge)
	})
}

func randString(length int) string {
	charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	outString := make([]byte, length)
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		outString[i] = randomChar
	}
	return string(outString)
}

func TestBodyByteLimitMega(t *testing.T) {
	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	rd := strings.NewReader(randString(1024 * 1024))
	rc := ioutil.NopCloser(rd)
	req.Body = rc

	bl := NewBodyByteLimit(5)

	Convey("When a request is made, and there a body byte limit, and the body is way too large, the handler returns properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not be reached")
		})

		rr := httptest.NewRecorder()

		handler := bl.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusRequestEntityTooLarge)
	})
}

func TestBodyByteLimitOk(t *testing.T) {
	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	rdString := randString(5)
	rd := strings.NewReader(rdString)
	rc := ioutil.NopCloser(rd)
	req.Body = rc

	bl := NewBodyByteLimit(5)

	Convey("When a request is made, and there a body byte limit, and the body is exactly max size, the request passes through fine", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := utils.ReadAll(r.Body)
			defer r.Body.Close()
			So(err, ShouldBeNil)
			So(string(b), ShouldEqual, rdString)
		})

		rr := httptest.NewRecorder()

		handler := bl.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestBodyByteLimitNoBody(t *testing.T) {
	req, err := http.NewRequest("GET", "/stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	bl := NewBodyByteLimit(5)

	Convey("When a request is made, and there a body byte limit, the handler does not panic and request passes through fine", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(true, ShouldBeTrue)
		})

		rr := httptest.NewRecorder()

		handler := bl.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestPoolID(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	poolid := PoolID{"APOOL"}

	Convey("When a request is made, and the PoolID handler is used, the header  is set properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.Header.Get("X-URLPOOL"), ShouldEqual, "APOOL")
		})

		rr := httptest.NewRecorder()

		handler := poolid.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestRouteIDSetup(t *testing.T) {
	req, err := http.NewRequest("GET", "/?ROUTEID=sdfsd8798s9df8ds", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, with the ROUTEID on the param list, it succeeds", t, func() {
		rr := httptest.NewRecorder()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		handler := SetupHandler(RouteIDInspectionHandler(testHandler))

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

		Convey("and the proper cookie now exists", func() {
			c, err := req.Cookie("ROUTEID")
			So(err, ShouldBeNil)
			So(c.Value, ShouldEqual, "sdfsd8798s9df8ds")
		})
	})
}

func TestRouteIDDupe(t *testing.T) {
	req, err := http.NewRequest("GET", "/?ROUTEID=sdfsd8798s9df8ds", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "ROUTEID",
		Value: "blueblueblue",
		Path:  "/",
	})

	Convey("When a request is made, with the ROUTEID on the param list and a different ROUTEID in a cookie, the parse succeeds", t, func() {
		rr := httptest.NewRecorder()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		handler := SetupHandler(RouteIDInspectionHandler(testHandler))

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

		Convey("but the cookie value is NOT the one from the param list", func() {
			c, err := req.Cookie("ROUTEID")
			So(err, ShouldBeNil)
			So(c.Value, ShouldNotEqual, "sdfsd8798s9df8ds")
		})
	})
}
