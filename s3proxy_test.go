package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

func TestIsBadMaybe(t *testing.T) {
	badexts := []string{".scr", ".exe", ".bat", ".msi", ".dll"}
	ok := "helloworld.oke"
	nok := "helloworld.exe"

	Convey("When a file with an unlisted extension is submitted, it succeeds (false)", t, func() {
		So(isBadFileMaybe(ok, badexts), ShouldBeFalse)
	})

	Convey("When a file with a listed extension is submitted, it fails (true)", t, func() {
		So(isBadFileMaybe(nok, badexts), ShouldBeTrue)
	})
}

func TestSanitizeFilename(t *testing.T) {
	ok := "thisis_afile.zip"
	nok := "this!is_a&file.zip"

	Convey("When a filename has no bad characters, it is returned unchanged", t, func() {
		f := sanitizeFilename(ok)
		So(f, ShouldEqual, ok)
	})

	Convey("When a filename has bad characters, it is returned properly", t, func() {
		f := sanitizeFilename(nok)
		So(f, ShouldEqual, ok)
	})

}
