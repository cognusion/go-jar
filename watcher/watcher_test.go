package watcher

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestWatcher(t *testing.T) {

	tfile := "/tmp/watchertest1"

	Convey("When a test file is created", t, func() {
		defer func() { os.Remove(tfile) }()
		terr := touchFile(tfile, "hello")
		So(terr, ShouldBeNil)

		Convey("and a new watcher intializes", func() {
			// set up the watcher

			w, err := NewWatcher()
			if err != nil {
				t.Errorf("Error creating watcher: %s\n", err)
				return
			}
			defer w.Close()

			var called int64
			w.Add(tfile, func(e fsnotify.Event) {
				atomic.AddInt64(&called, 1)
				time.Sleep(10 * time.Millisecond)
			})

			Convey("we see our function called when the file is touched", func() {
				//run the tests
				terr := touchFile(tfile, "goodbye")
				So(terr, ShouldBeNil)
				<-time.After(15 * time.Millisecond)
				So(atomic.LoadInt64(&called), ShouldEqual, 1)
			})
		})

	})
}

func TestWatcherPrefixed(t *testing.T) {

	tfile := "file:///tmp/watchertest2"

	Convey("When a test file is created", t, func() {
		defer func() { os.Remove(tfile) }()
		terr := touchFile(tfile, "hello")
		So(terr, ShouldBeNil)

		Convey("and a new watcher intializes", func() {
			// set up the watcher

			w, err := NewWatcher()
			if err != nil {
				t.Errorf("Error creating watcher: %s\n", err)
				return
			}
			defer w.Close()

			var called int64
			w.Add(tfile, func(e fsnotify.Event) {
				atomic.AddInt64(&called, 1)
				time.Sleep(10 * time.Millisecond)
			})

			Convey("we see our function called when the file is touched", func() {
				//run the tests
				terr := touchFile(tfile, "goodbye")
				So(terr, ShouldBeNil)
				<-time.After(5 * time.Millisecond)
				So(atomic.LoadInt64(&called), ShouldEqual, 1)
			})
		})

	})
}

func TestWatcherDelete(t *testing.T) {

	tfile := "/tmp/watchertest3"

	Convey("When a test file is created", t, func() {
		defer func() { os.Remove(tfile) }()
		terr := touchFile(tfile, "hello")
		So(terr, ShouldBeNil)

		Convey("and a new watcher intializes", func() {
			// set up the watcher

			w, err := NewWatcher()
			if err != nil {
				t.Errorf("Error creating watcher: %s\n", err)
				return
			}
			defer w.Close()

			var called int64
			w.Add(tfile, func(e fsnotify.Event) {
				atomic.AddInt64(&called, 1)
				time.Sleep(10 * time.Millisecond)
			})

			Convey("we see our function called when the file is touched", func() {
				//run the tests
				w.Delete(tfile)

				terr := touchFile(tfile, "goodbye")
				So(terr, ShouldBeNil)
				<-time.After(5 * time.Millisecond)
				So(atomic.LoadInt64(&called), ShouldEqual, 0)
			})
		})

	})
}

func TestWatcherDeletePrefix(t *testing.T) {

	tfile := "file:///tmp/watchertest3"

	Convey("When a test file is created", t, func() {
		defer func() { os.Remove(tfile) }()
		terr := touchFile(tfile, "hello")
		So(terr, ShouldBeNil)

		Convey("and a new watcher intializes", func() {
			// set up the watcher

			w, err := NewWatcher()
			if err != nil {
				t.Errorf("Error creating watcher: %s\n", err)
				return
			}
			defer w.Close()

			var called int64
			w.Add(tfile, func(e fsnotify.Event) {
				atomic.AddInt64(&called, 1)
				time.Sleep(10 * time.Millisecond)
			})

			Convey("we see our function called when the file is touched", func() {
				//run the tests
				w.Delete(tfile)

				terr := touchFile(tfile, "goodbye")
				So(terr, ShouldBeNil)
				<-time.After(5 * time.Millisecond)
				So(atomic.LoadInt64(&called), ShouldEqual, 0)
			})
		})

	})
}

func TestWatcherLoopwrites(t *testing.T) {

	tfile := "/tmp/watchertest4"

	Convey("When a test file is created, and a new watcher intializes, ", t, func() {
		defer func() { os.Remove(tfile) }()
		terr := touchFile(tfile, "hello")
		So(terr, ShouldBeNil)

		// set up the watcher

		w, err := NewWatcher()
		if err != nil {
			t.Errorf("Error creating watcher: %s\n", err)
			return
		}
		defer w.Close()

		Convey("we see our function called when the file is touched many times", func() {

			var called int64
			w.Add(tfile, func(e fsnotify.Event) {
				atomic.AddInt64(&called, 1)
				time.Sleep(1 * time.Millisecond)
			})

			// Wait a spell
			<-time.After(10 * time.Millisecond)

			for i := 0; i < 10; i++ {
				terr := touchFile(tfile, fmt.Sprintf("goodbye: %d", i))
				So(terr, ShouldBeNil)
				<-time.After(10 * time.Millisecond)
			}

			<-time.After(15 * time.Millisecond)
			So(atomic.LoadInt64(&called), ShouldEqual, 10)
		})

	})
}

func touchFile(filename, phrase string) error {
	// create file
	defer DebugOut.Printf("Touched '%s' with '%s'\n", filename, phrase)
	filename = strings.TrimPrefix(filename, "file://")

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(phrase)
	if err != nil {
		return err
	}

	return nil
}
