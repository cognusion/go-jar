package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	//TimingOut = log.New(os.Stderr, "[TIMING] ", 0)
}

func setupHMAC() *HMAC {
	return NewHMAC(string(randBytes(64)), "", 0)
}

func Test_HMACSigner(t *testing.T) {

	hmac := setupHMAC()
	hmacExp := setupHMAC()
	hmacExp.Expiration, _ = time.ParseDuration("5ms")

	Convey("When a signing request is made to HMAC, it is signed", t, func() {
		uri := "/stuff"
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		hmac.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusTemporaryRedirect)
		loc := rr.Header().Get("Location")
		So(loc, ShouldStartWith, uri)

		Convey("... and when it is requested, the request passed through", func() {
			req, err := http.NewRequest("GET", loc, nil)
			if err != nil {
				t.Fatal(err)
			}

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				So(r.RequestURI, ShouldBeEmpty) // not always set
				So(r.URL.Path, ShouldEqual, uri)
			})

			rr := httptest.NewRecorder()

			handler := hmac.Handler(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
		})
	})

	Convey("When a signing request is made to HMAC with an expiration, it is signed", t, func() {
		uri := "/stuff"
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		hmacExp.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusTemporaryRedirect)
		loc := rr.Header().Get("Location")
		So(loc, ShouldStartWith, uri)

		Convey("... and when it is requested, the request passed through", func() {
			req, err := http.NewRequest("GET", loc, nil)
			if err != nil {
				t.Fatal(err)
			}

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				So(r.RequestURI, ShouldBeEmpty) // not always set
				So(r.URL.Path, ShouldEqual, uri)
			})

			rr := httptest.NewRecorder()

			handler := hmacExp.Handler(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_HMACVerifier(t *testing.T) {

	hmac := setupHMAC()
	hmacExp := setupHMAC()
	hmacExp.Expiration, _ = time.ParseDuration("5ms")

	Convey("When an unsigned request is made to an HMAC-secured path it is Forbidden", t, func() {
		req, err := http.NewRequest("GET", "/stuff", nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fail()
		})

		rr := httptest.NewRecorder()

		handler := hmac.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("When an incorrectly-signed request is made to an HMAC-secured path it is Forbidden", t, func() {
		req, err := http.NewRequest("GET", "/stuff/K1Cumlhd6nZM6QJsqr4IACs2", nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fail()
		})

		rr := httptest.NewRecorder()

		handler := hmac.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("When a correctly-signed request is made to an HMAC-secured path, it is passed through", t, func() {
		uri := "/stuff"
		hash := signHMAC([]byte(uri), hmac.key, hmac.salt)
		signedURI := craftURI(fmt.Sprintf("%s/%s", uri, hash), make(url.Values))

		req, err := http.NewRequest("GET", signedURI, nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.RequestURI, ShouldBeEmpty) // not always set
			So(r.URL.Path, ShouldEqual, uri)
		})

		rr := httptest.NewRecorder()

		handler := hmac.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

	})

	Convey("When a correctly-signed request with query parameters is made to an HMAC-secured path, it is passed through", t, func() {
		uri := "/stuff"
		qp := "?p1=hello&p2=world"
		hash := signHMAC([]byte(uri+qp), hmac.key, hmac.salt)
		signedURI := craftURI(fmt.Sprintf("%s/%s%s", uri, hash, qp), make(url.Values))

		req, err := http.NewRequest("GET", signedURI, nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.RequestURI, ShouldBeEmpty) // not always set
			So(r.URL.Path, ShouldEqual, uri)
		})

		rr := httptest.NewRecorder()

		handler := hmac.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

	})

	Convey("When a correctly-signed request with query parameters that have been altered is made to an HMAC-secured path, it is Forbidden", t, func() {
		uri := "/stuff"
		qp := "?p1=hello&p2=world"
		qp2 := "?p1=gobye&p2=world"
		hash := signHMAC([]byte(uri+qp), hmac.key, hmac.salt)
		signedURI := craftURI(fmt.Sprintf("%s/%s%s", uri, hash, qp2), make(url.Values))

		req, err := http.NewRequest("GET", signedURI, nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fail()
		})

		rr := httptest.NewRecorder()

		handler := hmac.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)

	})

	Convey("When a correctly-signed request with query parameters and an expiration is made to an HMAC-secured path, it is passed through", t, func() {
		uri := "/stuff"
		qp := fmt.Sprintf("?expiration=%d&p1=hello&p2=world", time.Now().Add(5*time.Millisecond).UnixMilli())
		hash := signHMAC([]byte(uri+qp), hmacExp.key, hmacExp.salt)
		signedURI := craftURI(fmt.Sprintf("%s/%s%s", uri, hash, qp), make(url.Values))

		req, err := http.NewRequest("GET", signedURI, nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.RequestURI, ShouldBeEmpty) // not always set
			So(r.URL.Path, ShouldEqual, uri)
		})

		rr := httptest.NewRecorder()

		handler := hmacExp.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

	})

	Convey("When a correctly-signed request with query parameters and an *expired* expiration is made to an HMAC-secured path, it is Forbidden", t, func() {
		uri := "/stuff"
		qp := fmt.Sprintf("?expiration=%d&p1=hello&p2=world", time.Now().Add(-5*time.Millisecond).UnixMilli())
		hash := signHMAC([]byte(uri+qp), hmacExp.key, hmacExp.salt)
		signedURI := craftURI(fmt.Sprintf("%s/%s%s", uri, hash, qp), make(url.Values))

		req, err := http.NewRequest("GET", signedURI, nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fail()
		})

		rr := httptest.NewRecorder()

		handler := hmacExp.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)

	})

	Convey("When a correctly-signed request with query parameters and a garbage expiration is made to an HMAC-secured path, it is Forbidden", t, func() {
		uri := "/stuff"
		qp := fmt.Sprintf("?expiration=%s&p1=hello&p2=world", "garbage")
		hash := signHMAC([]byte(uri+qp), hmacExp.key, hmacExp.salt)
		signedURI := craftURI(fmt.Sprintf("%s/%s%s", uri, hash, qp), make(url.Values))

		req, err := http.NewRequest("GET", signedURI, nil)
		if err != nil {
			t.Fatal(err)
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fail()
		})

		rr := httptest.NewRecorder()

		handler := hmacExp.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)

	})
}
