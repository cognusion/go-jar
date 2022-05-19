package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"html/template"
	//"log"
	"net/http"
	"net/http/httptest"

	//"os"
	"testing"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	//TimingOut = log.New(os.Stderr, "[TIMING] ", OutFormat)
	//ErrorOut = log.New(os.Stderr, "[ERROR] ", OutFormat)

	// Set up errorhandler
	// TODO: Mock something for this :(
	templ := "tests/errorhandler/errors.tmpl"
	Conf.Set(ConfigErrorHandlerTemplate, templ)

	var err error
	ErrorTemplate, err = template.ParseFiles(templ)
	if err != nil {
		ErrorOut.Fatalf("Unable to parse error template '%s': %s", templ, err)
	}
}

// TestErrorHandlerTemplateNoError tests TemplateErrorHandler when there is no error
func TestErrorHandlerTemplateNoError(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("Hello World"))
	})

	Convey("When a request is made, and there is no error, the template error handler demurs", t, func() {
		rr := httptest.NewRecorder()
		e := ErrorWrapper{HandleTemplateWrapper}
		handler := e.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldEqual, "Hello World")

	})
}

// TestErrorHandlerTemplateError tests TemplateErrorHandler when there is and error
func TestErrorHandlerTemplateError(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(403)
		w.Write([]byte("Hello World"))
	})

	Convey("When a request is made, and there is an error, the template error handler does the right thing", t, func() {
		rr := httptest.NewRecorder()
		e := ErrorWrapper{HandleTemplateWrapper}
		handler := e.Handler(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		req.URL.Scheme = "http"

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)
		So(rr.Body.String(), ShouldNotEqual, "Hello World")
		So(rr.Body.String(), ShouldContainSubstring, "Hello World")
	})
}

// TestErrorHandlerGenericNoError tests GenericErrorHandler when there is no error
func TestErrorHandlerGenericNoError(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("Hello World"))
	})

	Convey("When a request is made, and there is no error, the generic error handler demurs", t, func() {
		rr := httptest.NewRecorder()
		e := ErrorWrapper{HandleGenericWrapper}
		handler := e.Handler(testHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Body.String(), ShouldEqual, "Hello World")

	})
}

// TestErrorHandlerGenericError tests GenericErrorHandler when there is and error
func TestErrorHandlerGenericError(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(403)
		w.Write([]byte("Hello World"))
	})

	Convey("When a request is made, and there is an error, the generic error handler does the right thing", t, func() {
		rr := httptest.NewRecorder()
		e := ErrorWrapper{HandleGenericWrapper}
		handler := e.Handler(testHandler)

		// Set up a dummy requestID
		ctx := WithRqID(req.Context(), "abc123")
		req = req.WithContext(ctx)

		req.URL.Scheme = "http"

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusForbidden)
		So(rr.Body.String(), ShouldEqual, "Hello World")

	})
}
