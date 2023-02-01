package reqtls

import (
	"github.com/cognusion/grace/gracehttp"
	"github.com/facebookgo/httpdown"
	. "github.com/smartystreets/goconvey/convey"

	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_ListenAndServeTLS(t *testing.T) {

	Convey("When a request is made to http.ListenAndServeTLS, req.TLS is set", t, func(c C) {

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client")
			c.So(r.TLS, ShouldNotBeNil)
		})

		ts := httptest.NewTLSServer(testHandler)
		defer ts.Close()

		client := ts.Client()
		_, err := client.Get(ts.URL)
		So(err, ShouldBeNil)
	})
}

func Test_Httpdown(t *testing.T) {

	Convey("When a request is made to httpdown.ListenAndServe, req.TLS is set", t, func(c C) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client")
			c.So(r.TLS, ShouldNotBeNil)
		})

		ts := httptest.NewTLSServer(handler)

		server := &http.Server{
			Addr:      ts.Listener.Addr().String(),
			Handler:   handler,
			TLSConfig: ts.TLS,
		}

		ts.Close()

		hd := &httpdown.HTTP{
			StopTimeout: 10 * time.Second,
			KillTimeout: 1 * time.Second,
		}

		e := make(chan error, 1)
		go func() {
			select {
			case e <- httpdown.ListenAndServe(server, hd):
			case <-time.After(4 * time.Second):
			}
		}()

		time.Sleep(1 * time.Second)
		_, err := ts.Client().Get(ts.URL)
		select {
		case cerr := <-e:
			So(cerr, ShouldBeNil)
		default:
		}
		So(err, ShouldBeNil)

	})
}

func Test_GraceServe(t *testing.T) {

	Convey("When a request is made to gracehttp.Serve, req.TLS is set", t, func(c C) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client")
			c.So(r.TLS, ShouldNotBeNil)
		})

		ts := httptest.NewTLSServer(handler)

		server := &http.Server{
			Addr:      ts.Listener.Addr().String(),
			Handler:   handler,
			TLSConfig: ts.TLS,
		}

		ts.Close()

		e := make(chan error, 1)
		go func() {
			select {
			case e <- gracehttp.Serve(server):
			case <-time.After(4 * time.Second):
			}
		}()

		time.Sleep(1 * time.Second)
		_, err := ts.Client().Get(ts.URL)
		select {
		case cerr := <-e:
			So(cerr, ShouldBeNil)
		default:
		}
		So(err, ShouldBeNil)

	})
}

func Test_GraceServeWithOptions(t *testing.T) {

	Convey("When a request is made to gracehttp.Serve, req.TLS is set", t, func(c C) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client")
			c.So(r.TLS, ShouldNotBeNil)
		})

		ts := httptest.NewTLSServer(handler)

		server := &http.Server{
			Addr:      ts.Listener.Addr().String(),
			Handler:   handler,
			TLSConfig: ts.TLS,
		}

		ts.Close()

		e := make(chan error, 1)
		go func() {
			select {
			case e <- gracehttp.ServeWithOptions([]*http.Server{server}, gracehttp.ListenerLimit(100)):
			case <-time.After(4 * time.Second):
			}
		}()

		time.Sleep(1 * time.Second)
		_, err := ts.Client().Get(ts.URL)
		select {
		case cerr := <-e:
			So(cerr, ShouldBeNil)
		default:
		}
		So(err, ShouldBeNil)

	})
}
