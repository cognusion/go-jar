//go:build plugins

package jar

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func Test_PluginConfig(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("Creating a PluginConfig from a known good config works as expected.", t, func() {

		c := viper.New()
		c.SetConfigFile("tests/plugins/hw.yaml")
		err := c.ReadInConfig()
		So(err, ShouldBeNil)

		pc, err := NewPluginConfig("plugins.helloworld", c)
		So(err, ShouldBeNil)

		h, err := pc.CreateHandler()
		So(err, ShouldBeNil)

		Convey("... and when testing the handler, it too works as expected", func() {
			// Give it a whirl
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})

			rr := httptest.NewRecorder()

			handler := h(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.String(), ShouldResemble, "Hello World")
		})
	})

	Convey("Creating a PluginConfig from a known good config with its own config works as expected.", t, func() {

		c := viper.New()
		c.SetConfigFile("tests/plugins/hw.yaml")
		err := c.ReadInConfig()
		So(err, ShouldBeNil)

		pc, err := NewPluginConfig("plugins.conftest", c)
		So(err, ShouldBeNil)

		h, err := pc.CreateHandler()
		So(err, ShouldBeNil)

		Convey("... and when testing the handler, it too works as expected", func() {
			// Give it a whirl
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})

			rr := httptest.NewRecorder()

			handler := h(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.String(), ShouldContainSubstring, "hello = world")
		})
	})
}

func Test_PluginHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When invoking PluginHandler to glue it all together, everything works as expected", t, func() {

		c := viper.New()
		c.SetConfigFile("tests/plugins/hw.yaml")
		err := c.ReadInConfig()
		So(err, ShouldBeNil)

		h, err := PluginHandler("plugins.helloworld", c)
		So(err, ShouldBeNil)

		Convey("... and when testing the handler, it too works as expected", func() {
			// Give it a whirl
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})

			rr := httptest.NewRecorder()

			handler := h(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.String(), ShouldResemble, "Hello World")
		})
	})
}
