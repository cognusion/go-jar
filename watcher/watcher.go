// Package watcher is used to keep an eye on file system events,
// and trigger actions when those events are of interest.
package watcher

import (
	"github.com/fsnotify/fsnotify"

	"io/ioutil"
	"log"
	"strings"
	"sync"

	"golang.org/x/sync/singleflight"
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(ioutil.Discard, "[DEBUG] ", 0)
	// ErrorOut is a log.Logger for error messages
	ErrorOut = log.New(ioutil.Discard, "", 0)
)

// WatchHandlerFunc takes an fsnotify.Event to add on. A WatchHandlerFunc can assume it will not be called more than once while running
type WatchHandlerFunc func(fsnotify.Event)

// Watcher is an implementation of fsnotify.Watcher with some rails
// and opinions.
type Watcher struct {
	// Close should be called when the Watcher is no longer of use. It is safe to call multiple times.
	Close func()

	fs        *fsnotify.Watcher
	watchMap  sync.Map
	errorChan chan error
}

// NewWatcher returns a Watcher or an error if a problem occurred creating it (very rare)
func NewWatcher() (*Watcher, error) {
	return NewWatcherWithErrorChan(make(chan error))
}

// NewWatcherWithErrorChan returns a Watcher or an error if a problem occurred creating it (very rare). The supplied chan will receive non-nil
// errors if they occur
func NewWatcherWithErrorChan(errorChan chan error) (*Watcher, error) {
	var (
		watcher Watcher
		err     error
	)

	watcher.errorChan = errorChan
	watcher.fs, err = fsnotify.NewWatcher()
	if err != nil {
		ErrorOut.Println(err)
		return nil, err
	}

	done := make(chan struct{})
	watcher.Close = func() {
		watcher.Close = func() {}
		close(done)
	}
	watcher.run(done)

	return &watcher, nil
}

// Add takes a filename, and a WatchHandlerFunc to call if the file changes
func (w *Watcher) Add(file string, wf WatchHandlerFunc) error {
	file = strings.TrimPrefix(file, "file://")

	err := w.fs.Add(file)
	if err != nil {
		return err
	}
	w.watchMap.Store(file, wf)

	return nil
}

// Delete takes a filename, and removes it from being watched
func (w *Watcher) Delete(file string) error {
	file = strings.TrimPrefix(file, "file://")

	err := w.fs.Remove(file)
	if err != nil {
		return err
	}
	w.watchMap.Delete(file)

	return nil
}

// run fires off a goto to run the main watch loop
func (w *Watcher) run(done chan struct{}) {
	go func() {
		defer w.fs.Close()
		defer close(w.errorChan)

		// Using singleflight to ensure we're not spamming WatchHandlerFuncs
		var requestGroup singleflight.Group

		for {
			select {
			case event, cok := <-w.fs.Events:
				DebugOut.Printf("Watched event: %+v\n", event)
				if !cok {
					ErrorOut.Printf("Fsnotify channel apparently closed!\n")
					return
				}
				if event.Name == "" {
					continue
				}

				var (
					wfunc WatchHandlerFunc
				)

				if wfunci, ok := w.watchMap.Load(event.Name); !ok {
					ErrorOut.Printf("Events for '%s' but no map entry?!\n", event.Name)
					continue
				} else {
					// Assert it
					wfunc = wfunci.(WatchHandlerFunc)
				}

				switch event.Op {
				case fsnotify.Write:

					go requestGroup.Do(event.Name, func() (interface{}, error) {
						wfunc(event)
						return nil, nil
					})

					/*
						case fsnotify.Remove, fsnotify.Rename:
							// TODO: TL;DR: This is largely broken. Thankfully, edge case.
							// It appears to be a bug in fsnotify, and I haven't waded through it. The 99% case is "Write", which is handled above and
							// works fine. If you use an editor to make changes that moves and swaps, then you get here, and the file ends up orphaned, even
							// though it shouldn't, and no errors occur, just that the next event that fires has no name, so it can't be handled.
							if waiting {
								continue
							}
							waiting = true // skip subsequent notifies until we handle one
							go HandleReload(&waiting, event.Name, mfiles)
							// We have to re-add the file. Lame, I know
							err := watcher.Add(event.Name)
							if err != nil {
								DebugOut.Printf("Error re-adding '%s' to the watcher: %s\n", event.Name, err)
							}
					*/
				}
			case err, ok := <-w.fs.Errors:
				if !ok {
					ErrorOut.Printf("Fsnotify channel apparently closed!\n")
					return
				}

				ErrorOut.Println("Fsnotify error:", err)
				// Don't block or freak out
				select {
				case w.errorChan <- err:
				default:
				}
			case <-done:
				return
			}
		}
	}()
}
