package jar

import (
	"bytes"
	"fmt"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vulcand/oxy/forward"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyResponseModifier(t *testing.T) {

	ourResponseModifier := func(resp *http.Response) error {
		delete(resp.Header, "Crap1")
		return nil
	}

	req, err := http.NewRequest("GET", "http://somewhere.elsewhere.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and there is a single ResponseModifier, the response is properly modified", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)
			So(r.Body, ShouldBeNil)

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(map[string][]string),
			}
			w.Header["Crap1"] = []string{"to delete"}
			w.Header["Crap2"] = []string{"to delete"}
			w.Header["Crap3"] = []string{"to delete"}
			w.Header["Crap4"] = []string{"to delete"}
			w.Header["Crap5"] = []string{"to delete"}
			return &w, nil
		}

		fwd, err := forward.New(forward.Rewriter(&reqRewriter{}), forward.ResponseModifier(ourResponseModifier), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldNotBeEmpty)
		So(rr.Header().Get("Crap1"), ShouldBeEmpty) // The one we've deleted
		So(rr.Header().Get("Crap2"), ShouldNotBeEmpty)
		So(rr.Header().Get("Crap3"), ShouldNotBeEmpty)
		So(rr.Header().Get("Crap4"), ShouldNotBeEmpty)
		So(rr.Header().Get("Crap5"), ShouldNotBeEmpty)
	})
}

func TestProxyResponseModifierChain(t *testing.T) {

	var ourResponseModifier ProxyResponseModifier

	prmc := ProxyResponseModifierChain{}

	its := func(key string) ProxyResponseModifier {
		return func(resp *http.Response) error {
			delete(resp.Header, key)
			return nil
		}
	}

	for ix := 1; ix < 6; ix++ {
		prmc.Add(its(fmt.Sprintf("Crap%d", ix)))
	}

	ourResponseModifier = prmc.ToProxyResponseModifier()

	req, err := http.NewRequest("GET", "http://somewhere.elsewhere.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and there is a chain of ResponseModifiers, the response is properly modified", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)
			So(r.Body, ShouldBeNil)

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(map[string][]string),
			}
			w.Header["Crap1"] = []string{"to delete"}
			w.Header["Crap2"] = []string{"to delete"}
			w.Header["Crap3"] = []string{"to delete"}
			w.Header["Crap4"] = []string{"to delete"}
			w.Header["Crap5"] = []string{"to delete"}
			return &w, nil
		}

		fwd, err := forward.New(forward.Rewriter(&reqRewriter{}), forward.ResponseModifier(ourResponseModifier), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldNotBeEmpty)
		So(rr.Header().Get("Crap1"), ShouldBeEmpty)
		So(rr.Header().Get("Crap2"), ShouldBeEmpty)
		So(rr.Header().Get("Crap3"), ShouldBeEmpty)
		So(rr.Header().Get("Crap4"), ShouldBeEmpty)
		So(rr.Header().Get("Crap5"), ShouldBeEmpty)
	})
}

func TestProxyResponseModifierChainError(t *testing.T) {

	var ourResponseModifier ProxyResponseModifier

	prmc := ProxyResponseModifierChain{}

	its := func(key string) ProxyResponseModifier {
		return func(resp *http.Response) error {
			delete(resp.Header, key)
			return nil
		}
	}

	for ix := 1; ix < 4; ix++ {
		prmc.Add(its(fmt.Sprintf("Crap%d", ix)))
	}

	ef := func(resp *http.Response) error {
		return fmt.Errorf("The sky is falling")
	}
	prmc.Add(ef) // broken one

	prmc.Add(its("Crap5")) // another working one, thereafter

	ourResponseModifier = prmc.ToProxyResponseModifier()

	req, err := http.NewRequest("GET", "http://somewhere.elsewhere.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and there is a chain of ResponseModifiers, and an error occurs, the response is completely binned", t, func() {

		dt := DebugTrip{}
		dt.RTFunc = func(r *http.Request) (*http.Response, error) {
			So(r, ShouldNotBeNil)
			So(r.Body, ShouldBeNil)

			w := http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewBufferString("OK")),
				Header:     make(map[string][]string),
			}
			w.Header["Crap1"] = []string{"to delete"}
			w.Header["Crap2"] = []string{"to delete"}
			w.Header["Crap3"] = []string{"to delete"}
			w.Header["Crap4"] = []string{"to delete"}
			w.Header["Crap5"] = []string{"to delete"}
			return &w, nil
		}

		fwd, err := forward.New(forward.Rewriter(&reqRewriter{}), forward.ResponseModifier(ourResponseModifier), forward.RoundTripper(&dt))
		So(err, ShouldBeNil)

		rr := httptest.NewRecorder()

		fwd.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusInternalServerError)
		So(rr.Header(), ShouldBeEmpty)
	})
}
