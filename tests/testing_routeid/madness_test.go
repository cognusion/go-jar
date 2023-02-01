package madness

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_PassthroughHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/?ROUTEID=sdfsd8798s9df8ds", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made through a non-inpecting handler, it succeeds", t, func() {
		rr := httptest.NewRecorder()

		handler := PassThroughHandler(TestHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

		Convey("but the proper cookie does not exist", func() {
			c, err := req.Cookie("ROUTEID")
			So(err, ShouldNotBeNil)
			So(c, ShouldBeNil)
		})
	})
}

func Test_ParamInspectionFormValue_Handler(t *testing.T) {
	req, err := http.NewRequest("GET", "/?ROUTEID=sdfsd8798s9df8ds", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, with the ROUTEID on the param list, it succeeds", t, func() {
		rr := httptest.NewRecorder()

		handler := ParamInspectionFormValueHandler(CookieInspectionHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

		Convey("and the proper cookie now exists", func() {
			c, err := req.Cookie("ROUTEID")
			So(err, ShouldBeNil)
			So(c.Value, ShouldEqual, "sdfsd8798s9df8ds")
		})
	})
}

func Test_ParamInspectionFormValueHandlerDupe(t *testing.T) {
	req, err := http.NewRequest("GET", "/?ROUTEID=sdfsd8798s9df8ds", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "ROUTEID",
		Value: "blueblueblue",
		Path:  "/",
	})

	Convey("When a request is made, with the ROUTEID on the param list and a different ROUTEID in a cookie, the parse succeeds", t, func() {
		rr := httptest.NewRecorder()

		handler := ParamInspectionFormValueHandler(CookieInspectionHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

		Convey("but the cookie value is NOT the one from the param list", func() {
			c, err := req.Cookie("ROUTEID")
			So(err, ShouldBeNil)
			So(c.Value, ShouldNotEqual, "sdfsd8798s9df8ds")
		})
	})
}

func Test_ParamInspectionURLQueryHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/?ROUTEID=sdfsd8798s9df8ds", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, with the ROUTEID on the param list, it succeeds", t, func() {
		rr := httptest.NewRecorder()

		handler := ParamInspectionURLQueryHandler(CookieInspectionHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

		Convey("and the proper cookie now exists", func() {
			c, err := req.Cookie("ROUTEID")
			So(err, ShouldBeNil)
			So(c.Value, ShouldEqual, "sdfsd8798s9df8ds")
		})
	})
}

func Test_ParamInspectionURLQueryContains_Handler(t *testing.T) {
	req, err := http.NewRequest("GET", "/?ROUTEID=sdfsd8798s9df8ds", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, with the ROUTEID on the param list, it succeeds", t, func() {
		rr := httptest.NewRecorder()

		handler := ParamInspectionURLQueryContainsHandler(CookieInspectionHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)

		Convey("and the proper cookie now exists", func() {
			c, err := req.Cookie("ROUTEID")
			So(err, ShouldBeNil)
			So(c.Value, ShouldEqual, "sdfsd8798s9df8ds")
		})
	})
}

func Test_ParamInspection_URLQueryContains_Handler_None(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, without the ROUTEID on the param list, it fails", t, func() {
		rr := httptest.NewRecorder()

		handler := ParamInspectionURLQueryContainsHandler(CookieInspectionHandler)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusBadRequest)
	})
}
func Benchmark_PassThroughHandler(b *testing.B) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}
	rr := httptest.NewRecorder()

	handler := PassThroughHandler(TestHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_ParamInspection_FormValue_Handler_Negative(b *testing.B) {
	req, err := http.NewRequest("GET", "/?BLHBLAH=dskjfskfhkseuhfkjsdkj", nil)
	if err != nil {
		b.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := ParamInspectionFormValueHandler(TestHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_ParamInspectionFormValueHandlerPositive(b *testing.B) {
	req, err := http.NewRequest("GET", "/?ROUTEID=dskjfskfhkseuhfkjsdkj", nil)
	if err != nil {
		b.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := ParamInspectionFormValueHandler(TestHandler) //TestHandler) //CookieInspectionHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_ParamInspectionURLQueryHandlerNegative(b *testing.B) {
	req, err := http.NewRequest("GET", "/?BLHBLAH=dskjfskfhkseuhfkjsdkj", nil)
	if err != nil {
		b.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := ParamInspectionURLQueryHandler(TestHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_ParamInspectionURLQueryHandlerPositive(b *testing.B) {
	req, err := http.NewRequest("GET", "/?ROUTEID=dskjfskfhkseuhfkjsdkj", nil)
	if err != nil {
		b.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := ParamInspectionURLQueryHandler(TestHandler) //TestHandler) //CookieInspectionHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_ParamInspectionURLQueryContainsHandlerNegative(b *testing.B) {
	req, err := http.NewRequest("GET", "/?BLHBLAH=dskjfskfhkseuhfkjsdkj", nil)
	if err != nil {
		b.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := ParamInspectionURLQueryContainsHandler(TestHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_ParamInspectionURLQueryContainsHandler_Positive(b *testing.B) {
	req, err := http.NewRequest("GET", "/?ROUTEID=dskjfskfhkseuhfkjsdkj", nil)
	if err != nil {
		b.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := ParamInspectionURLQueryContainsHandler(TestHandler) //TestHandler) //CookieInspectionHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}
