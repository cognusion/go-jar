package tus

import (
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"github.com/tus/tusd/pkg/s3store"

	"io"
	"log"
	"net/http"
	"strings"
)

const (
	// ErrBadTargetPrefix is returned by HandleFinisher if the requested TUS targetURL prefix does not exist
	ErrBadTargetPrefix = Error("requested targetURI prefix is not valid")
	// ErrUnimplemented is returned by the default CallBack in the even callbacks were enabled but a callback handler wasn't
	ErrUnimplemented = Error("unimplemented")
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
)

// Error is an error type
type Error string

// Error returns the stringified version of Error
func (e Error) Error() string {
	return string(e)
}

// TUS is a Finisher implementing the tus.io Open Protocol for Resumable Uploads
type TUS struct {
	targetURI string
	basePath  string
	handler   *tusd.UnroutedHandler
	config    *tusd.Config
	CallBack  func(tusd.HookEvent) error
}

func (t *TUS) defaultCallBack(e tusd.HookEvent) error {
	return ErrUnimplemented
}

func (t *TUS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		t.handler.PostFile(w, r)
	case http.MethodHead:
		t.handler.HeadFile(w, r)
	case http.MethodPatch:
		t.handler.PatchFile(w, r)
	case http.MethodGet:
		if !t.config.DisableDownload {
			t.handler.GetFile(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	case http.MethodDelete:
		// Only attach the DELETE handler if the Terminate() method is provided
		if t.config.StoreComposer.UsesTerminater && !t.config.DisableTermination {
			t.handler.DelFile(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// jarTUSDataStore is a rage interface because tusd.DataStore doesn't require
// UseIn, although it's implemented universally and not doing so means major
// implmentation friction. JAR requires DataStores to implment it. Period.
type jarTUSDataStore interface {
	tusd.DataStore
	UseIn(composer *tusd.StoreComposer)
}

// NewTUS returns an initialized TUS, or an error if targetURI is not one of “s3://“ or “file://“ or
// the target itself is a problem. basePath should be the URI base.
func NewTUS(targetURI, basePath string) (*TUS, error) {

	var (
		store jarTUSDataStore
	)

	// Check the prefix
	if strings.HasPrefix(strings.ToLower(targetURI), "s3://") {
		// Handle S3
		store = s3store.New(targetURI, nil)

	} else if strings.HasPrefix(strings.ToLower(targetURI), "file://") {
		// Handle local file
		trimTargetURI := strings.TrimPrefix(targetURI, "file://")
		store = filestore.New(trimTargetURI)

	} else {
		return nil, ErrBadTargetPrefix
	}

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	tConfig := tusd.Config{
		BasePath:              basePath,
		StoreComposer:         composer,
		Logger:                DebugOut,
		NotifyCompleteUploads: false, // TODO
		DisableDownload:       true,  // TODO
		DisableTermination:    true,  // TODO
	}
	handler, err := tusd.NewUnroutedHandler(tConfig)
	if err != nil {
		return nil, err
	}

	var t TUS
	t = TUS{
		targetURI: targetURI,
		basePath:  basePath,
		handler:   handler,
		config:    &tConfig,
		CallBack:  t.defaultCallBack,
	}
	return &t, nil
}
