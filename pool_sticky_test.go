package jar

import (
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vulcand/oxy/buffer"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/oxy/roundrobin/stickycookie"

	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestPoolRoundRobinSticky(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	cookieName := "STICKYCOOKIE"

	Convey("When a two-member roundrobin is created with a buffer, and requests are pinned to one instance, they stay that way", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)
		// explicit close now
		//twoServer.Close()

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		sc := http.Cookie{
			Name:  cookieName,
			Value: twoServer.URL,
		}
		req.AddCookie(&sc)
		sticky := roundrobin.EnableStickySession(roundrobin.NewStickySession(cookieName))

		lb, err := roundrobin.New(fwd, sticky)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		for i := 0; i < 10; i++ {
			buff.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		So(oneCount, ShouldEqual, 0)
		So(twoCount, ShouldEqual, 10)
	})
}

func TestPoolRoundRobinStickyFailReissue(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	cookieName := "STICKYCOOKIE"

	Convey("When a two-member roundrobin is created with a buffer, and requests are pinned to one instance, but that instance fails, they get bounced over with a new cookie", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		//defer twoServer.Close()
		//twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)
		// explicit close now
		twoServer.Close()

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		sc := http.Cookie{
			Name:  cookieName,
			Value: twoServer.URL,
		}
		req.AddCookie(&sc)
		sticky := roundrobin.EnableStickySession(roundrobin.NewStickySession(cookieName))

		lb, err := roundrobin.New(fwd, sticky)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		//lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		for i := 0; i < 10; i++ {
			buff.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
			cookies := rr.Result().Cookies()
			So(len(cookies), ShouldBeGreaterThan, 0)
			So(cookies[0].Name, ShouldEqual, cookieName)
			So(cookies[0].Value, ShouldEqual, oneServer.URL)
		}

		So(oneCount, ShouldEqual, 10)
		So(twoCount, ShouldEqual, 0)
	})
}

func TestPoolRoundRobinStickyCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	cookieName := "STICKYCOOKIE"

	Convey("When a two-member roundrobin is created with a buffer and using a sticky cookie, and requests are pinned to one instance, they stay that way", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)
		// explicit close now
		//twoServer.Close()

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		ao, err := setupStickyCookie([]byte("1234567890abcdef"), 0)
		So(err, ShouldBeNil)

		cookieValue := ao.Get(twoURL)
		sc := http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		}
		req.AddCookie(&sc)

		sticky := roundrobin.EnableStickySession(roundrobin.NewStickySession(cookieName).SetCookieValue(ao))

		lb, err := roundrobin.New(fwd, sticky)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		for i := 0; i < 10; i++ {
			buff.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
		}

		So(oneCount, ShouldEqual, 0)
		So(twoCount, ShouldEqual, 10)
	})
}

func TestPoolRoundRobinStickyCookieOptions(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	cookieName := "STICKYCOOKIE"

	Convey("When a two-member roundrobin is created with a buffer and using a sticky cookie and with HTTPOnly and Secure set, requests pin to one instance, they stay that way, and the cookies are correct", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)
		// explicit close now
		//twoServer.Close()

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		ao, err := setupStickyCookie([]byte("1234567890abcdef"), 0)
		So(err, ShouldBeNil)

		sticky := roundrobin.EnableStickySession(roundrobin.NewStickySessionWithOptions(cookieName, roundrobin.CookieOptions{HTTPOnly: true, Secure: true}).SetCookieValue(ao))

		lb, err := roundrobin.New(fwd, sticky)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		var (
			oneUp bool
			twoUp bool
		)
		for i := 0; i < 10; i++ {
			buff.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Result().Cookies(), ShouldNotBeEmpty)
			for _, c := range rr.Result().Cookies() {
				if c.Name == cookieName {
					So(c.HttpOnly, ShouldBeTrue)
					So(c.Secure, ShouldBeTrue)
					req.AddCookie(c)
				}
			}

			if !oneUp && !twoUp {
				if oneCount > 0 {
					oneUp = true
				} else {
					twoUp = true
				}
			}
		}

		if twoUp {
			So(oneCount, ShouldEqual, 0)
			So(twoCount, ShouldEqual, 10)
		} else {
			So(oneCount, ShouldEqual, 10)
			So(twoCount, ShouldEqual, 0)
		}
	})
}

func TestPoolRoundRobinStickyCookieOptionsDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	cookieName := "STICKYCOOKIE"

	Convey("When a two-member roundrobin is created with a buffer and using a sticky cookie and with HTTPOnly and Secure defaulting (false), requests pin to one instance, they stay that way, and the cookies are correct", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)
		// explicit close now
		//twoServer.Close()

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		ao, err := setupStickyCookie([]byte("1234567890abcdef"), 0)
		So(err, ShouldBeNil)

		sticky := roundrobin.EnableStickySession(roundrobin.NewStickySession(cookieName).SetCookieValue(ao))

		lb, err := roundrobin.New(fwd, sticky)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		var (
			oneUp bool
			twoUp bool
		)
		for i := 0; i < 10; i++ {
			buff.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
			So(rr.Result().Cookies(), ShouldNotBeEmpty)
			for _, c := range rr.Result().Cookies() {
				if c.Name == cookieName {
					So(c.HttpOnly, ShouldBeFalse)
					So(c.Secure, ShouldBeFalse)
					req.AddCookie(c)
				}
			}

			if !oneUp && !twoUp {
				if oneCount > 0 {
					oneUp = true
				} else {
					twoUp = true
				}
			}
		}

		if twoUp {
			So(oneCount, ShouldEqual, 0)
			So(twoCount, ShouldEqual, 10)
		} else {
			So(oneCount, ShouldEqual, 10)
			So(twoCount, ShouldEqual, 0)
		}
	})
}

func TestPoolRoundRobinStickyCookieFailReissue(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	cookieName := "STICKYCOOKIE"

	Convey("When a two-member roundrobin is created with a buffer and using a sticky cookie, and requests are pinned to one instance, but that instance fails, they get bounced over with a new cookie", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		//defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)
		// explicit close now
		twoServer.Close()

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		ao, err := setupStickyCookie([]byte("1234567890abcdef"), 0)
		So(err, ShouldBeNil)

		cookieValue := ao.Get(twoURL)
		sc := http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		}
		req.AddCookie(&sc)

		sticky := roundrobin.EnableStickySession(roundrobin.NewStickySession(cookieName).SetCookieValue(ao))

		lb, err := roundrobin.New(fwd, sticky)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		//lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		for i := 0; i < 10; i++ {
			buff.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusOK)
			cookies := rr.Result().Cookies()
			So(len(cookies), ShouldBeGreaterThan, 0)
			So(cookies[0].Name, ShouldEqual, cookieName)

			cv, cErr := ao.FindURL(cookies[0].Value, []*url.URL{oneURL, twoURL})
			So(cErr, ShouldBeNil)
			So(cv, ShouldEqual, oneURL)
			//So(ao.Normalize(cookies[0].Value), ShouldEqual, oneServer.URL)
		}

		So(oneCount, ShouldEqual, 10)
		So(twoCount, ShouldEqual, 0)
	})
}

func TestPoolRoundRobinStickyCookieExpireReissue(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	logrusLogger := logrus.New()
	logrusLogger.Out = io.Discard

	cookieName := "STICKYCOOKIE"

	Convey("When a two-member roundrobin is created with a buffer and using a sticky cookie, and requests are pinned to one instance, but the cookie expires, they get a new cookie pinned to the other instance", t, func(c C) {

		rr := httptest.NewRecorder()

		oneCount := 0
		one := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oneCount++
			w.Write([]byte("Ok"))
		})
		oneServer := httptest.NewServer(one)
		defer oneServer.Close()
		oneURL, err := url.Parse(oneServer.URL)
		So(err, ShouldBeNil)

		twoCount := 0
		two := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			twoCount++
			w.Write([]byte("Ok"))
		})
		twoServer := httptest.NewServer(two)
		defer twoServer.Close()
		twoURL, err := url.Parse(twoServer.URL)
		So(err, ShouldBeNil)

		fwd, err := forward.New()
		So(err, ShouldBeNil)

		// First ao has a 1ns expiration, so we know it will be expired
		firstao, err := setupStickyCookie([]byte("1234567890abcdef"), 1*time.Nanosecond)
		So(err, ShouldBeNil)

		// ao has a 1s expiration, so it'll last a bit
		ao, err := setupStickyCookie([]byte("1234567890abcdef"), 1*time.Second)
		So(err, ShouldBeNil)

		cookieValue := firstao.Get(twoURL)
		sc := http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		}
		req.AddCookie(&sc)

		sticky := roundrobin.EnableStickySession(roundrobin.NewStickySession(cookieName).SetCookieValue(ao))

		lb, err := roundrobin.New(fwd, sticky)
		So(err, ShouldBeNil)

		lb.UpsertServer(oneURL)
		lb.UpsertServer(twoURL)

		buff, err := buffer.New(lb, buffer.Retry(fmt.Sprintf("IsNetworkError() && Attempts() < %d", 2)), buffer.Logger(logrusLogger))
		So(err, ShouldBeNil)

		buff.ServeHTTP(rr, req)
		So(rr.Code, ShouldEqual, http.StatusOK)
		cookies := rr.Result().Cookies()
		So(len(cookies), ShouldBeGreaterThan, 0)
		So(cookies[0].Name, ShouldEqual, cookieName)
		cv, cErr := ao.FindURL(cookies[0].Value, []*url.URL{oneURL, twoURL})
		So(cErr, ShouldBeNil)
		So(cv, ShouldEqual, oneURL)
		//So(ao.Normalize(cookies[0].Value), ShouldEqual, oneServer.URL)

		So(oneCount, ShouldEqual, 1)
		So(twoCount, ShouldEqual, 0)
	})
}

func setupStickyCookie(clearKey []byte, cookieLife time.Duration) (stickycookie.CookieValue, error) {
	var (
		ao  stickycookie.CookieValue
		err error
	)

	if cookieLife > 0 {
		ao, err = stickycookie.NewAESValue(clearKey, cookieLife)
		if err != nil {
			return nil, err
		}
	} else {
		ao, err = stickycookie.NewAESValue(clearKey, time.Duration(0))
		if err != nil {
			return nil, err
		}
	}
	return ao, nil
}
