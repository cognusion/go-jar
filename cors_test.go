package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	//ErrorOut = log.New(os.Stderr, "[ERROR] ", OutFormat)
	ErrorOut = log.New(ioutil.Discard, "", 0) // Silence error output, explicitly
}

func CorsHandlersInit() {
	Conf = InitConfig()

	// Run the InitFuncs
	InitFuncs.Call()

	// Setup for CorsHandler testing
	Conf.Set(ConfigCORSOrigins, []string{"^https://.*\\.garbagio\\.(com|net|eu)(?::\\d{1,5})?$", "https://test.test.com"})

	cmap := map[string]string{
		"allowheaders":     Conf.GetString(ConfigCORSAllowHeaders),
		"allowmethods":     Conf.GetString(ConfigCORSAllowMethods),
		"allowcredentials": Conf.GetString(ConfigCORSAllowCredentials),
		"maxage":           Conf.GetString(ConfigCORSMaxAge),
	}
	c, err := NewCORSFromConfig(Conf.GetStringSlice(ConfigCORSOrigins), cmap)
	if err != nil {
		panic(err)
	}
	CorsHandler = c.Handler

	// Setup for ResponseHeaders testing
	Conf.Set(ConfigHeaders, []string{"Crazy-Header %%VERSION", "Dumb"})
}

// TestCorsBadRegexp tests the error-handling around bad regexps
func TestCorsBadRegexp(t *testing.T) {

	cmap := map[string]string{
		"allowheaders":     Conf.GetString(ConfigCORSAllowHeaders),
		"allowmethods":     Conf.GetString(ConfigCORSAllowMethods),
		"allowcredentials": Conf.GetString(ConfigCORSAllowCredentials),
		"maxage":           Conf.GetString(ConfigCORSMaxAge),
	}

	Convey("When a CORS is created with an awful regexp, an error is returned, and everything is ok", t, func() {

		cors, err := NewCORSFromConfig([]string{"(\\"}, cmap)

		So(cors, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

// TestCorsBadConfig tests the error-handling around bad config maps
func TestCorsBadConfig(t *testing.T) {

	cmap := map[string]string{}

	Convey("When a CORS is created with an empty map, everything is ok", t, func() {

		cors, err := NewCORSFromConfig(Conf.GetStringSlice(ConfigCORSOrigins), cmap)

		So(cors, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

// TestCorsHandlerNoOrigin tests CorsHandler when there is no Origin set
func TestCorsHandlerNoOrigin(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and no Origin is set in the request, no Access-Control headers are passed", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			for k := range w.Header() {
				So(k, ShouldNotContainSubstring, "Access-Control")
			}
		})

		rr := httptest.NewRecorder()
		handler := CorsHandler(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestCorsHandlerWithGoodOrigin tests CorsHandler when an Origin header is set to a whitelisted origin
func TestCorsHandlerWithGoodOrigin(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Origin", "https://test.test.com")

	Convey("When a request is made, and an Origin is set to a whitelisted domain in the request, Access-Control headers are passed", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				found := false
				for k := range rec.Header() {
					if strings.Contains(k, "Access-Control") {
						found = true
						break
					}
				}
				So(found, ShouldBeTrue)
			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := testHandler(CorsHandler(dumbHandler))

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestCorsHandlerWithGoodOriginFlush tests CorsHandler when an Origin header is set to a whitelisted origin, and a Flush is triggered downstream
func TestCorsHandlerWithGoodOriginFlush(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Origin", "https://test.test.com")

	Convey("When a request is made, and an Origin is set to a whitelisted domain in the request, Access-Control headers are passed", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				found := false
				for k := range rec.Header() {
					if strings.Contains(k, "Access-Control") {
						found = true
						break
					}
				}
				So(found, ShouldBeTrue)
			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		})

		rr := httptest.NewRecorder()

		handler := testHandler(CorsHandler(dumbHandler))

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestCorsHandlerOptionsWithGoodOrigin tests CorsHandler when an Origin header is set to a whitelisted origin, and OPTIONS is the method
func TestCorsHandlerOptionsWithGoodOrigin(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("OPTIONS", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Origin", "https://test.test.com")

	Convey("When a OPTIONS request is made, and an Origin is set to a whitelisted domain in the request, Access-Control headers are passed", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				found := false
				for k := range rec.Header() {
					if strings.Contains(k, "Access-Control") {
						found = true
						break
					}
				}
				So(found, ShouldBeTrue)

			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := testHandler(CorsHandler(dumbHandler))

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

// TestCorsHandlerWithBadOrigin tests CorsHandler when an Origin header is set to an unlisted origin
func TestCorsHandlerWithBadOrigin(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Origin", "https://garbage.test.com")
	Convey("When a request is made, and an Origin is set to a non-whitelisted domain in the request, no Access-Control headers are passed", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				for k := range w.Header() {
					So(k, ShouldNotContainSubstring, "Access-Control")
				}

			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := testHandler(CorsHandler(dumbHandler))

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestCorsHandlerWhackRegexpOrigin(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Origin", "https://lkfdslsldkfs.garbagio5com")

	Convey("When a request is made, and an Origin is set to a messed up-should be banned domain in the request, Access-Control headers are not passed", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				So(rec.Header(), ShouldBeEmpty)

			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := testHandler(CorsHandler(dumbHandler))

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestCorsHandlerHttpOrigin(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Origin", "http://lkfdslsldkfs.garbagio.com")

	Convey("When a request is made, and an Origin is set to an http URL (as opposed to https), Access-Control headers are not passed", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				So(rec.Header(), ShouldBeEmpty)

			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := testHandler(CorsHandler(dumbHandler))

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestCorsHandlerGoodRegexpOrigin(t *testing.T) {

	CorsHandlersInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Origin", "https://lkfdslsldkfs.garbagio.com")

	Convey("When a request is made, and an Origin is set to a good domain in the request, Access-Control headers are passed", t, func() {
		testHandler := func(next http.Handler) http.Handler {

			fn := func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)

				So(rec.Header(), ShouldNotBeEmpty)

				found := false
				for k := range rec.Header() {
					if strings.Contains(k, "Access-Control") {
						found = true
						break
					}
				}
				So(found, ShouldBeTrue)

			}
			return http.HandlerFunc(fn)
		}

		dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		})

		rr := httptest.NewRecorder()

		handler := testHandler(CorsHandler(dumbHandler))

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}
