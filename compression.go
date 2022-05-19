package jar

import (
	"github.com/NYTimes/gziphandler"

	"net/http"
)

func init() {
	ConfigAdditions[ConfigCompression] = make([]string, 0)
}

// Compression is used to support GZIP compression of data en route to a client
type Compression struct {
	contentTypes []string
}

// NewCompression returns a pointer to a Compression struct with the specified MIME-types baked in
func NewCompression(contentTypes []string) *Compression {
	return &Compression{contentTypes}
}

// Handler is a middleware to potentially GZIP-compress outgoing response bodies
func (c *Compression) Handler(next http.Handler) http.Handler {

	handler, err := gziphandler.GzipHandlerWithOpts(gziphandler.ContentTypes(c.contentTypes))
	if err != nil {
		// This is a rare ErrorOut without a requestid, because we don't have the
		// request object at this point
		ErrorOut.Printf("Error creating GzipHandler, skipping: %s\n", err)
		return next
	}

	return handler(next)
}
