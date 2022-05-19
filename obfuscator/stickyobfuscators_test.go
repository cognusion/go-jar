package obfuscator

import (
	. "github.com/smartystreets/goconvey/convey"

	"crypto/rand"
	"encoding/binary"
	"io"
	//"log"
	//"os"
	"testing"
	"time"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	//TimingOut = log.New(os.Stderr, "[TIMING] ", OutFormat)
	//ErrorOut = log.New(os.Stderr, "[ERROR] ", OutFormat)
}

/*
Benchmark_AesObfuscator128-8         	 1000000	      1093 ns/op	     256 B/op	       7 allocs/op
Benchmark_AesNormalizer128-8         	 5000000	       332 ns/op	     144 B/op	       4 allocs/op
Benchmark_AesObfuscator128Ttl-8      	 1000000	      1490 ns/op	     416 B/op	      10 allocs/op
Benchmark_AesNormalizer128Ttl-8      	 2000000	       639 ns/op	     272 B/op	       7 allocs/op
Benchmark_HexObfuscator-8            	20000000	        94.6 ns/op	      64 B/op	       2 allocs/op
Benchmark_HexNormalizer-8            	20000000	       100 ns/op	      48 B/op	       2 allocs/op
Benchmark_NonceRandom12-8            	 2000000	       648 ns/op	      16 B/op	       1 allocs/op
Benchmark_NonceRandom4-8             	 2000000	       640 ns/op	       4 B/op	       1 allocs/op
Benchmark_NonceTimeRandom4-8         	 2000000	       703 ns/op	       4 B/op	       1 allocs/op
*/

func TestAesObfuscator128(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key", t, func() {

		message := "This is a test"
		o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a message is obfuscated with it", func() {
			obfuscated := o.Obfuscate(message)
			So(obfuscated, ShouldNotBeEmpty)
			So(obfuscated, ShouldNotEqual, message)

			Convey("the message is recoverable", func() {

				clear := o.Normalize(obfuscated)
				So(clear, ShouldEqual, message)
			})

		})

	})
}

func TestAesObfuscatorBadKey(t *testing.T) {

	Convey("When an AesObfuscator is created with a bad 15byte key, if fails as expected", t, func() {

		o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z"))
		So(err.Error(), ShouldEqual, "crypto/aes: invalid key size 15")
		So(o, ShouldBeNil)

	})
}

func TestAesObfuscator128Ttl(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key, and a future TTL", t, func() {

		message := "This is a test"
		o, err := NewAesObfuscatorWithExpiration([]byte("95Bx9JkKX3xbd7z3"), 5*time.Second)
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a message is obfuscated with it", func() {
			obfuscated := o.Obfuscate(message)
			So(obfuscated, ShouldNotBeEmpty)
			So(obfuscated, ShouldNotEqual, message)

			Convey("the message is recoverable", func() {

				clear := o.Normalize(obfuscated)
				So(clear, ShouldEqual, message)
			})

		})

	})
}

func TestAesObfuscator128TtlFail(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key, and a future TTL", t, func() {

		message := "This is a test"
		o, err := NewAesObfuscatorWithExpiration([]byte("95Bx9JkKX3xbd7z3"), 1*time.Second)
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a message is obfuscated with it", func() {
			obfuscated := o.Obfuscate(message)
			So(obfuscated, ShouldNotBeEmpty)
			So(obfuscated, ShouldNotEqual, message)

			Convey("after sleeping past the TTL, the message is NOT recoverable", func() {
				time.Sleep(1100 * time.Millisecond)
				clear := o.Normalize(obfuscated)
				So(clear, ShouldBeEmpty)
			})

		})

	})
}

func TestAesObfuscator128TtlBadExpiration(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key, and a future TTL", t, func() {

		message := "This is a test"
		o, err := NewAesObfuscatorWithExpiration([]byte("95Bx9JkKX3xbd7z3"), 5*time.Second)
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		no, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
		So(err, ShouldBeNil)
		So(no, ShouldNotBeNil)

		Convey("and a message is obfuscated with the same key, but no TTL (contrived)", func() {
			obfuscated := no.Obfuscate(message)
			So(obfuscated, ShouldNotBeEmpty)
			So(obfuscated, ShouldNotEqual, message)

			Convey("the message is not recoverable", func() {

				clear := o.Normalize(obfuscated)
				So(clear, ShouldEqual, "")
			})

		})

	})
}

func TestAesObfuscator192(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 24byte key", t, func() {

		message := "This is a test"
		o, err := NewAesObfuscator([]byte("cf2nO99ZuWtc4lXsRNONCbp7"))
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a message is obfuscated with it", func() {
			obfuscated := o.Obfuscate(message)
			So(obfuscated, ShouldNotBeEmpty)
			So(obfuscated, ShouldNotEqual, message)

			Convey("the message is recoverable", func() {

				clear := o.Normalize(obfuscated)
				So(clear, ShouldEqual, message)
			})

		})

	})
}
func TestAesObfuscator256(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 32byte key", t, func() {

		message := "This is a test"
		o, err := NewAesObfuscator([]byte("fOFWV7E4fFuj6cvNPHYbCCD0C90dUnQx"))
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a message is obfuscated with it", func() {
			obfuscated := o.Obfuscate(message)
			So(obfuscated, ShouldNotBeEmpty)
			So(obfuscated, ShouldNotEqual, message)

			Convey("the message is recoverable", func() {

				clear := o.Normalize(obfuscated)
				So(clear, ShouldEqual, message)
			})

		})

	})
}

func TestAesObfuscatorNothingNormalized(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key", t, func() {

		message := ""
		o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and an empty message is Normalized with it, an empty string is returned", func() {
			obfuscated := o.Normalize(message)
			So(obfuscated, ShouldBeEmpty)
		})

	})
}

func TestAesObfuscatorGarbageNormalized(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key", t, func() {

		message := "sdflsdkjf4wSDfsdfksjd4RSDFFFv"
		o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a garbage message is Normalized with it, an empty string is returned", func() {
			obfuscated := o.Normalize(message)
			So(obfuscated, ShouldBeEmpty)
		})

	})
}

func TestAesObfuscatorBase64dGarbageNormalized(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key", t, func() {

		message := "aGVsbG8K"
		o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a garbage message is Normalized with it, an empty string is returned", func() {
			obfuscated := o.Normalize(message)
			So(obfuscated, ShouldBeEmpty)
		})

	})
}

func TestAesObfuscatorFixedNonceNormalized(t *testing.T) {

	Convey("When an AesObfuscator is created with a known 16byte key", t, func() {

		message := "ylFZ5v1JgjWWAJMDXPLLkRiwI3ielRPBSef55he-CVaHV3NYJVgexA"
		o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
		So(err, ShouldBeNil)
		So(o, ShouldNotBeNil)

		Convey("and a previously fixed-nonce message is Normalized with it, an empty string is returned", func() {
			obfuscated := o.Normalize(message)
			So(obfuscated, ShouldBeEmpty)
		})

	})
}

func TestHexObfuscator(t *testing.T) {

	Convey("When a HexObfuscator is created", t, func() {

		message := "This is a test"
		o := HexObfuscator{}
		So(o, ShouldNotBeNil)

		Convey("and a message is obfuscated with it", func() {
			obfuscated := o.Obfuscate(message)
			So(obfuscated, ShouldNotBeEmpty)
			So(obfuscated, ShouldNotEqual, message)

			Convey("the message is recoverable", func() {

				clear := o.Normalize(obfuscated)
				So(clear, ShouldEqual, message)
			})

		})

	})
}

func Test_NonceTimeRandom12Uniqish(t *testing.T) {

	t.Skip("Silly long test that barely proves anything")

	list := make(map[string]bool)
	for i := 0; i < 10000000; i++ {
		nonce := make([]byte, 12)
		binary.PutVarint(nonce, time.Now().UnixNano())
		rpend := make([]byte, 4)
		if _, err := io.ReadFull(rand.Reader, rpend); err != nil {
			panic(err.Error())
		}
		for i := 0; i < 4; i++ {
			nonce[i+8] = rpend[i]
		}
		if _, ok := list[string(nonce)]; ok {
			t.Fail()
		}
		list[string(nonce)] = true
	}
}

func Benchmark_AesObfuscator128(b *testing.B) {

	o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
	if err != nil {
		b.Fatalf("Creating new AesObfuscator failed: %s\n", err)
	}

	s := "This is a test"
	b.ResetTimer()
	var os string
	for i := 0; i < b.N; i++ {
		os = o.Obfuscate(s)
		if os == "" {
			b.Fail()
		}
	}
}

func Benchmark_AesNormalizer128(b *testing.B) {

	o, err := NewAesObfuscator([]byte("95Bx9JkKX3xbd7z3"))
	if err != nil {
		b.Fatalf("Creating new AesObfuscator failed: %s\n", err)
	}

	s := "C7Gr2cONX6h7o8sZzMkHnVHPnLLBxa_gR5GxcV47zpru2rb0qv5B0REz"
	var ns string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns = o.Normalize(s)
		if ns == "" {
			b.Fail()
		}
	}
}

func Benchmark_AesObfuscator128Ttl(b *testing.B) {

	o, err := NewAesObfuscatorWithExpiration([]byte("95Bx9JkKX3xbd7z3"), 5*time.Second)
	if err != nil {
		b.Fatalf("Creating new AesObfuscator failed: %s\n", err)
	}

	s := "This is a test"
	var os string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os = o.Obfuscate(s)
		if os == "" {
			b.Fail()
		}
	}
}

func Benchmark_AesNormalizer128Ttl(b *testing.B) {

	o, err := NewAesObfuscatorWithExpiration([]byte("95Bx9JkKX3xbd7z3"), 5*time.Second)
	if err != nil {
		b.Fatalf("Creating new AesObfuscator failed: %s\n", err)
	}

	s := "This is a test"
	os := o.Obfuscate(s)
	var ns string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns = o.Normalize(os)
		if ns == "" {
			b.Fail()
		}
	}
}

func Benchmark_HexObfuscator(b *testing.B) {

	o := HexObfuscator{}

	s := "This is a test"
	var os string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os = o.Obfuscate(s)
		if os == "" {
			b.Fail()
		}
	}
}

func Benchmark_HexNormalizer(b *testing.B) {

	o := HexObfuscator{}

	s := "5468697320697320612074657374"
	var ns string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns = o.Normalize(s)
		if ns == "" {
			b.Fail()
		}
	}
}

func Benchmark_NonceRandom12(b *testing.B) {

	for i := 0; i < b.N; i++ {
		nonce := make([]byte, 12)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			panic(err.Error())
		}
	}
}

func Benchmark_NonceRandom4(b *testing.B) {

	for i := 0; i < b.N; i++ {
		nonce := make([]byte, 4)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			panic(err.Error())
		}
	}
}

func Benchmark_NonceTimeRandom4(b *testing.B) {

	for i := 0; i < b.N; i++ {
		nonce := make([]byte, 12)
		binary.PutVarint(nonce, time.Now().UnixNano())
		rpend := make([]byte, 4)
		if _, err := io.ReadFull(rand.Reader, rpend); err != nil {
			panic(err.Error())
		}
		for i := 0; i < 4; i++ {
			nonce[i+8] = rpend[i]
		}
	}
}
