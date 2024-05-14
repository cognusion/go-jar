package jar

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vulcand/oxy/v2/forward"

	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPoolConsistentHashInvalidSource(t *testing.T) {
	Convey("When an invalid source type is passed, New fails appropriately", t, func() {
		_, err := NewConsistentHashPool("nonsense", "url", nil, nil)
		So(err, ShouldEqual, ErrConsistentHashInvalidSource)
	})
}

func TestPoolConsistentHash(t *testing.T) {

	URI := "/hello"
	URI2 := "/gbye"
	req, err := http.NewRequest("GET", URI, nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err2 := http.NewRequest("GET", URI2, nil)
	if err2 != nil {
		t.Fatal(err2)
	}

	Convey("When a three-member ch is created, with two distinct requests, they balance properly", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("ONE"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("TWO"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)

		threeCount := 0
		three := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			threeCount++
			w.Write([]byte("THREE"))
		})
		threeServer := httptest.NewServer(three)
		defer threeServer.Close()
		threeURL, err := url.Parse(threeServer.URL)
		So(err, ShouldBeNil)

		fwd := forward.New(false)
		lb, err := NewConsistentHashPool("request", "url", nil, fwd)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)
		lb.UpsertServer(threeURL)

		servers := lb.Servers()
		So(len(servers), ShouldEqual, 3)

		lb.ServeHTTP(rr, req) // one
		So(rr.Code, ShouldEqual, http.StatusOK)

		var first *int
		if oneCount > 0 {
			first = &oneCount
			oneCount = 0
		} else if twoCount > 0 {
			first = &twoCount
			twoCount = 0
		} else {
			first = &threeCount
			threeCount = 0
		}

		lb.ServeHTTP(rr, req2) // one
		So(rr.Code, ShouldEqual, http.StatusOK)

		var second *int
		if oneCount > 0 {
			second = &oneCount
			oneCount = 0
		} else if twoCount > 0 {
			second = &twoCount
			twoCount = 0
		} else {
			second = &threeCount
			threeCount = 0
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req2) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req2) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		if first == second {
			So(*first, ShouldEqual, *second)
			So(*first, ShouldEqual, 40)
		} else {
			So(*first, ShouldEqual, 20)
			So(*second, ShouldEqual, 20)
		}

		Convey("... when two are removed, requests continue on the remaining", func(c C) {

			lb.RemoveServer(twoURL)
			lb.RemoveServer(threeURL)

			rr1 := httptest.NewRecorder()
			oneCount = 0
			for i := 0; i < 10; i++ {
				lb.ServeHTTP(rr1, req) // one
				c.So(rr1.Code, ShouldEqual, http.StatusOK)
			}

			rr2 := httptest.NewRecorder()
			for i := 0; i < 10; i++ {
				lb.ServeHTTP(rr2, req2) // one
				c.So(rr2.Code, ShouldEqual, http.StatusOK)
			}

			c.So(oneCount, ShouldEqual, 20)
		})

		Convey("... when three are removed, Service Unavailable", func(c C) {
			rr := httptest.NewRecorder()

			lb.RemoveServer(oneURL)
			lb.RemoveServer(twoURL)
			lb.RemoveServer(threeURL)

			lb.ServeHTTP(rr, req) // one
			c.So(rr.Code, ShouldEqual, http.StatusServiceUnavailable)

			Convey("... ... and one is added back, everything works", func(c C) {
				rr := httptest.NewRecorder()
				err := lb.UpsertServer(oneURL)
				c.So(err, ShouldBeNil)

				lb.ServeHTTP(rr, req) // one
				c.So(rr.Code, ShouldEqual, http.StatusOK)
			})

		})
	})

}

func TestPoolConsistentHashNoHash(t *testing.T) {

	URI := "/hello"
	URI2 := "/gbye"
	req, err := http.NewRequest("GET", URI, nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err2 := http.NewRequest("GET", URI2, nil)
	if err2 != nil {
		t.Fatal(err2)
	}

	Convey("When a two-member ch is created, but no valid hashkey is used requests hotspot on one member", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("ONE"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("TWO"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)

		threeCount := 0
		three := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			threeCount++
			w.Write([]byte("THREE"))
		})
		threeServer := httptest.NewServer(three)
		defer threeServer.Close()
		threeURL, err := url.Parse(threeServer.URL)
		So(err, ShouldBeNil)

		fwd := forward.New(false)
		lb, err := NewConsistentHashPool("header", "url", nil, fwd)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)
		lb.UpsertServer(threeURL)

		servers := lb.Servers()
		So(len(servers), ShouldEqual, 3)

		lb.ServeHTTP(rr, req) // one
		So(rr.Code, ShouldEqual, http.StatusOK)

		var first *int
		if oneCount > 0 {
			first = &oneCount
			oneCount = 0
		} else if twoCount > 0 {
			first = &twoCount
			twoCount = 0
		} else {
			first = &threeCount
			threeCount = 0
		}

		lb.ServeHTTP(rr, req2) // one
		So(rr.Code, ShouldEqual, http.StatusOK)

		var second *int
		if oneCount > 0 {
			second = &oneCount
			oneCount = 0
		} else if twoCount > 0 {
			second = &twoCount
			twoCount = 0
		} else {
			second = &threeCount
			threeCount = 0
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req2) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		for i := 0; i < 10; i++ {
			lb.ServeHTTP(rr, req2) // one
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		So(first, ShouldEqual, second)
		So(*first, ShouldEqual, *second)
		So(*first, ShouldEqual, 40)
	})
}
