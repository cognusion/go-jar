package jar

import (
//. "github.com/smartystreets/goconvey/convey"

//"testing"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	//TimingOut = log.New(os.Stderr, "[TIMING] ", OutFormat)
	//ErrorOut = log.New(os.Stderr, "[ERROR] ", OutFormat)
}

/* Commenting out LDAP-specific tests. May revisit.
func TestNewLdapSourceFromUrlGoodTests(t *testing.T) {

	Convey("When an ldap-prefixed URL is passed, without a port, everything is ok", t, func() {
		url := "ldap://example.domain.com/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{"example.domain.com", 389, false, "blahblah"})
		So(e, ShouldBeNil)
	})

	Convey("When an ldaps-prefixed URL is passed, without a port, everything is ok", t, func() {
		url := "ldaps://example.domain.com/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{"example.domain.com", 389, true, "blahblah"})
		So(e, ShouldBeNil)
	})

	Convey("When an ldap-prefixed URL is passed, with a port, everything is ok", t, func() {
		url := "ldap://example.domain.com:6612/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{"example.domain.com", 6612, false, "blahblah"})
		So(e, ShouldBeNil)
	})
}

func TestNewLdapSourceFromUrlBadPrefix(t *testing.T) {

	Convey("When a non-ldap-prefixed URL is passed, an error is returned", t, func() {
		url := "https://example.domain.com"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "prefix must be")
	})
}

func TestNewLdapSourceFromUrlNoQuery(t *testing.T) {

	Convey("When an ldap-prefixed URL is passed, without a query, an error is returned", t, func() {
		url := "ldap://example.domain.com"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "must contain at least host/query")
	})
}

func TestNewLdapSourceFromUrlPortNoPort(t *testing.T) {

	Convey("When an ldap-prefixed URL is passed, with a colon but no port, an error is returned", t, func() {
		url := "ldap://example.domain.com:/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "must contain at least host:port")
	})

	Convey("When an ldap-prefixed URL is passed, with a colon and port but no hostname, an error is returned", t, func() {
		url := "ldap://:17/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "must contain at least host:port")
	})
}

func TestNewLdapSourceFromUrlPortErrors(t *testing.T) {

	Convey("When an ldap-prefixed URL is passed, with a colon but non-int port, an error is returned", t, func() {
		url := "ldap://example.domain.com:nan/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "invalid syntax") // from ParseInt
	})

	Convey("When an ldap-prefixed URL is passed, with a colon but way-too-large port, an error is returned", t, func() {
		url := "ldap://example.domain.com:160000000000000000000000000000000000/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "value out of range") // from ParseInt
	})

	Convey("When an ldap-prefixed URL is passed, with a colon but too small port, an error is returned", t, func() {
		url := "ldap://example.domain.com:0/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "must be between 1 and 65535")
	})

	Convey("When an ldap-prefixed URL is passed, with a colon but too large port, an error is returned", t, func() {
		url := "ldap://example.domain.com:65536/blahblah"

		l, e := NewLdapSourceFromUrl(url)
		So(l, ShouldResemble, LdapSource{})
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldContainSubstring, "must be between 1 and 65535")
	})
}
*/
