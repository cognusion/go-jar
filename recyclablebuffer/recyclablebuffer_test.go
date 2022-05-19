package recyclablebuffer

import (
	. "github.com/smartystreets/goconvey/convey"

	"bytes"
	"io"
	"sync"
	"testing"
)

// HOWTO implement a goro-safe sync.Pool for RecyclableBuffers
func Example() {

	// sync.Pool allows us to have a never-ending font of RecyclableBuffers.
	// If the Pool is empty, a new one is created. If there is one someone put
	// back, then it is returned. Saves on allocs like crazy. <3
	var rPool sync.Pool
	rPool = sync.Pool{
		New: func() interface{} {
			// New RecyclableBuffers will be homed to this Pool,
			// automagically whenever there is a Close()
			return NewRecyclableBuffer(&rPool, make([]byte, 0))
		},
	}

	// Let's grab a RecyclableBuffer
	rb := rPool.Get().(*RecyclableBuffer)

	// And immediately reset the value, as we can't trust it to be empty
	rb.Reset([]byte("Hello World"))

	// Unlike most buffers, we can re-read it:
	for i := 0; i < 10; i++ {
		if string(rb.Bytes()) != "Hello World" {
			panic("OMG! Can't reread?!!!")
		}
	}

	// Or get the string value, if you prefer (and know it's safe)
	for i := 0; i < 10; i++ {
		if rb.String() != "Hello World" {
			panic("OMG! Can't reread?!!!")
		}
	}

	// Appending to it as an io.Writer works as well
	io.WriteString(rb, ", nice day?")
	if string(rb.Bytes()) != "Hello World, nice day?" {
		panic("OMG! Append failed?!")
	}

	// Lastly, when you're all done, just close it.
	rb.Close() // and it will go back into the Pool.
	// Please don't use it anymore. Get a fresh one.

	rb = rPool.Get().(*RecyclableBuffer) // See, not hard?
	defer rb.Close()                     // Just remember to close it, unless you're passing it elsewhere

	/* HINTS:
	* Makes awesome ``http.Request.Body``s, especially since they get automatically ``.Close()``d when done with
	* Replaces ``bytes.Buffer`` and ``bytes.Reader`` for most uses
	* Isa Stringer and an error
	* As a Writer and a Reader can be used in pipes and elsewhere
	  * You can also pipe them to themselves, but that is a very bad idea unless you love watching OOMs
	*/

}

func TestBytePool(t *testing.T) {
	var bytePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0)
		},
	}

	Convey("When an element from BytePool is fetched, it is a []byte", t, func() {
		b := bytePool.Get()
		So(b, ShouldHaveSameTypeAs, make([]byte, 0))

		Convey("And setting it does not panic, and appears correct", func() {
			c := func() {
				b = []byte("Hello World")
			}
			So(c, ShouldNotPanic)
			So(b, ShouldResemble, []byte("Hello World"))

			bytePool.Put(b)
		})
	})

}

func TestRecyclableBuffer(t *testing.T) {

	var rPool sync.Pool

	rPool = sync.Pool{
		New: func() interface{} {
			return NewRecyclableBuffer(&rPool, make([]byte, 0))
		},
	}

	Convey("When a RecyclableBuffer is fetched from a RecyclableBufferPool, it is a RecyclableBuffer", t, func() {
		rbx := rPool.Get()
		So(rbx, ShouldHaveSameTypeAs, &RecyclableBuffer{})

		rb := rPool.Get().(*RecyclableBuffer)
		Convey("And setting it appears correct", func() {
			rb.Reset([]byte("Hello World"))
			So(rb.Bytes(), ShouldResemble, []byte("Hello World"))
			So(rb.Error(), ShouldEqual, "Hello World")
			So(rb.String(), ShouldEqual, "Hello World")

			Convey("And re-reading from it multiple times works too", func() {
				for i := 0; i < 10; i++ {
					So(rb.Bytes(), ShouldResemble, []byte("Hello World"))
				}
			})

			Convey("Appending to it as an io.Writer works as well", func() {
				n, err := io.WriteString(rb, ", nice day?")
				So(n, ShouldBeGreaterThan, 0)
				So(err, ShouldBeNil)
				So(rb.Bytes(), ShouldResemble, []byte("Hello World, nice day?"))
			})
		})

		Convey("Resetting it using an io.Reader works as expected", func() {
			buff := bytes.NewBufferString("Hola Mundo")
			rb.ResetFromReader(buff)
			So(rb.Bytes(), ShouldResemble, []byte("Hola Mundo"))

			Convey("... and checking the string value is similarly correct", func() {
				So(rb.String(), ShouldEqual, "Hola Mundo")
			})
		})

		Convey("Resetting it, but limited, using an io.Reader works as expected", func() {
			buff := bytes.NewBufferString("Oh My Gosh")
			err := rb.ResetFromLimitedReader(buff, 20)
			So(err, ShouldBeNil)
			So(rb.Bytes(), ShouldResemble, []byte("Oh My Gosh"))

			Convey("... and when it's over the limit, that is handled as expected", func() {
				buff2 := bytes.NewBufferString("This is a long sentence")
				err := rb.ResetFromLimitedReader(buff2, 4)
				So(err, ShouldEqual, ErrTooLarge)
				So(rb.Bytes(), ShouldResemble, []byte("This"))
			})
		})

		Convey("Putting it back in the pool doesn't freak out", func() {
			So(func() { rb.Close() }, ShouldNotPanic)

			Convey("... doing it twice doesn't either (but don't ever do that, ever)...", func() {
				So(func() { rb.Close() }, ShouldNotPanic)
			})
		})
	})

}

// Tests grabbing an new RB, using the NewFunc directly
func BenchmarkRBNewRaw(b *testing.B) {

	var RecyclableBufferPool sync.Pool
	RecyclableBufferPool = sync.Pool{
		New: func() interface{} {
			return NewRecyclableBuffer(&RecyclableBufferPool, make([]byte, 0))
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb := NewRecyclableBuffer(&RecyclableBufferPool, make([]byte, 0))
		//rb.Close()
		rb.Len()
	}
}

// Tests grabbing an new RB, using Pool.Get, and a pre-seeded BytePool to feed from in the NewFunc
func BenchmarkRBNewGet(b *testing.B) {

	var RecyclableBufferPool sync.Pool
	RecyclableBufferPool = sync.Pool{
		New: func() interface{} {
			return NewRecyclableBuffer(&RecyclableBufferPool, make([]byte, 0))
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb := RecyclableBufferPool.Get().(*RecyclableBuffer)
		rb.Close()
	}
}

// Tests grabbing an new RB, using Pool.Get, but make([]byte) in the NewFunc
func BenchmarkRBNewMake(b *testing.B) {

	var RecyclableBufferPool sync.Pool
	RecyclableBufferPool = sync.Pool{
		New: func() interface{} {
			return NewRecyclableBuffer(&RecyclableBufferPool, make([]byte, 0))
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb := RecyclableBufferPool.Get().(*RecyclableBuffer)
		//rb.Close()
		rb.Len()
	}
}

// Test grabbing an new RB, using Pool.Get, and Reseting using a premade fixed []byte
func BenchmarkRBNewGetResetFixed(b *testing.B) {
	var empty = make([]byte, 0)

	var RecyclableBufferPool sync.Pool
	RecyclableBufferPool = sync.Pool{
		New: func() interface{} {
			return NewRecyclableBuffer(&RecyclableBufferPool, make([]byte, 0))
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb := RecyclableBufferPool.Get().(*RecyclableBuffer)
		rb.Reset(empty)
		rb.Close()
	}
}

// Tests grabbing an new RB, using Pool.Get, and Reseting using make([]byte) every time
func BenchmarkRBNewGetResetMake(b *testing.B) {
	var RecyclableBufferPool sync.Pool
	RecyclableBufferPool = sync.Pool{
		New: func() interface{} {
			return NewRecyclableBuffer(&RecyclableBufferPool, make([]byte, 0))
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb := RecyclableBufferPool.Get().(*RecyclableBuffer)
		rb.Reset(make([]byte, 0))
		rb.Close()
	}
}

func BenchmarkBytes(b *testing.B) {
	r := NewRecyclableBuffer(nil, make([]byte, 0))
	r.Reset([]byte("Hello World"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Bytes()
	}
}

func BenchmarkReset(b *testing.B) {
	r := NewRecyclableBuffer(nil, make([]byte, 0))
	r.Reset([]byte("Hello World"))

	var ok = []byte("ok")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Reset(ok)
	}
}

func BenchmarkWrite(b *testing.B) {
	r := NewRecyclableBuffer(nil, make([]byte, 0))
	r.Reset([]byte("Hello World"))

	var ok = []byte("ok")
	var hello = []byte("Hello World")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Write(ok)
		r.Reset(hello)
	}
}
