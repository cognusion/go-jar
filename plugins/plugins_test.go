package plugins

import (
	. "github.com/smartystreets/goconvey/convey"

	"net/http"
	"net/http/httptest"
	"testing"
)

func testHandler(next http.Handler) http.Handler {
	o := []byte("Hello World")
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write(o)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// testHandlerSrc should be identical to testHandler, less the
// single import line. NOTE: If we pre-import net/http (as
// everything will always need it) there is no statistical
// improvement, as the overall process takes the same amount
// of time.
const testHandlerSrc = `
import "net/http"
func TestHandler(next http.Handler) http.Handler {
	o := []byte("Hello World")
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write(o)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
`

const testHandlerConfigSrc = `
import "net/http"
import "fmt"
var config = make(map[string]string)
func SetConfig(c map[string]string) {
	config = c
}

func TestHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		for k,v := range config {
			w.Write([]byte(fmt.Sprintf("%s = %s\n",k,v)))
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
`

// badHandlerSrc should be go-compilable but not a proper
// handler.
const badHandlerSrc = `
func Bad() {

}
`

// noImportHandlerSrc should be identical to testHandlerSrc,
// but omit the "import" and thus throw a bolt.
const noImportHandlerSrc = `
func TestHandler(next http.Handler) http.Handler {
	o := []byte("Hello World")
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write(o)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
`

// bugHandlerSrc should not compile due to a defect.
const bugHandlerSrc = `
import "net/http"
func TestHandler(bob http.Handler) http.Handler {
	o := []byte("Hello World")
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write(o)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
`

func Test_HandlerPlugin(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to a runtime-loaded HandlerPlugin, it works correctly.", t, func() {

		phandler, err := NewHandlerPlugin(testHandlerSrc, "TestHandler")
		So(err, ShouldBeNil)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})

		rr := httptest.NewRecorder()

		handler := phandler.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldResemble, "Hello World")
	})

	Convey("When a request is made to a runtime-loaded HandlerPlugin and a Config is passed, it works correctly.", t, func() {

		var c = map[string]string{
			"hello":   "world",
			"goodbye": "moon",
			"i love":  "you",
		}

		phandler, err := NewHandlerPluginWithConfig(testHandlerConfigSrc, "TestHandler", c)
		So(err, ShouldBeNil)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})

		rr := httptest.NewRecorder()

		handler := phandler.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldContainSubstring, "hello = world")
	})

	Convey("When a HandlerPlugin is loaded that forgot its imports, it fails properly.", t, func() {

		phandler, err := NewHandlerPlugin(noImportHandlerSrc, "TestHandler")
		So(err, ShouldNotBeNil)
		So(phandler, ShouldBeNil)
	})

	Convey("When a HandlerPlugin is loaded that is not a proper handlerfunc, it fails properly.", t, func() {

		phandler, err := NewHandlerPlugin(badHandlerSrc, "Bad")
		So(err, ShouldNotBeNil)
		So(phandler, ShouldBeNil)
	})

	Convey("When a HandlerPlugin is loaded that has a compile-time defect, it fails properly.", t, func() {

		phandler, err := NewHandlerPlugin(bugHandlerSrc, "TestHandler")
		So(err, ShouldNotBeNil)
		So(phandler, ShouldBeNil)
	})

	Convey("When a HandlerPlugin is loaded correctly, but is called with the wrong funcName, it fails properly.", t, func() {

		phandler, err := NewHandlerPlugin(testHandlerSrc, "Tandler")
		So(err, ShouldNotBeNil)
		So(phandler, ShouldBeNil)
	})

	Convey("When a HandlerPlugin is loaded correctly, the Handler function is copied out of context and executed, it works correctly.", t, func() {

		var hfunc func(http.Handler) http.Handler

		{
			phandler, err := NewHandlerPlugin(testHandlerSrc, "TestHandler")
			So(err, ShouldBeNil)

			hfunc = phandler.CopyHandler()

			*phandler = HandlerPlugin{} // for emphasis
		}

		{
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			})

			rr := httptest.NewRecorder()

			handler := hfunc(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.String(), ShouldResemble, "Hello World")
		}
	})
}

func Benchmark_NativeHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	nilHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := testHandler(nilHandler)

	rr := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_YaegiHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	nilHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	phandler, err := NewHandlerPlugin(testHandlerSrc, "TestHandler")
	if err != nil {
		panic(err)
	}
	rr := httptest.NewRecorder()

	handler := phandler.Handler(nilHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}
