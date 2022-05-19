package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {
	if ErrorOut == nil {
		ErrorOut = log.New(os.Stderr, "[ERROR] ", OutFormat)
	}
}

func MapfileInit() {
	Conf = InitConfig()
	Conf.Set("mapfiles", map[string]string{
		"endpoints": "tests/maps/endpoints.map",
		"ids":       "tests/maps/ids.map",
	})

	Conf.Set(ConfigSwitchHandlerStripPrefix, "stjar")

	Conf.Set(ConfigURLRouteHeaders, true)
	Conf.Set(ConfigURLRouteIDHeaderName, "X-JARID")
	Conf.Set(ConfigURLRouteEndpointHeaderName, "X-JARE")
	Conf.Set(ConfigURLRouteNameHeaderName, "X-JARNAME")

	if mapfiles := Conf.GetStringMapString("mapfiles"); len(mapfiles) == 0 {
		ErrorOut.Fatalf("mapfiles config is empty!\n")
	} else {
		// Load the maps
		for n, f := range mapfiles {
			err := SwitchMaps.Load(n, f)
			if err != nil {
				ErrorOut.Fatalf("Error loading map file '%s': %s\n", f, err)
			}
		}
	}
}

func TestSwitch_IPhost(t *testing.T) {

	MapfileInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	host := "127.0.0.1"

	req.Host = host

	Convey("When a request is made, and the host is an IP address, SwitchHandler declines the request properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Next handler errantly executed!")
		})

		rr := httptest.NewRecorder()

		handler := SwitchHandler(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestSwitch_Urlname(t *testing.T) {

	MapfileInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	host := "testname1.jar.com"
	urlname := "testname1"
	id := "48357034875034"
	endpoint := "ep1"

	req.Host = host

	Convey("When a request is made, and the host is known org, SwitchHandler sets the X-JAR* headers properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.Header.Get("X-JARNAME"), ShouldEqual, urlname)
			So(r.Header.Get("X-JARID"), ShouldEqual, id)
			So(r.Header.Get("X-JARE"), ShouldEqual, endpoint)
		})

		rr := httptest.NewRecorder()

		handler := SwitchHandler(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestSwitch_StripUrlname(t *testing.T) {

	MapfileInit()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	host := "stjar98708457340857034957304-testname1.jar.com"
	urlname := "testname1"
	id := "48357034875034"
	endpoint := "ep1"

	req.Host = host

	Convey("When a request is made, and the host is a tile-subhost (stjar<whatever>-org) of a known org, SwitchHandler sets the X-JAR* headers properly", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(r.Header.Get("X-JARNAME"), ShouldEqual, urlname)
			So(r.Header.Get("X-JARID"), ShouldEqual, id)
			So(r.Header.Get("X-JARE"), ShouldEqual, endpoint)
		})

		rr := httptest.NewRecorder()

		handler := SwitchHandler(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}
