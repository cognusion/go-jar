package jar

import (
	"github.com/davecgh/go-spew/spew"

	"github.com/cognusion/go-jar/recyclablebuffer"

	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	authoritativeDomains = make([]string, 0)

	// CheckAuthoritative compares domain suffixes in the "authoritativedomains" against the requested URL.Hostname()
	// and returns true if it matches or if "authoritativedomains" is not used
	CheckAuthoritative = func(*http.Request) bool { return true }

	// RecyclableBufferPool is a sync.Pool of RecyclableBuffers that are safe to Get() and use (after a reset), and then
	// Close() them when you're done, to put them back in the Pool
	RecyclableBufferPool sync.Pool
)

func init() {
	InitFuncs.Add(func() {
		if ds := Conf.GetStringSlice(ConfigAuthoritativeDomains); len(ds) > 0 {
			authoritativeDomains = ds
			CheckAuthoritative = checkAuthoritative
		}
	})

	RecyclableBufferPool = sync.Pool{
		New: func() interface{} {
			return recyclablebuffer.NewRecyclableBuffer(&RecyclableBufferPool, make([]byte, 0))
		},
	}
}

// checkAuthoritative compares domain suffixes in the ConfigAuthoritativeDomains against the requested URL.Hostname()
// and returns true if it matches or if ConfigAuthoritativeDomains is not used
func checkAuthoritative(r *http.Request) bool {

	comphost := r.URL.Hostname()
	DebugOut.Printf("CheckAuthoritative '%s'\n", comphost)
	if comphost == "" {
		comphost = r.Host
		if strings.Contains(comphost, ":") {
			comphost = strings.Split(comphost, ":")[0]
		}
		DebugOut.Printf("\tAdjusted '%s'\n", comphost)
	}

	comphost = strings.ToLower(comphost) // case-insensitive

	for _, ad := range authoritativeDomains {
		if strings.HasSuffix(comphost, ad) {
			// the hostname has the suffix
			return true
		}
		if strings.TrimPrefix(ad, ".") == comphost {
			// the hostname is the domain name
			return true
		}
	}
	return false
}

// StringIfCtx will return a non-empty string if the suppled Request
// has a Context.WithValue() of the specified name
func StringIfCtx(r *http.Request, name interface{}) string {
	if s := r.Context().Value(name); s != nil {
		return s.(string)
	}
	return ""
}

// FileExists returns true if the provided path exists, and is not a directory
func FileExists(filePath string) bool {
	if inf, err := os.Stat(filePath); err == nil {
		return !inf.IsDir()
	}
	return false
}

// FolderExists returns true if the provided path exists, and is a directory
func FolderExists(filePath string) bool {
	if inf, err := os.Stat(filePath); err == nil {
		return inf.IsDir()
	}
	return false
}

// GetRequestID is returns a requestID from a context, or the empty string
func GetRequestID(ctx context.Context) string {
	if rid := ctx.Value(requestIDKey); rid != nil {
		return rid.(string)
	}
	return ""
}

// CopyRequest provides a safe copy of a bodyless request into a new request
func CopyRequest(req *http.Request) *http.Request {
	o := *req
	o.URL = CopyURL(req.URL)
	o.Header = make(http.Header)
	CopyHeaders(o.Header, req.Header)
	o.ContentLength = req.ContentLength
	return &o
}

// CopyURL provides update safe copy by avoiding shallow copying User field
func CopyURL(i *url.URL) *url.URL {
	out := *i
	if i.User != nil {
		iu := *i.User
		out.User = &iu
	}
	return &out
}

// CopyHeaders copies http headers from source to destination, it
// does not overide, but adds multiple headers
func CopyHeaders(dst http.Header, src http.Header) {
	for k, vv := range src {
		dst[k] = append(dst[k], vv...)
	}
}

// TrimPrefixURI standardizes the prefix trimming of the Request.URL.Path and Request.RequestURI, which are squirrely at best.
func TrimPrefixURI(r *http.Request, prefix string) {
	ReplaceURI(r, strings.TrimPrefix(r.URL.Path, prefix), strings.TrimPrefix(r.RequestURI, prefix))
}

// ReplaceURI standardizes the replacement of the Request.URL.Path and Request.RequestURI, which are squirrely at best.
func ReplaceURI(r *http.Request, urlPath, requestURI string) {
	// Replacing RequestURI appears to be required, but we also want to make sure URL.Path
	// is consistent for downstream inspectors
	DebugOut.Print(RequestErrorString(r, fmt.Sprintf("ReplaceURI was: '%s' '%s'\n", r.RequestURI, r.URL.Path)))

	r.URL.Path = urlPath
	if r.URL.RawPath != "" {
		r.URL.RawPath = "" // empty it first
		r.URL.RawPath = r.URL.EscapedPath()
	}

	if r.RequestURI != "" {
		r.RequestURI = requestURI
	}
	DebugOut.Print(RequestErrorString(r, fmt.Sprintf("ReplaceURI now: '%s' '%s'\n", r.RequestURI, r.URL.Path)))
}

// PrettyPrint returns the a JSONified version of the string, or %+v if that's not possible
func PrettyPrint(v interface{}) string {
	return spew.Sdump(v)
}

// FlashEncoding returns a URL-encoded version of the provided string,
// with "+" additionally converted to "%2B"
func FlashEncoding(src string) string {
	uenc := url.QueryEscape(src)
	return strings.Replace(uenc, "+", "%2B", -1)
}

func ipOnly(ip string) string {
	if strings.Contains(ip, ":") {
		iparts := strings.Split(ip, ":")
		ip = iparts[0]
	}

	return ip
}

// ReaderToString reads from a Reader into a Buffer, and then returns the string value of that
func ReaderToString(r io.Reader) string {
	if r == nil {
		return ""
	}
	b := RecyclableBufferPool.Get().(*recyclablebuffer.RecyclableBuffer)
	defer b.Close()
	b.ResetFromReader(r)
	return b.String()
}

// NoopResponseWriter is a hack to support a Response with a status and headers,
// but no body. This is almost never what you want. Really.
type NoopResponseWriter struct {
	code   int
	header http.Header
}

// NewNoopResponseWriter returns a NoopResponseWriter that you almost definitely
// do not want to use.
func NewNoopResponseWriter() NoopResponseWriter {
	return NoopResponseWriter{
		header: make(http.Header),
	}
}

// Header returns an http.Header
func (n *NoopResponseWriter) Header() http.Header {
	return n.header
}

// Write completely ignores whatever you've written, but lies and
// returns the size of whatever you wrote to it, and never an error
func (n *NoopResponseWriter) Write(bytes []byte) (int, error) {
	if n.code == 0 {
		n.WriteHeader(http.StatusOK)
	}
	return len(bytes), nil
}

// WriteHeader changes the response code
func (n *NoopResponseWriter) WriteHeader(statusCode int) {
	n.code = statusCode
}
