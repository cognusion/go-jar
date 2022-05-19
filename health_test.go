package jar

import (
	"github.com/cognusion/go-health"
	. "github.com/smartystreets/goconvey/convey"

	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	InitFuncs.Call()
}

func TestStack(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to the Stack dumper, it is OK, text, and at least starts to look ok", t, func() {

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Stack)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Header().Get("Content-Type"), ShouldStartWith, "text/plain")
		So(rr.Body.String(), ShouldStartWith, "goroutine")
	})
}

func TestHealthCheckAsync(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to the HealthCheck, it is OK, application/json", t, func() {

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(healthCheckAsync)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Header().Get("Content-Type"), ShouldEqual, "application/json")

		Convey("and is marshallable into an Check", func() {
			Check, err := health.NewCheckfromJSON(rr.Body.Bytes())
			So(err, ShouldBeNil)
			So(Check.OverallStatus, ShouldNotBeEmpty)
		})

	})
}

func TestHealthCheckSync(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made to the HealthCheck, it is OK, application/json", t, func() {

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(healthCheckSync)

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
		So(rr.Header().Get("Content-Type"), ShouldEqual, "application/json")

		Convey("and is marshallable into an Check", func() {
			Check, err := health.NewCheckfromJSON(rr.Body.Bytes())
			So(err, ShouldBeNil)
			So(Check.OverallStatus, ShouldNotBeEmpty)
		})

	})
}

func TestStringToHealthCheckStatus(t *testing.T) {

	Convey("When a known good HealthCheckStatus string is used to make a HealthCheckStatus, everything is ok", t, func() {
		hc, err := StringToHealthCheckStatus("ok")
		So(err, ShouldBeNil)
		So(hc.String(), ShouldEqual, "Ok")
	})
}

func TestStringToHealthCheckStatusOops(t *testing.T) {

	Convey("When a known bad HealthCheckStatus string is used to make a HealthCheckStatus, everything is error", t, func() {
		_, err := StringToHealthCheckStatus("crap")
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrNoSuchHealthCheckStatus)
	})
}

func Benchmark_HealthCheckAsync(b *testing.B) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	handler := http.HandlerFunc(healthCheckAsync)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Build a healthcheck
	hc := health.NewCheck()
	CurrentHealthCheck.Store(*getHC(&hc))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func Benchmark_HealthCheckSync(b *testing.B) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	handler := http.HandlerFunc(healthCheckSync)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}
