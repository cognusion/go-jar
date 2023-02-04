package jar

import (
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPathStripPrefix(t *testing.T) {

	req, err := http.NewRequest("GET", "/garbage/plate/food", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "/garbage/plate/food"

	Convey("When a request is made, and prefixes are stripped, that prefix should not be in the Request or its URL", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			So(r.URL.EscapedPath(), ShouldEqual, "/food")
			So(r.RequestURI, ShouldEqual, "/food")
		})

		pr := PathStripper{
			Prefix: "/garbage/plate",
		}
		handler := pr.Handler(testHandler)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestPathHost(t *testing.T) {
	router := mux.NewRouter()

	pathw := Path{
		Name:     "TestPathWidget",
		Path:     "/",
		Absolute: false,
		Host:     "somehost.somewhere.com",
		Finisher: "Ok",
	}

	pathd := Path{
		Name:     "default",
		Path:     "/",
		Absolute: false,
		Finisher: "Forbidden",
	}

	if _, err := BuildPath(pathw, 0, router); err != nil {
		t.Errorf("Error creating pathw: %s\n", err)
	}
	if _, err := BuildPath(pathd, 1, router); err != nil {
		t.Errorf("Error creating pathd: %s\n", err)
	}

	Convey("When a request is made, with a host specified, a the correct match is found", t, func() {

		m := mux.RouteMatch{}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "somehost.somewhere.com"

		matched := router.Match(req, &m)
		So(matched, ShouldBeTrue)

		So(m.Route, ShouldNotBeNil)
		So(m.MatchErr, ShouldBeNil)

		host, err := m.Route.GetHostTemplate()
		So(err, ShouldBeNil)
		So(host, ShouldEqual, "somehost.somewhere.com")

	})
}

func TestPathForbidden(t *testing.T) {
	router := mux.NewRouter()

	cback := *Conf                   // back it up
	defer func() { *Conf = cback }() // restore it

	Conf.Set(ConfigForbiddenPaths, []string{
		"/.*/thing",
	})

	pathw := Path{
		Name:     "TestPathWidget",
		Path:     "/something",
		Finisher: "Ok",
	}

	pathd := Path{
		Name:     "AForbiddenPath",
		Path:     "/some/thing",
		Finisher: "Ok",
	}

	Convey("When a Path is built that does not match a ForbiddenPath, it creates without an error", t, func() {
		_, err := BuildPath(pathw, 0, router)
		So(err, ShouldBeNil)
	})

	Convey("When a Path is built that matches a ForbiddenPath, it fails with the correct error", t, func() {
		_, err := BuildPath(pathd, 1, router)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "matches a stated ForbiddenPath")
	})
}

func TestPathHostNope(t *testing.T) {
	router := mux.NewRouter()

	pathw := Path{
		Name:     "TestPathWidget",
		Path:     "/",
		Absolute: false,
		Host:     "somehost.somewhere.com",
		Finisher: "Ok",
	}

	if _, err := BuildPath(pathw, 0, router); err != nil {
		t.Errorf("Error creating pathw: %s\n", err)
	}

	Convey("When a request is made, with a host specified, no match is found", t, func() {

		m := mux.RouteMatch{}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "somehost.somewhereelse.com"

		matched := router.Match(req, &m)
		So(matched, ShouldBeFalse)

		So(m.Route, ShouldBeNil)
		So(m.MatchErr, ShouldNotBeNil)
	})
}

func TestPathHostPortUnset(t *testing.T) {
	router := mux.NewRouter()

	pathw := Path{
		Name:     "TestPathWidget",
		Path:     "/",
		Absolute: false,
		Host:     "somehost.somewhere.com",
		Finisher: "Ok",
	}

	pathd := Path{
		Name:     "default",
		Path:     "/",
		Absolute: false,
		Finisher: "Forbidden",
	}

	if _, err := BuildPath(pathw, 0, router); err != nil {
		t.Errorf("Error creating pathw: %s\n", err)
	}
	if _, err := BuildPath(pathd, 1, router); err != nil {
		t.Errorf("Error creating pathd: %s\n", err)
	}

	Convey("When a request is made, with a host and port specified, the correct match is found (no port specified)", t, func() {

		m := mux.RouteMatch{}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "somehost.somewhere.com:443"

		matched := router.Match(req, &m)
		So(matched, ShouldBeTrue)

		So(m.Route, ShouldNotBeNil)
		So(m.MatchErr, ShouldBeNil)

		host, err := m.Route.GetHostTemplate()
		So(err, ShouldBeNil)
		So(host, ShouldEqual, "somehost.somewhere.com")

	})
}

func TestPathHostPortSet(t *testing.T) {
	router := mux.NewRouter()

	pathw := Path{
		Name:     "TestPathWidget",
		Path:     "/",
		Absolute: false,
		Host:     "somehost.somewhere.com:443",
		Finisher: "Ok",
	}

	if _, err := BuildPath(pathw, 0, router); err != nil {
		t.Errorf("Error creating pathw: %s\n", err)
	}

	Convey("When a request is made, with a host and port specified, and the port does match, a match is found", t, func() {

		m := mux.RouteMatch{}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "somehost.somewhere.com:443"

		matched := router.Match(req, &m)
		So(matched, ShouldBeTrue)

		So(m.Route, ShouldNotBeNil)
		So(m.MatchErr, ShouldBeNil)

		host, err := m.Route.GetHostTemplate()
		So(err, ShouldBeNil)
		So(host, ShouldEqual, "somehost.somewhere.com:443")

	})
}

func TestPathHostPortSetNope(t *testing.T) {
	router := mux.NewRouter()

	pathw := Path{
		Name:     "TestPathWidget",
		Path:     "/",
		Absolute: false,
		Host:     "somehost.somewhere.com:555",
		Finisher: "Ok",
	}

	if _, err := BuildPath(pathw, 0, router); err != nil {
		t.Errorf("Error creating pathw: %s\n", err)
	}

	Convey("When a request is made, with a host and port specified, and the port does not match, a match is NOT found", t, func() {

		m := mux.RouteMatch{}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "somehost.somewhere.com:443"

		matched := router.Match(req, &m)
		So(matched, ShouldBeFalse)

		So(m.Route, ShouldBeNil)
		So(m.MatchErr, ShouldNotBeNil)

	})
}

func TestPathOptions(t *testing.T) {

	rawpath := make(map[string]interface{})
	rawpath["Name"] = "TestPath"
	rawpath["Path"] = "/"
	rawpath["Finisher"] = "ok"
	rawpath["Options"] = map[string]interface{}{
		"BlindMirrorRequest.Mirrors": []string{"a", "b", "c"},
	}

	Conf.Set(ConfigPaths, []interface{}{rawpath})

	Convey("When a path is set with options, it unmarshals without error", t, func() {
		var err error

		paths := make([]Path, 1)
		err = Conf.UnmarshalKey(ConfigPaths, &paths)
		So(err, ShouldBeNil)

		Convey("and the resulting []Path should be correct", func() {
			So(paths[0].Name, ShouldEqual, "TestPath")
			So(paths[0].Path, ShouldEqual, "/")
			So(paths[0].Finisher, ShouldEqual, "ok")

			po := PathOptions{"BlindMirrorRequest.Mirrors": []string{"a", "b", "c"}}
			So(paths[0].Options, ShouldResemble, po)
		})
	})

}

func TestPathOptionsGetters(t *testing.T) {

	rawpath := make(map[string]interface{})
	rawpath["Name"] = "TestPath"
	rawpath["Path"] = "/"
	rawpath["Finisher"] = "ok"
	rawpath["Options"] = map[string]interface{}{
		"BlindMirrorRequest.Mirrors": []string{"a", "b", "c"},
		"HotString":                  "yes!",
		"ABool":                      true,
	}

	Conf.Set(ConfigPaths, []interface{}{rawpath})

	Convey("When a path is set with options, it unmarshals without error", t, func() {
		var err error

		paths := make([]Path, 1)
		err = Conf.UnmarshalKey(ConfigPaths, &paths)
		So(err, ShouldBeNil)

		Convey("and the resulting []Path should be correct", func() {
			So(paths[0].Name, ShouldEqual, "TestPath")
			So(paths[0].Path, ShouldEqual, "/")
			So(paths[0].Finisher, ShouldEqual, "ok")
		})

		Convey("and the resulting PathOptions.GetStringSlice should look correct", func() {
			So(paths[0].Options.GetStringSlice("BlindMirrorRequest.Mirrors"), ShouldResemble, []string{"a", "b", "c"})
		})

		Convey("and the resulting PathOptions.GetString should look correct", func() {
			So(paths[0].Options.GetString("HotString"), ShouldEqual, "yes!")
		})

		Convey("and the resulting PathOptions.GetBool should look correct", func() {
			So(paths[0].Options.GetBool("ABool"), ShouldBeTrue)
		})

	})

}
