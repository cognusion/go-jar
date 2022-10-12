package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

func TestBootstrapConfigMin(t *testing.T) {

	config := "tests/conf/min.yaml"

	Convey("When a minimal config is loaded, it loads cleanly", t, func() {
		Conf = InitConfig()
		err := LoadConfig(config, Conf)
		So(err, ShouldBeNil)

		Convey("and bootstrap() cleanly creates a server", func() {
			done, servers := bootstrap()
			So(done, ShouldBeFalse)
			So(len(servers), ShouldEqual, 1)
		})
	})
}

func TestBootstrapConfigMinTLS(t *testing.T) {

	config := "tests/conf/mintls.yaml"

	Convey("When a minimal config with TLS is loaded, it loads cleanly", t, func() {
		Conf = InitConfig()
		err := LoadConfig(config, Conf)
		So(err, ShouldBeNil)

		Convey("and bootstrap() cleanly creates two servers", func() {
			done, servers := bootstrap()
			So(done, ShouldBeFalse)
			So(len(servers), ShouldEqual, 2)
		})
	})
}

func TestBootstrapConfigMinTLSBrokenCert(t *testing.T) {

	config := "tests/conf/mintlsbroken.yaml"

	Convey("When a minimal config with TLS is loaded, it loads cleanly", t, func() {
		Conf = InitConfig()
		err := LoadConfig(config, Conf)
		So(err, ShouldBeNil)

		Convey("and bootstrap() should panic because of a missing certificate ", func() {
			So(func() { bootstrap() }, ShouldPanic)
		})
	})
}

func TestBootstrapConfigVersionMiss(t *testing.T) {

	config := "tests/conf/versionmiss.yaml"

	Convey("When a minimal config with an impossible version requirement is loaded, it loads cleanly", t, func() {
		Conf = InitConfig()
		err := LoadConfig(config, Conf)
		So(err, ShouldBeNil)

		Convey("and bootstrap() should panic because of the version nonsense", func() {
			So(func() { bootstrap() }, ShouldPanic)
		})
	})
}

func TestBootstrapConfigVersionOk(t *testing.T) {

	config := "tests/conf/versionok.yaml"

	Convey("When a minimal config with a sane version requirement is loaded, it loads cleanly", t, func() {
		Conf = InitConfig()
		err := LoadConfig(config, Conf)
		So(err, ShouldBeNil)

		Convey("and bootstrap() should not panic", func() {
			So(func() { bootstrap() }, ShouldNotPanic)
		})
	})
}

func TestBootstrapConfigVersionNone(t *testing.T) {

	config := "tests/conf/versionnone.yaml"

	Convey("When a minimal config with no version requirement is loaded, it loads cleanly", t, func() {
		Conf = InitConfig()
		err := LoadConfig(config, Conf)
		So(err, ShouldBeNil)

		Convey("and bootstrap() should not panic", func() {
			So(func() { bootstrap() }, ShouldNotPanic)
		})
	})
}
