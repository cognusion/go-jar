package tus

import (
	tusc "github.com/eventials/go-tus"
	. "github.com/smartystreets/goconvey/convey"

	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestTUS(t *testing.T) {

	//DebugOut = log.New(os.Stderr, "[DEBUG] ", 0)

	Convey("When a TUS is created with an unsupported target prefix, the appropriate error is returned", t, func() {

		tfile := fmt.Sprintf("crap://%s/tustest.fil", "")
		tus, err := NewTUS(tfile, "/")
		So(err, ShouldEqual, ErrBadTargetPrefix)
		So(tus, ShouldBeNil)
	})

	Convey("When a TUS request is made to a TUS URI path, the upload succeeds and the files match", t, func() {
		// Buffer a meg of random data
		buff := bytes.NewBufferString(randString(1024 * 1024))

		tdir, err := os.MkdirTemp("", "tustemp")
		So(err, ShouldBeNil)
		defer os.RemoveAll(tdir)

		tfile := fmt.Sprintf("file://%s/", tdir)

		tus, err := NewTUS(tfile, "/")
		So(err, ShouldBeNil)

		srv := httptest.NewServer(tus)
		defer srv.Close()

		// create the tus client.
		tConfig := tusc.DefaultConfig()
		tConfig.HttpClient = srv.Client()
		client, err := tusc.NewClient(srv.URL, tConfig)
		So(err, ShouldBeNil)
		client.Header.Add("X-Request-ID", "U"+randString(7))

		// create an upload from the buffer.
		upload := tusc.NewUploadFromBytes(buff.Bytes())
		So(upload, ShouldNotBeNil)

		// create the uploader.
		uploader, err := client.CreateUpload(upload)
		So(err, ShouldBeNil)

		DebugOut.Printf("URL: %s\n", uploader.Url())
		// start the uploading process.
		uErr := uploader.Upload()
		So(uErr, ShouldBeNil)

		// Check the result
		uParts := strings.Split(uploader.Url(), "/")
		fName := fmt.Sprintf("%s/%s", tdir, uParts[len(uParts)-1])
		f, fErr := os.Stat(fName)
		So(fErr, ShouldBeNil)
		So(f.Size(), ShouldEqual, int64(1024*1024))

		Convey("... When a HEAD request is made to an existing TUS URI path, the value is congruent", func() {
			req, err := http.NewRequest("HEAD", uploader.Url(), nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("X-Request-ID", "H"+randString(7))

			resp, rErr := srv.Client().Do(req)
			So(rErr, ShouldBeNil)
			defer resp.Body.Close()

			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(resp.ContentLength, ShouldEqual, 1024*1024)
		})

		Convey("... When a GET request is made to an existing TUS URI path, it is Forbidden", func() {
			req, err := http.NewRequest("GET", uploader.Url(), nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("X-Request-ID", "G"+randString(7))

			resp, rErr := srv.Client().Do(req)
			So(rErr, ShouldBeNil)
			defer resp.Body.Close()

			So(resp.StatusCode, ShouldEqual, http.StatusForbidden)
			So(resp.ContentLength, ShouldEqual, 0)
		})

		Convey("... When a DELETE request is made to an existing TUS URI path, it is Forbidden", func() {
			req, err := http.NewRequest("DELETE", uploader.Url(), nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("X-Request-ID", "D"+randString(7))

			resp, rErr := srv.Client().Do(req)
			So(rErr, ShouldBeNil)
			defer resp.Body.Close()

			So(resp.StatusCode, ShouldEqual, http.StatusForbidden)
			So(resp.ContentLength, ShouldEqual, 0)
		})
	})
}

func randString(length int) string {
	charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	outString := make([]byte, length)
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		outString[i] = randomChar
	}
	return string(outString)
}
