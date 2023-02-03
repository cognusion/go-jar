package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func init() {
	ErrorOut = log.New(io.Discard, "", 0) // Silence error output, explicitly

}

func TestHandleFinisherLowerCase(t *testing.T) {

	Convey("When a request for a known-finisher is made, and the name is lower-cased, it is found", t, func() {
		finisher, err := HandleFinisher("forbidden", nil)
		So(err, ShouldBeNil)
		So(finisher, ShouldNotBeNil)
	})
}

func TestHandleFinisherMixedCase(t *testing.T) {
	Convey("When a request for a known-finisher is made, and the name is mix-cased, it is found", t, func() {
		finisher, ok := HandleFinisher("ForBiDDeN", nil)
		So(ok, ShouldBeNil)
		So(finisher, ShouldNotBeNil)
	})
}

func TestFinisherForbidden(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and the Forbidden finisher is used, the request is Forbidden", t, func() {

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Forbidden)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)
	})
}

func TestFinisherStatusForbidden(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	sf := StatusFinisher(http.StatusForbidden).Finisher

	Convey("When a request is made, and the StatusFinisher is used and set to 403 Forbidden, the request is Forbidden", t, func() {

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(sf)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)
	})
}

func TestFinisherStatusOk(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	sf := StatusFinisher(http.StatusOK).Finisher

	Convey("When a request is made, and the StatusFinisher is used and set to 200 Ok, the request is Ok", t, func() {

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(sf)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestFinisherRedirect(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and the Redirect finisher is used with MovedPermanently, the request is redirected thusly", t, func() {

		rr := httptest.NewRecorder()

		red := Redirect{
			URL:  "http://somewhere.com/",
			Code: http.StatusMovedPermanently,
		}
		handler := http.HandlerFunc(red.Finisher)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusMovedPermanently)
		So(rr.Header().Get("Location"), ShouldEqual, "http://somewhere.com/")
	})
}

func TestFinisherRedirectRegexp(t *testing.T) {

	req, err := http.NewRequest("GET", "http://elsewhere.com/files.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	//req.Host = "elsewhere.com"

	Convey("When a request is made, and the Redirect finisher is used with MovedPermanently, the request is redirected thusly", t, func() {

		rr := httptest.NewRecorder()

		re, rerr := regexp.Compile("(?i)(.*)where")
		So(rerr, ShouldBeNil)

		red := Redirect{
			URL:    "http://somewhere$1.com%1",
			Code:   http.StatusMovedPermanently,
			Regexp: re,
		}
		handler := http.HandlerFunc(red.Finisher)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusMovedPermanently)
		So(rr.Header().Get("Location"), ShouldEqual, "http://somewhereelse.com/files.html")
	})
}

func TestFinisherRedirectWTF(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and the Redirect finisher is used with a non-redirect code, the request is redirected with MovedPermanent instead", t, func() {

		rr := httptest.NewRecorder()

		red := Redirect{
			URL:  "http://somewhere.com/",
			Code: http.StatusForbidden,
		}
		handler := http.HandlerFunc(red.Finisher)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusMovedPermanently)
	})
}
