package jar

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var cc *CacheCluster

func cacheTestSetup() {
	//TimingOut = log.New(os.Stdout, "[TIMING] ", 0)
	//DebugOut = log.New(os.Stdout, "[DEBUG] ", log.Lshortfile)
	//ErrorOut = log.New(os.Stdout, "[ERROR] ", log.Lshortfile)
	if cc == nil {
		cc = NewCacheCluster(":8086", 10*time.Second, []string{"http://127.0.0.1:8086"})
	}
}

func Test_PageCache(t *testing.T) {
	cacheTestSetup()

	cacheName := "tpc"
	p, perr := cc.NewPageCache(cacheName, 128<<20, 0, 0, "private") // 128MB cache, no size limit, no expiration
	if perr != nil {
		panic(perr)
	}
	p.syncCacheIt = true // testing faster than the async cachewrites can hit

	Convey("When a new Page is created it looks correct", t, func() {

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello world!!"))
			w.Write([]byte(r.URL.String()))
		})

		handler := p.Handler(testHandler)

		Convey("Cache miss looks correct", func() {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.Bytes(), ShouldResemble, []byte("Hello world!!/"))
			So(rr.Header().Get("Cache-Control"), ShouldContainSubstring, "private")

			Convey("And a manual look at the cache shows it's there", func() {
				//time.Sleep(10 * time.Millisecond)
				doc, ok := p.cluster.Get(cacheName, url.PathEscape("/"))
				So(ok, ShouldBeTrue)
				So(string(doc.([]byte)), ShouldContainSubstring, string([]byte("Hello world!!/")))
			})
		})

		Convey("Cache hit looks correct", func() {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.Bytes(), ShouldResemble, []byte("Hello world!!/"))
			So(rr.Header().Get("Cache-Control"), ShouldContainSubstring, "private")
		})
	})
}

func Test_PageCacheDupe(t *testing.T) {
	cacheTestSetup()

	Convey("When a new Page is created it looks correct", t, func() {
		cacheName := "tpc123"
		_, perr := cc.NewPageCache(cacheName, 128<<20, 0, 0, "") // 128MB cache, no size limit, no expiration
		So(perr, ShouldBeNil)
		_, perr2 := cc.NewPageCache(cacheName, 128<<20, 0, 0, "") // 128MB cache, no size limit, no expiration
		So(perr2, ShouldEqual, CacheAlreadyDefinedError)
	})

}

func Test_PageCacheTooBig(t *testing.T) {
	cacheTestSetup()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	cacheName := "tpctb"
	p, perr := cc.NewPageCache(cacheName, 128<<20, 1, 0, "private") // 128MB cache, 1B size limit, no expiration
	if perr != nil {
		panic(perr)
	}
	p.syncCacheIt = true // testing faster than the async cachewrites can hit

	Convey("When a new Page is created it looks correct", t, func() {

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello world!!"))
			w.Write([]byte(r.URL.String()))
		})

		handler := p.Handler(testHandler)

		Convey("Initial cache miss looks correct", func() {
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.Bytes(), ShouldResemble, []byte("Hello world!!/"))
			So(rr.Header().Get("Cache-Control"), ShouldBeBlank)

			Convey("And a manual look at the cache shows it's not there", func() {
				_, ok := p.cluster.Get(cacheName, "/")
				So(ok, ShouldBeFalse)
			})
		})

		Convey("Second cache miss looks correct", func() {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Body.Bytes(), ShouldResemble, []byte("Hello world!!/"))
			So(rr.Header().Get("Cache-Control"), ShouldBeBlank)

			Convey("And a manual look at the cache shows it's not there", func() {
				_, ok := p.cluster.Get(cacheName, "/")
				So(ok, ShouldBeFalse)
			})
		})
	})
}

func Test_IsCacheable(t *testing.T) {
	Convey("A trival GET with an empty-yet-OK response is Cacheable!", t, func() {
		req, _ := http.NewRequest("GET", "/", nil)
		resp := http.Response{
			StatusCode: 200,
		}
		ok, reasons := isCacheableReasons(req.Header, req.Method, resp.Header, resp.StatusCode)
		/*
			for _, reason := range *reasons {
				Printf("Reason: %d %s\n", reason, reason.String())
			}
		*/
		So(ok, ShouldBeTrue)
		So(reasons, ShouldBeNil)
	})

	Convey("A trival GET with an empty-yet-NOT OK response isn't Cacheable!", t, func() {
		req, _ := http.NewRequest("GET", "/", nil)
		resp := http.Response{
			StatusCode: 500,
		}
		ok, reasons := isCacheableReasons(req.Header, req.Method, resp.Header, resp.StatusCode)
		/*
			for _, reason := range *reasons {
				Printf("Reason: %d %s\n", reason, reason.String())
			}
		*/
		So(ok, ShouldBeFalse)
		So(reasons, ShouldNotBeEmpty)
	})

	Convey("A trival POST with an empty-yet-OK response isn't Cacheable!", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		resp := http.Response{
			StatusCode: 200,
		}
		ok, reasons := isCacheableReasons(req.Header, req.Method, resp.Header, resp.StatusCode)
		/*
			for _, reason := range *reasons {
				Printf("Reason: %d %s\n", reason, reason.String())
			}
		*/
		So(ok, ShouldBeFalse)
		So(reasons, ShouldNotBeEmpty)
	})
}

func Benchmark_IsCacheableSimple(b *testing.B) {
	req, _ := http.NewRequest("GET", "/", nil)
	resp := http.Response{
		StatusCode: 200,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isCacheable(req.Header, req.Method, resp.Header, resp.StatusCode)
	}
}
