package jar

import (
	"github.com/cognusion/go-timings"
	"github.com/vulcand/oxy/v2/forward"

	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"
)

var (
/*
// defaultDebugTrip is a very useful debugging RoundTripper

	defaultDebugTrip = DebugTrip{
		RTFunc: func(r *http.Request) (*http.Response, error) {
			if r == nil {
				return nil, fmt.Errorf("nil request")
			}

			var (
				body []byte
				err  error
				b    string
			)

			if r.Body != nil {
				body, err = io.ReadAll(r.Body)
				r.Body.Close()
			}
			w := http.Response{}
			b = fmt.Sprintf("Request: %+v\n\n", r)
			if err != nil {
				w.StatusCode = http.StatusBadRequest
				b += fmt.Sprintf("Body: Error: %s\n\n", err)
			} else {
				w.StatusCode = http.StatusOK
				b += fmt.Sprintf("Body: %s\n\n", string(body))
			}

			w.Status = http.StatusText(w.StatusCode)
			buf := RecyclableBufferPool.Get()
			buf.Reset([]byte(b))
			w.Body = buf
			return &w, nil
		},
	}
*/
)

func init() {
	Handlers["dumphandler"] = DumpHandler
	Finishers["dumpfinisher"] = DumpFinisher
	Finishers["ok"] = OkFinisher
	Finishers["date"] = DateFinisher
	Finishers["minutedelayer"] = MinuteDelayer
	Finishers["test"] = TestFinisher
	Finishers["minutestreamer"] = MinuteStreamer
	Finishers["requestid"] = RequestIDFinisher
}

// TestFinisher is a special finisher that outputs some detectables
func TestFinisher(w http.ResponseWriter, r *http.Request) {
	defer timings.Track("TestFinisher", time.Now(), TimingOut)

	me, err := os.Hostname()
	if err != nil {
		me = "localhost"
	}
	hr := forward.HeaderRewriter{TrustForwardHeader: true, Hostname: me}
	hr.Rewrite(r)

	w.Write([]byte(fmt.Sprintf("\nProtocol: %s\n  Major: %d\n  Minor: %d\n", r.Proto, r.ProtoMajor, r.ProtoMinor)))

	if r.TLS != nil {
		w.Write([]byte("\nTLS:\n"))
		w.Write([]byte(fmt.Sprintf("  Version: %s\n", SslVersions.Suite(r.TLS.Version))))
		w.Write([]byte(fmt.Sprintf("  CipherSuite: %s\n", Ciphers.Suite(r.TLS.CipherSuite))))
	}

	w.Write([]byte("\nHeaders:\n"))
	hkeys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		hkeys = append(hkeys, k)
	}
	sort.Strings(hkeys)

	for _, k := range hkeys {
		v := r.Header.Values(k)

		if !Conf.GetBool(ConfigDebug) && (k == "Cookie") {
			w.Write([]byte(fmt.Sprintf("  %s: <..redacted..>\n", k)))
			continue
		}
		for _, av := range v {
			w.Write([]byte(fmt.Sprintf("  %s: %s\n", k, av)))
		}
	}

	w.Write([]byte("\nCookies:\n"))
	for _, v := range r.Cookies() {
		val := v.Value
		w.Write([]byte(fmt.Sprintf("  %s: %s\n", v.Name, val)))
	}

}

// DumpHandler is a special handler that ships a ton of request output to DebugLog
func DumpHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		DebugOut.Printf("\nRequest: %+v\nHeaders: %+v\nCookies: %+v\nContext: %+v\n", r, r.Header, r.Cookies(), r.Context())

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// DumpFinisher is a special finisher that reflects a ton of request output
func DumpFinisher(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Request: %+v\nHeaders: %+v\nCookies: %+v\nContext: %+v\n", r, r.Header, r.Cookies(), r.Context())))
}

// MinuteDelayer is a special finisher that waits for 60s before returning
func MinuteDelayer(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Minute)
	w.Write([]byte("Ok"))
}

// OkFinisher is a Finisher that simply returns "Ok", for throughput testing.
func OkFinisher(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok"))
}

// DateFinisher is a Finisher that simply returns the current system datestamp as a string, for cache testing.
func DateFinisher(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}

// RequestIDFinisher is a Finisher that simply returns the current requestID a random number of times, for grind testing.
func RequestIDFinisher(w http.ResponseWriter, r *http.Request) {
	var requestID string
	if rid := r.Context().Value(requestIDKey); rid != nil {
		requestID = rid.(string)
	}
	w.Header().Set(Conf.GetString(ConfigRequestIDHeaderName), requestID)
	rnd, _ := rand.Int(rand.Reader, randMax)
	rv := rnd.Int64()/2 + 1
	out := []byte(fmt.Sprintf("%s %d\n", requestID, rv))
	for i := int64(0); i < rv; i++ {
		w.Write(out)
	}
}

// MinuteStreamer is a special finisher that writes the next number, once a secondish, for 60 iterations
func MinuteStreamer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	ctx := r.Context()

	for i := 1; i <= 60; i++ {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine
		case <-time.After(time.Second):
			w.Write([]byte(fmt.Sprintf("%d\n", i)))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// DebugTrip is an http.RoundTripper with a pluggable core func to aid in debugging
type DebugTrip struct {
	// RTFunc is executed when RoundTrip() is called on a request.
	// It can be changed at any point to aid in changing conditions
	RTFunc func(*http.Request) (*http.Response, error)
}

// RoundTrip is the Request executor
func (d *DebugTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	return d.RTFunc(r)
}
