package jar

import (
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vulcand/oxy/buffer"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"

	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPoolStripPrefix(t *testing.T) {

	req, err := http.NewRequest("GET", "/garbage/plate/food", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "/garbage/plate/food"

	Convey("When a request is made '/garbage/plate/food', but StripPrefix is set to '/garbage/plate', the URI is correct in the end", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)

			So(r.Body, ShouldBeNil)

			So(r.URL.Path, ShouldEqual, "/food")
			So(r.RequestURI, ShouldEqual, "")

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}
			return &w, nil
		}

		fwd, err := forward.New(forward.Rewriter(&reqRewriter{StripPrefix: "/garbage/plate"}), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldNotBeEmpty)
	})
}
func TestPoolForwardHostHeader(t *testing.T) {

	req, err := http.NewRequest("GET", "http://somewhere.elsewhere.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Host", "somewhere")

	Convey("When a request is made the Host header is properly passed along", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)

			So(r.Body, ShouldBeNil)

			So(r.Header.Get("Host"), ShouldEqual, "somewhere")

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}
			return &w, nil
		}

		fwd, err := forward.New(forward.PassHostHeader(true), forward.Rewriter(&reqRewriter{}), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldNotBeEmpty)
	})
}

func TestPoolForwardHostHeaderNope(t *testing.T) {

	req, err := http.NewRequest("GET", "http://somewhere.elsewhere.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Host", "somewhere")

	Convey("When a request is made the Host header is properly removed", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)

			So(r.Body, ShouldBeNil)

			So(r.Header.Get("Host"), ShouldEqual, "")

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}
			return &w, nil
		}

		fwd, err := forward.New(forward.PassHostHeader(true), forward.Rewriter(&reqRewriter{Headers: []string{"Host"}}), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldNotBeEmpty)
	})
}

func TestPoolReplacePath(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to '/', but ReplacePath is set to '/somewhereelse/', the URI is correct in the end", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)

			So(r.Body, ShouldBeNil)

			So(r.URL.Path, ShouldEqual, "/somewhereelse/")

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}
			return &w, nil
		}

		fwd, err := forward.New(forward.Rewriter(&reqRewriter{To: "/somewhereelse/"}), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldNotBeEmpty)
	})
}

func TestPoolReplacePathNope(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to '/', and ReplacePath is not set, the URI does not change", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)

			So(r.Body, ShouldBeNil)

			So(r.URL.Path, ShouldEqual, "/")

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
			}
			return &w, nil
		}

		fwd, err := forward.New(forward.Rewriter(&reqRewriter{}), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		So(req.URL.Path, ShouldEqual, "/")

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldNotBeEmpty)
	})
}

func TestPoolRoundRobinFair(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a two-member roundrobin is created, they both get hit evenly", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		lb, err := roundrobin.New(fwd)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req) // one
		}

		So(oneCount, ShouldEqual, twoCount)
		So(oneCount, ShouldEqual, 5)
		So(twoCount, ShouldEqual, 5)
	})
}

func TestPoolRoundRobinWeight(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a two-member roundrobin is created, but there is a 2:1 weight difference, they get hit proportionally", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		lb, err := roundrobin.New(fwd)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL, roundrobin.Weight(1))

		lb.UpsertServer(twoURL, roundrobin.Weight(2))

		for i := 0; i < 12; i++ {
			lb.ServeHTTP(rr, req) // one
		}

		So(oneCount, ShouldEqual, 4)
		So(twoCount, ShouldEqual, 8)
	})
}

func TestPoolRoundRobinWeightZero(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a two-member roundrobin is created, but there is a 2^64:1 weight difference, the one-weighted never gets hit in 1000 requests", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		lb, err := roundrobin.New(fwd)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL, roundrobin.Weight(1))

		lb.UpsertServer(twoURL, roundrobin.Weight(math.MaxInt64))

		for i := 0; i < 1000; i++ {
			lb.ServeHTTP(rr, req) // one
		}

		So(oneCount, ShouldEqual, 0)
		So(twoCount, ShouldEqual, 1000)
	})
}

func TestPoolRoundRobinFailWell(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	Convey("When a two-member roundrobin is created with a buffer, and one \"crashes\", the failover is proper", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		// defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)
		// explicit close now
		twoServer.Close()

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		lb, err := roundrobin.New(fwd)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		for i := 0; i < 10; i++ {
			buff.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		So(oneCount, ShouldEqual, 10)
		So(twoCount, ShouldEqual, 0)
	})
}
