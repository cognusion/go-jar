package utils

import (
	. "github.com/smartystreets/goconvey/convey"

	"bytes"
	"io/ioutil"
	"testing"
)

func Test_ReadAllSimple(t *testing.T) {
	hw := []byte("Hello World")

	Convey("When using ReadAll on an io.Reader with a known value, the bytes are consistent", t, func() {
		buff := bytes.NewBuffer(hw)
		val, err := ReadAll(buff)
		So(err, ShouldBeNil)
		So(val, ShouldResemble, hw)

	})
}

func BenchmarkIoUtilReadAll(b *testing.B) {
	var (
		val  []byte
		err  error
		hw   = []byte("Hello World")
		shw  = string(hw)
		buff = bytes.NewReader(hw)
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val, err = ioutil.ReadAll(buff)
		if err != nil {
			panic(err)
		} else if string(val) != shw {
			panic("WTF " + string(val))
		}
		buff.Seek(0, 0)
	}
}

func BenchmarkReadAll(b *testing.B) {
	var (
		val  []byte
		err  error
		hw   = []byte("Hello World")
		shw  = string(hw)
		buff = bytes.NewReader(hw)
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val, err = ReadAll(buff)
		if err != nil {
			panic(err)
		} else if string(val) != shw {
			panic("WTF " + string(val))
		}
		buff.Seek(0, 0)
	}
}
