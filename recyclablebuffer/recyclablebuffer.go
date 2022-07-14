// Package recyclablebuffer provides the RecyclableBuffer, a multiuse buffer that very reusable,
// supports re-reading the contained buffer, and when Close()d, will return home to its sync.Pool
package recyclablebuffer

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

// ErrTooLarge is returned when ResetFromLimitedReader is used and the supplied Reader writes too much
var ErrTooLarge = errors.New("read byte count too large")

// RecyclableBuffer is an io.Reader, io.ReadCloser, io.Writer, and more that comes best from a sync.Pool.
// Its Close method puts itself back in the Pool it came from
type RecyclableBuffer struct {
	bytes.Reader

	home *sync.Pool
}

// NewRecyclableBuffer returns a RecyclableBuffer with a proper home
func NewRecyclableBuffer(home *sync.Pool, bytes []byte) *RecyclableBuffer {
	return &RecyclableBuffer{
		home: home,
	}
}

// Close puts itself back in the Pool it came from. This should absolutely **never** be
// called more than once per RecyclableBuffer life.
// Implements “io.Closer“ (also “io.ReadCloser“ and “io.WriteCloser“)
func (r *RecyclableBuffer) Close() error {
	r.home.Put(r)
	return nil
}

// ResetFromReader performs a Reset() using the contents of the supplied Reader as the new content
func (r *RecyclableBuffer) ResetFromReader(reader io.Reader) {
	b, _ := io.ReadAll(reader)
	r.Reset(b)
}

// ResetFromLimitedReader performs a Reset() using the contents of the supplied Reader as the new content,
// up to at most max bytes, returning ErrTooLarge if it's over. The error is not terminal, and the buffer
// may continue to be used, understanding the contents will be limited
func (r *RecyclableBuffer) ResetFromLimitedReader(reader io.Reader, max int64) error {
	lr := io.LimitReader(reader, max+1)
	b, _ := io.ReadAll(lr)
	if int64(len(b)) > max {
		r.Reset(b[0:max])
		return ErrTooLarge
	}
	r.Reset(b)
	return nil
}

// Bytes returns the contents of the buffer, and sets the seek pointer back to the beginning
func (r *RecyclableBuffer) Bytes() []byte {
	b, _ := io.ReadAll(&r.Reader)
	r.Seek(0, 0) // reset the seeker
	return b
}

// String returns the contents of the buffer as a string, and sets the seek pointer back to the beginning
func (r *RecyclableBuffer) String() string {
	b, _ := io.ReadAll(&r.Reader)
	r.Seek(0, 0) // reset the seeker
	return string(b)
}

// Error returns the contents of the buffer as a string. Implements “error“
func (r *RecyclableBuffer) Error() string {
	return r.String()
}

// Writer adds the bytes the written to the buffer. Implements “io.Writer“
func (r *RecyclableBuffer) Write(p []byte) (n int, err error) {
	b, err := io.ReadAll(&r.Reader)
	if err != nil {
		return 0, err
	}
	b = append(b, p...)

	r.Reset(b)
	return len(p), nil
}
