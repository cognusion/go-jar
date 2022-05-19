package jar

import (
	. "github.com/smartystreets/goconvey/convey"

	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type baloneyStringKey string

func TestCheckAuthoritativeEmpty(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://somewhere.notgarbage.com:1234/some/where/over/the/rainbow", nil)

	authoritativeDomains = []string{}
	CheckAuthoritative = func(*http.Request) bool { return true }

	Convey("When a request is made and we are not using authoritativedomains, it succeeds", t, func() {
		So(CheckAuthoritative(r), ShouldBeTrue)
	})
}

func TestCheckAuthoritativeOk(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://somewhere.notgarbage.com:1234/some/where/over/the/rainbow", nil)

	authoritativeDomains = []string{".notgarbage.com"}
	CheckAuthoritative = checkAuthoritative

	Convey("When a request is made for a domain we are explicitly authoritative for, it succeeds", t, func() {
		So(CheckAuthoritative(r), ShouldBeTrue)
	})
}

func TestCheckAuthoritativeNoDotOk(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://notgarbage.com:1234/some/where/over/the/rainbow", nil)

	authoritativeDomains = []string{".notgarbage.com"}
	CheckAuthoritative = checkAuthoritative

	Convey("When a request is made for just the domain name we are explicitly authoritative for, it succeeds", t, func() {
		So(CheckAuthoritative(r), ShouldBeTrue)
	})
}

func TestCheckAuthoritativeWeirdBad(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://somewhere.else.notgarbage0com:1234/some/where/over/the/rainbow", nil)

	authoritativeDomains = []string{".notgarbage.com"}
	CheckAuthoritative = checkAuthoritative

	Convey("When a request is made for a domain we are not explicitly authoritative for, it fails", t, func() {
		So(CheckAuthoritative(r), ShouldBeFalse)
	})
}

func TestCheckAuthoritativeBad(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://somewhere.garbage.com:1234/some/where/over/the/rainbow", nil)

	authoritativeDomains = []string{".notgarbage.com"}
	CheckAuthoritative = checkAuthoritative

	Convey("When a request is made for a domain we are not explicitly authoritative for, it fails", t, func() {
		So(CheckAuthoritative(r), ShouldBeFalse)
	})
}

// TestFileExistsFileExists tests FileExists when the file exists
func TestFileExistsFileExists(t *testing.T) {

	if !FileExists("./helpers_test.go") {
		t.Errorf("FileExists says we don't exist.")
	}
}

// TestFileExistsNotExists tests FileExists when the file does not exist
func TestFileExistsNotExists(t *testing.T) {
	Convey("When a FileExists is asked if a non-existent file exists, it properly denies", t, func() {
		So(FileExists("./helpers_testbaloney.go"), ShouldBeFalse)
	})
}

// TestFileExistsFileFolder tests FileExists when the file is a folder
func TestFileExistsFileFolder(t *testing.T) {
	Convey("When a FileExists is asked if an existing forlder exists, it properly denies", t, func() {
		So(FileExists("."), ShouldBeFalse)
	})
}

// TestFolderExistsFolderExists tests FolderExists when the folder exists
func TestFolderExistsFolderExists(t *testing.T) {

	Convey("When a FolderExists is asked if an existing folder exists, it properly agrees", t, func() {
		So(FolderExists("."), ShouldBeTrue)
	})
}

// TestFolderExistsNotExists tests FolderExists when the folder does not exist
func TestFolderExistsNotExists(t *testing.T) {

	if FolderExists("./baloneyfolder") {
		t.Errorf("FolderExists says baloney.")
	}
}

// TestFolderExistsFolderFile tests FolderExists when the folder is a file
func TestFolderExistsFolderFile(t *testing.T) {
	Convey("When a FolderExists is asked if an existing file exists, it properly denies", t, func() {
		So(FolderExists("./helper_test.go"), ShouldBeFalse)
	})
}

func TestStringIfCtxEmpty(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	var baloney baloneyStringKey = "baloney"

	Convey("When a context doesn't exist, StringIfCtx returns empty", t, func() {
		So(StringIfCtx(req, baloney), ShouldBeEmpty)
	})
}

func TestStringIfCtxValue(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	var baloney baloneyStringKey = "baloney"

	Convey("When a context exists, and has value, StringIfCtx returns the value", t, func() {
		req = req.WithContext(context.WithValue(req.Context(), baloney, "test"))

		So(StringIfCtx(req, baloney), ShouldEqual, "test")
	})
}

func TestIpOnly(t *testing.T) {

	Convey("When ipOnly is passed solely an IP address, it is returned as-is", t, func() {
		So(ipOnly("1.2.3.4"), ShouldEqual, "1.2.3.4")
	})

	Convey("When ipOnly is passed an IP address and port, the IP address is returned", t, func() {
		So(ipOnly("1.2.3.4:1234"), ShouldEqual, "1.2.3.4")
	})

	Convey("When ipOnly is passed and empty string, and empty string is returned", t, func() {
		So(ipOnly(""), ShouldEqual, "")
	})
}

func TestGetRequestIDOk(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and a requestid is in the context, the value can be retrieved", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(GetRequestID(r.Context()), ShouldNotBeEmpty)
		})

		handler := SetupHandler(testHandler)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})

}

func TestGetRequestIDNope(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a request is made, and a requestid is not in the context, the empty string is returned", t, func() {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			So(GetRequestID(r.Context()), ShouldBeEmpty)
		})

		handler := testHandler

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}

func TestCopyRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	Convey("When a headerless, bodyless request is copied, they should resemble each other", t, func() {
		r2 := CopyRequest(req)
		So(r2, ShouldResemble, req)
	})
}

func TestCopyRequestHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Hello", "world")
	req.Header.Add("Woofer", "tweeter")

	Convey("When a bodyless request with headers is copied, they should resemble each other", t, func() {
		r2 := CopyRequest(req)
		So(r2, ShouldResemble, req)
	})
}

func TestCopyRequestPOST(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Hello", "world")
	req.Header.Add("Woofer", "tweeter")

	Convey("When a bodyless POST request with headers is copied, they should resemble each other", t, func() {
		r2 := CopyRequest(req)
		So(r2, ShouldResemble, req)
	})
}

func TestReplaceURI(t *testing.T) {

	req, err := http.NewRequest("GET", "/garbage/plate/food", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "/garbage/plate/food"

	Convey("When a request is made, and the URI is replaced, that URI should be in the Request and its URL", t, func() {
		So(req.URL.EscapedPath(), ShouldEqual, "/garbage/plate/food")
		So(req.RequestURI, ShouldEqual, "/garbage/plate/food")

		ReplaceURI(req, "/food", "/food")

		So(req.URL.EscapedPath(), ShouldEqual, "/food")
		So(req.RequestURI, ShouldEqual, "/food")

	})
}

func TestReplaceURIPrefix(t *testing.T) {

	req, err := http.NewRequest("GET", "/garbage/plate/food", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RequestURI = "/garbage/plate/food"

	Convey("When a request is made, and prefixes are stripped, that prefix should not be in the Request or its URL", t, func() {
		So(req.URL.EscapedPath(), ShouldEqual, "/garbage/plate/food")
		So(req.RequestURI, ShouldEqual, "/garbage/plate/food")

		TrimPrefixURI(req, "/garbage/plate")

		So(req.URL.EscapedPath(), ShouldEqual, "/food")
		So(req.RequestURI, ShouldEqual, "/food")

	})
}

func TestFlashEncoding(t *testing.T) {
	s := "This string is 1+1"

	Convey("When \"FlashEncoding\" a string, it should never have a space, '+', or '/', because boom", t, func() {
		e := FlashEncoding(s)
		So(e, ShouldNotContainSubstring, " ")
		So(e, ShouldNotContainSubstring, "+")
		So(e, ShouldNotContainSubstring, "/")
	})
}

func TestReaderToString(t *testing.T) {
	s := "This is a ridiculous!!\n\t\t\t\t\t\r\ndsfldfjhsldfjhasjfahslkjfhldjfhalskhdfs"
	b := bytes.NewBufferString(s)

	Convey("When a Reader is passed to ReaderToString, the appropriate string is returned", t, func() {
		So(ReaderToString(b), ShouldEqual, s)
	})
}

func TestReaderToStringNil(t *testing.T) {

	Convey("When a nil Reader is passed to ReaderToString, the empty string is returned", t, func() {
		So(ReaderToString(nil), ShouldEqual, "")
	})
}
