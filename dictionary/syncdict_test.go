package dictionary

import (
	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

func TestSyncDict(t *testing.T) {

	var s SyncDict
	s.AddValues(map[string]string{
		"NAME":         "Tester",
		"VERSION":      "1.0",
		"SHORTVERSION": "%%NAME/%%VERSION",
	})

	Convey("When a SimpleDict is created, strings containing dictionary items are properly replaced", t, func() {
		So(s.Replacer("My name is %%NAME"), ShouldEqual, "My name is Tester")

		Convey("and dictionary items containing other dictionary items are properly expanded", func() {
			exp := s.Replacer("%%NAME %%VERSION %%SHORTVERSION")
			So(exp, ShouldEqual, "Tester 1.0 Tester/1.0")
		})
	})

}

func TestSyncDictResolve(t *testing.T) {

	var s SyncDict
	s.AddValues(map[string]string{
		"NAME":         "Tester",
		"VERSION":      "1.0",
		"SHORTVERSION": "%%NAME/%%VERSION",
		"FULLVERSION":  "%%SHORTVERSION Moar Words %%NAME",
	})
	s.Resolve()

	Convey("When a SimpleDict is created and then Resolve(), strings containing dictionary items are properly replaced", t, func() {
		So(s.Replacer("My name is %%NAME"), ShouldEqual, "My name is Tester")

		Convey("and dictionary items containing other dictionary items are properly expanded", func() {
			exp := s.Replacer("%%NAME %%VERSION %%SHORTVERSION")
			So(exp, ShouldEqual, "Tester 1.0 Tester/1.0")
		})

		Convey("and strings containing two levels of dictionary items are properly replaced", func() {
			So(s.Replacer("%%FULLVERSION"), ShouldEqual, "Tester/1.0 Moar Words Tester")
		})

		Convey("and strings with dictionary-item-looking substrings do not hang up the processing", func() {
			So(s.Replacer("This %%FULLVERSION is %%NOTHING"), ShouldEqual, "This Tester/1.0 Moar Words Tester is %%NOTHING")
		})
	})

}

func TestSyncDict2ndLevelRecursion(t *testing.T) {

	var s SyncDict
	s.AddValues(map[string]string{
		"NAME":         "Tester",
		"VERSION":      "1.0",
		"SHORTVERSION": "%%NAME/%%VERSION",
		"FULLVERSION":  "%%SHORTVERSION Moar Words %%NAME",
	})

	Convey("When a SimpleDict is created, strings containing two levels of dictionary items are properly replaced", t, func() {
		So(s.Replacer("%%FULLVERSION"), ShouldEqual, "Tester/1.0 Moar Words Tester")

		Convey("and strings with dictionary-item-looking substrings do not hang up the processing", func() {
			So(s.Replacer("This %%FULLVERSION is %%NOTHING"), ShouldEqual, "This Tester/1.0 Moar Words Tester is %%NOTHING")
		})
	})

}

func TestSyncDict20thLevelRecursion(t *testing.T) {

	var s SyncDict
	s.AddValues(map[string]string{
		"A": "A",
		"B": "%%A",
		"C": "%%B",
		"D": "%%C",
		"E": "%%D",
		"F": "%%E",
		"G": "%%F",
		"H": "%%G",
		"I": "%%H",
		"J": "%%I",
		"K": "%%J",
		"L": "%%K",
		"M": "%%L",
		"N": "%%M",
		"O": "%%N",
		"P": "%%O",
		"Q": "%%P",
		"R": "%%Q",
		"S": "%%R",
		"T": "%%S", // 20th
		"U": "%%T",
		"V": "%%U",
		"W": "%%V",
		"X": "%%W",
		"Y": "%%X",
		"Z": "%%Y",
	})

	Convey("When a SimpleDict is created, strings containing two levels of dictionary items are properly replaced", t, func() {
		So(s.Replacer("%%T"), ShouldEqual, "A")
	})
	// It's worth calling out here, that when testing %%Z, it may succeed, or fail, even with the 20-level recursion circuit breaker
	// because of random ordering within the string map that may allow same-pass replacements of multiple values

}

func Benchmark_SyncDictNull(b *testing.B) {
	var s SyncDict

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Replacer("%%Z %%Z %%Z %%Z")
	}
}

func Benchmark_SyncDictSimple(b *testing.B) {
	var s SyncDict
	s.AddValues(map[string]string{
		"A": "A",
		"B": "%%A",
		"C": "%%B",
		"D": "%%C",
		"E": "%%D",
		"F": "%%E",
		"G": "%%F",
		"H": "%%G",
		"I": "%%H",
		"J": "%%I",
		"K": "%%J",
		"L": "%%K",
		"M": "%%L",
		"N": "%%M",
		"O": "%%N",
		"P": "%%O",
		"Q": "%%P",
		"R": "%%Q",
		"S": "%%R",
		"T": "%%S", // 20th
		"U": "%%T",
		"V": "%%U",
		"W": "%%V",
		"X": "%%W",
		"Y": "%%X",
		"Z": "%%Y",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Replacer("%%A")
	}
}

func Benchmark_SyncDictAwful(b *testing.B) {
	var s SyncDict
	s.AddValues(map[string]string{
		"A": "A",
		"B": "%%A",
		"C": "%%B",
		"D": "%%C",
		"E": "%%D",
		"F": "%%E",
		"G": "%%F",
		"H": "%%G",
		"I": "%%H",
		"J": "%%I",
		"K": "%%J",
		"L": "%%K",
		"M": "%%L",
		"N": "%%M",
		"O": "%%N",
		"P": "%%O",
		"Q": "%%P",
		"R": "%%Q",
		"S": "%%R",
		"T": "%%S", // 20th
		"U": "%%T",
		"V": "%%U",
		"W": "%%V",
		"X": "%%W",
		"Y": "%%X",
		"Z": "%%Y",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Replacer("%%Z %%Z %%Z %%Z")
	}
}

func Benchmark_SyncDictAwfulResolve(b *testing.B) {
	var s SyncDict
	s.AddValues(map[string]string{
		"A": "A",
		"B": "%%A",
		"C": "%%B",
		"D": "%%C",
		"E": "%%D",
		"F": "%%E",
		"G": "%%F",
		"H": "%%G",
		"I": "%%H",
		"J": "%%I",
		"K": "%%J",
		"L": "%%K",
		"M": "%%L",
		"N": "%%M",
		"O": "%%N",
		"P": "%%O",
		"Q": "%%P",
		"R": "%%Q",
		"S": "%%R",
		"T": "%%S", // 20th
		"U": "%%T",
		"V": "%%U",
		"W": "%%V",
		"X": "%%W",
		"Y": "%%X",
		"Z": "%%Y",
	})

	s.Resolve()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Replacer("%%Z %%Z %%Z %%Z")
	}
}
