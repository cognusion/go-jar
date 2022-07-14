package jar

import (
	//pr "github.com/cognusion/go-jar/presponsewriter"

	//"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Benchmark_AccessLogHandler(b *testing.B) {
	oa := AccessOut
	AccessOut = log.New(io.Discard, "", 0)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	ok := []byte("ok")
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(ok)
	})

	rr := httptest.NewRecorder()

	handler := AccessLogHandler(testHandler)

	handler.ServeHTTP(rr, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
	AccessOut = oa
}
