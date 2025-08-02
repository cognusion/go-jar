package tus

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	tusc "github.com/bdragon300/tusgo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTUS(t *testing.T) {

	//DebugOut = log.New(os.Stderr, "[DEBUG] ", 0)
	//ErrorOut = log.New(os.Stderr, "[ERROR] ", 0)

	Convey("When a TUS is created with an unsupported target prefix, the appropriate error is returned", t, func() {

		tfile := fmt.Sprintf("crap://%s/tustest.fil", "")
		tus, err := New("/", Config{TargetURI: tfile})
		So(err, ShouldEqual, ErrBadTargetPrefix)
		So(tus, ShouldBeNil)
	})

	Convey("When a TUS request is made to a TUS URI path, the upload succeeds and the files match", t, func() {
		var err error

		// Buffer a meg of random data
		buff := bytes.NewBufferString(randString(1024 * 1024))

		tdir, err := os.MkdirTemp("", "tustemp")
		So(err, ShouldBeNil)
		defer os.RemoveAll(tdir)

		tfile := fmt.Sprintf("file://%s/", tdir)
		tfn := "test.txt"

		tus, err := New("/tus/", Config{TargetURI: tfile, AppendFilename: true})
		So(err, ShouldBeNil)
		srv := httptest.NewServer(http.StripPrefix("/tus/", tus))
		defer srv.Close()

		// create the tus client.
		curl, _ := url.Parse(srv.URL + "/tus/")
		client := tusc.NewClient(srv.Client(), curl)

		DebugOut.Printf("Client: %+v\n", client)

		// create the upload
		buffer := bytes.NewReader(buff.Bytes())
		metadata := map[string]string{
			"filename": tfn,
		}
		upload := tusc.Upload{}
		_, err = client.CreateUpload(&upload, buffer.Size(), false, metadata)
		So(err, ShouldBeNil)

		// create the uploader.
		s := tusc.NewUploadStream(client, &upload)

		// Set stream and file pointers to be equal to the remote pointer
		_, err = s.Sync()
		So(err, ShouldBeNil)

		_, err = buffer.Seek(s.Tell(), io.SeekStart)
		So(err, ShouldBeNil)

		written, err := io.Copy(s, buffer)
		if err != nil {
			err = fmt.Errorf("Written %d bytes, error: %w, last response: %v", written, err, s.LastResponse)
		}
		So(err, ShouldBeNil)

		// POST: Upload is done

		// loc is now the uploaded location
		loc := s.Upload.Location

		// Check the result
		time.Sleep(2 * time.Millisecond) // The file may not be there immediately. So we wait a sec
		uParts := strings.Split(loc, "/")
		fName := fmt.Sprintf("%s/%s-%s", tdir, uParts[len(uParts)-1], tfn)
		f, fErr := os.Stat(fName)
		So(fErr, ShouldBeNil)
		So(f.Size(), ShouldEqual, int64(1024*1024))

		// correctURL is proper because of our renaming
		correctURL := fmt.Sprintf("%s-%s", loc, tfn)

		//Printf("Loc: %s\nURL: %s\n", loc, correctURL)
		/*
			Convey("... When a HEAD request is made to an existing TUS URI path, the value is congruent", func() {
				req, err := http.NewRequest("HEAD", loc, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Tus-Resumable", "1.0.0")
				req.Header.Set("X-Request-ID", "H"+randString(7))

				resp, rErr := srv.Client().Do(req)
				So(rErr, ShouldBeNil)
				defer resp.Body.Close()

				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(resp.ContentLength, ShouldEqual, 1024*1024)
			})
		*/

		Convey("... When a GET request is made to an existing TUS URI path, it is Forbidden", func() {
			req, err := http.NewRequest("GET", correctURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Request-ID", "G"+randString(7))
			req.Header.Set("Tus-Resumable", "1.0.0")

			resp, rErr := srv.Client().Do(req)
			So(rErr, ShouldBeNil)
			defer resp.Body.Close()

			So(resp.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
		})

		Convey("... When a DELETE request is made to an existing TUS URI path, it is Forbidden", func() {
			req, err := http.NewRequest("DELETE", correctURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Request-ID", "D"+randString(7))
			req.Header.Set("Tus-Resumable", "1.0.0")

			resp, rErr := srv.Client().Do(req)
			So(rErr, ShouldBeNil)
			defer resp.Body.Close()

			So(resp.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
		})

	})
}

func randString(length int) string {
	charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	outString := make([]byte, length)
	for i := range length {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		outString[i] = randomChar
	}
	return string(outString)
}
