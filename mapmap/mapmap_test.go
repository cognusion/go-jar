package mapmap

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_NewMapMap(t *testing.T) {

	routefiles := map[string]string{
		"endpoints": "tests/endpoints.map",
		"ids":       "tests/ids.map",
	}

	mapMap := NewMapMap()

	Convey("When given a list of map files to load, they load without error", t, func() {
		for n, f := range routefiles {
			err := mapMap.Load(n, f)
			So(err, ShouldBeNil)
		}

		Convey("... the MapMap is the correct size", func() {
			So(mapMap.Size(), ShouldEqual, 2)
		})

		Convey("... values are correct", func() {
			So(mapMap.Get("ids", "goodone"), ShouldEqual, "d5a89f10b4f211eca52e54e1ad26168b")
			So(mapMap.Get("ids", "goodtwo"), ShouldEqual, "e10b2f30b4f211ec85b954e1ad26168b")
			So(mapMap.Get("ids", "gthree"), ShouldEqual, "f60d54bcb4f211ec97df54e1ad26168b")

			So(mapMap.Get("endpoints", "goodone"), ShouldEqual, "loc1")
			So(mapMap.Get("endpoints", "goodtwo"), ShouldEqual, "loc2")
			So(mapMap.Get("endpoints", "gthree"), ShouldEqual, "loc1")

			So(mapMap.GetURLRoute("goodone"), ShouldResemble, &URLRoute{"goodone", "d5a89f10b4f211eca52e54e1ad26168b", "loc1"})
			So(mapMap.GetURLRoute("goodtwo"), ShouldResemble, &URLRoute{"goodtwo", "e10b2f30b4f211ec85b954e1ad26168b", "loc2"})
			So(mapMap.GetURLRoute("gthree"), ShouldResemble, &URLRoute{"gthree", "f60d54bcb4f211ec97df54e1ad26168b", "loc1"})
		})

		Convey("... bad values aren't drama", func() {
			So(mapMap.Get("crapmap", "garbage"), ShouldBeBlank) // test bad map and key
			So(mapMap.Get("ids", "garbage"), ShouldBeBlank)     // test good map, bad key
			So(mapMap.GetURLRoute("garbage"), ShouldResemble, &URLRoute{"garbage", "", ""})
		})
	})
}
