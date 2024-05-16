package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	//ErrorOut = log.New(os.Stderr, "[ERROR] ", OutFormat)

}

// TestAccessAddIPs tests Access adding just IP addresses
func TestAccessAddIPs(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)
	})
}

// TestAccessAddBadIP tests Access adding a bad IP address
func TestAccessAddBadIPNetworks(t *testing.T) {

	Convey("When NewAccessFromStrings is called with a bad IP address, it errors accordingly", t, func() {
		_, err := NewAccessFromStrings("256.10.1.3", "")
		So(err, ShouldNotBeNil)
	})

	Convey("When NewAccessFromStrings is called with a domain name, it errors accordingly", t, func() {
		_, err := NewAccessFromStrings("a.domain.com", "")
		So(err, ShouldNotBeNil)
	})

	Convey("When NewAccessFromStrings is called with a proper IP but bad network mask, it errors accordingly", t, func() {
		_, err := NewAccessFromStrings("", "127.0.0.1/36")
		So(err, ShouldNotBeNil)
	})

	Convey("When NewAccessFromStrings is called with a domain name and mask (yes, really, why not?), it errors accordingly", t, func() {
		_, err := NewAccessFromStrings("", "hello.world.com/32")
		So(err, ShouldNotBeNil)
	})
}

// TestAccessValidateBadIP tests Access validating a bad IP address
func TestAccessValidateBadIP(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		Convey("and when asked to validate a known-bad IP, it refuses", func() {
			So(a.Validate("315.3.4.1"), ShouldBeFalse)
		})
	})
}

// TestAccessNoAddress tests Access when the address is empty
func TestAccessNoAddress(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("and when asked to validate an empty IP, it refuses", func() {
			So(a.Validate(""), ShouldBeFalse)
		})

	})
}

// TestAccessAllAll tests Access when "all" are explicitly allowed and denied
func TestAccessAllAll(t *testing.T) {

	Convey("When NewAccessFromStrings is called with \"all,all\" arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("all", "all")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("and when asked to validate an IP address, it allows it (Allow: all trumps Deny: all)", func() {
			So(a.Validate("123.123.123.123"), ShouldBeTrue)
		})
	})
}

// TestAccessNoneNone tests Access when "none" are explicitly allowed and denied
func TestAccessNoneNone(t *testing.T) {

	Convey("When NewAccessFromStrings is called with \"none,none\" arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("none", "none")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("and when asked to validate an IP address, it allows it (Deny: none trumps Allow: none)", func() {
			So(a.Validate("123.123.123.123"), ShouldBeTrue)
		})
	})

}

// TestAccessDenyAll tests Access when "all" are explicitly denied
func TestAccessDenyAll(t *testing.T) {

	Convey("When NewAccessFromStrings is called with \"deny all\" arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "all")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("and when asked to validate an IP address, it is denied", func() {
			So(a.Validate("123.123.123.123"), ShouldBeFalse)
		})
	})
}

// TestAccessAllowMixed tests Access against allowed IPs in a mixed setup
func TestAccessAllowMixed(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("when asked to validate a list of known-approved IP addresses, they are approved", func() {
			for _, address := range []string{"192.168.0.1", "127.0.0.1", "137.143.110.101"} {
				So(a.Validate(address), ShouldBeTrue)
			}
		})
	})
}

// TestAccessDenyMixed tests Access against denied IPs in a mixed setup
func TestAccessDenyMixed(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("when asked to validate a list of known-unapproved IP addresses, they are denied", func() {
			for _, address := range []string{"127.0.0.2", "137.143.110.102"} {
				So(a.Validate(address), ShouldBeFalse)
			}
		})
	})
}

// TestAccessAllowAllowOnly tests Access against allowed IPs in an Allow-only setup
func TestAccessAllowAllowOnly(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Allow-Only Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessAllow(a, t)
		So(a, ShouldNotBeNil)

		Convey("when asked to validate a list of known-approved IP addresses, they are approved", func() {
			for _, address := range []string{"192.168.0.1", "127.0.0.1", "137.143.110.101"} {
				So(a.Validate(address), ShouldBeTrue)
			}
		})
	})
}

// TestAccessDenyAllowOnly Access against denied IPs in an Allow-only setup
func TestAccessDenyAllowOnly(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Allow-Only Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessAllow(a, t)
		So(a, ShouldNotBeNil)

		Convey("when asked to validate a list of known-denied IP addresses, they are denied", func() {
			for _, address := range []string{"127.0.0.2", "137.143.110.102"} {
				So(a.Validate(address), ShouldBeFalse)
			}
		})
	})

}

// TestAccessAllowDenyOnly tests Access against allowed IPs in an Deny-only setup
func TestAccessAllowDenyOnly(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Deny-Only Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessDeny(a, t)
		So(a, ShouldNotBeNil)

		Convey("when asked to validate a list of known-approved IP addresses, they are approved", func() {
			for _, address := range []string{"192.168.0.1", "127.0.0.1"} {
				So(a.Validate(address), ShouldBeTrue)
			}
		})
	})
}

// TestAccessDenyDenyOnly tests Access against denied IPs in an Deny-only setup
func TestAccessDenyDenyOnly(t *testing.T) {

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Deny-Only Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessDeny(a, t)
		So(a, ShouldNotBeNil)

		Convey("when asked to validate a list of known-denied IP addresses, they are denied", func() {
			for _, address := range []string{"137.143.110.101", "127.0.0.2"} {
				So(a.Validate(address), ShouldBeFalse)
			}
		})
	})
}

// TestAccessHandlerAllow tests tests AccessHandler against allowed IPs
func TestAccessHandlerAllow(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("and a request is made from a known-ok IPv4 address, it is allowed", func() {
			req.RemoteAddr = "127.0.0.1"

			rr := httptest.NewRecorder()
			handler := a.AccessHandler(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
		})

		Convey("and a request is made from a known-ok IPv6 address, it is allowed", func() {
			req.RemoteAddr = "[::1]"

			rr := httptest.NewRecorder()
			handler := a.AccessHandler(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusOK)
		})
	})
}

// TestAccessHandlerDeny tests tests AccessHandler against denied IPs
func TestAccessHandlerDeny(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	Convey("When NewAccessFromStrings is called with empty arguments, it doesn't freak out. Setting up Mixed Access succeeds.", t, func() {
		a, err := NewAccessFromStrings("", "")
		So(err, ShouldBeNil)

		a = setupAccessMixed(a, t)
		So(a, ShouldNotBeNil)

		Convey("and a request is made from a known-denied IP address, it is denied", func() {
			req.RemoteAddr = "127.0.0.2"

			rr := httptest.NewRecorder()
			handler := a.AccessHandler(testHandler)

			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusForbidden)
		})
	})
}

func setupAccessAllow(a *Access, t *testing.T) *Access {

	for _, address := range []string{"192.168.0.1/24", "127.0.0.1", "137.143.110.101"} {
		err := a.AddAddress(address, true)
		So(err, ShouldBeNil)
	}

	return a
}

func setupAccessDeny(a *Access, t *testing.T) *Access {

	for _, address := range []string{"127.0.0.2", "137.143.0.0/16"} {
		err := a.AddAddress(address, false)
		So(err, ShouldBeNil)
	}

	return a
}

func setupAccessMixed(a *Access, t *testing.T) *Access {

	for _, address := range []string{"192.168.0.1/24", "127.0.0.1", "137.143.110.101", "::1"} {
		err := a.AddAddress(address, true)
		So(err, ShouldBeNil)
	}

	for _, address := range []string{"127.0.0.2", "137.143.0.0/16"} {
		err := a.AddAddress(address, false)
		So(err, ShouldBeNil)
	}

	return a
}
