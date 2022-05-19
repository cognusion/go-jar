// Package utils is a collection of stand-alone utility functions that are used across JAR
package utils

import (
	"bytes"
	"io"
	"sync"
)

var (
	buffPool sync.Pool
)

func init() {
	buffPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
}

// ReadAll is a custom version of io/ioutil.ReadAll() that uses a sync.Pool of bytes.Buffer to rock the reading,
// with Zero allocs and 7x better performance
func ReadAll(r io.Reader) (b []byte, err error) {
	/*
		From:
		BenchmarkBytes-8                	 3270517	       383 ns/op	    2048 B/op	       2 allocs/op
		BenchmarkWrite-8                	 3149658	       384 ns/op	    2048 B/op	       2 allocs/op

		To:
		BenchmarkBytes-8                	23543131	        50.4 ns/op	       0 B/op	       0 allocs/op
		BenchmarkWrite-8                	20136890	        59.2 ns/op	       0 B/op	       0 allocs/op
	*/
	buf := buffPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer buffPool.Put(buf)

	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()

	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
