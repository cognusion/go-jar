package tus

import (
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"github.com/tus/tusd/pkg/memorylocker"
	"github.com/tus/tusd/pkg/s3store"

	"io"
	"log"
	"net/http"
	"strings"
)

const (
	// ErrBadTargetPrefix is returned by HandleFinisher if the requested TUS targetURL prefix does not exist
	ErrBadTargetPrefix = Error("requested targetURI prefix is not valid")
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
	handler *tusd.Handler
	config  *tusd.Config
}

func (t *TUS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pass the request on to TUS
	DebugOut.Println("TUS Handler...")
	t.handler.ServeHTTP(w, r)
}

// NewTUS returns an initialized TUS for targetURIs of `file://`.
// basePath should be the URI base.
func NewTUS(targetURI, basePath string) (*TUS, error) {
	return newTUS(targetURI, basePath, nil)
}

// NewTUSwithS3 returns an initialized TUS for targetURIs of `s3://`.
// s3api should be an s3.S3. basePath should be the URI base.
func NewTUSwithS3(targetURI, basePath string, s3api s3store.S3API) (*TUS, error) {
	return newTUS(targetURI, basePath, s3api)
}

func newTUS(targetURI, basePath string, s3api s3store.S3API) (*TUS, error) {

	composer := tusd.NewStoreComposer()

	// Check the prefix
	if strings.HasPrefix(strings.ToLower(targetURI), "s3://") {
		// Handle S3
		trimTargetURI := strings.TrimPrefix(targetURI, "s3://")
		DebugOut.Printf("NewTUSwithS3: %s -> %s\n", basePath, trimTargetURI)
		store := s3store.New(trimTargetURI, s3api)
		store.UseIn(composer)

		locker := memorylocker.New()
		locker.UseIn(composer)
	} else if strings.HasPrefix(strings.ToLower(targetURI), "file://") {
		// Handle local file
		trimTargetURI := strings.TrimPrefix(targetURI, "file://")
		DebugOut.Printf("NewTUS: %s -> %s\n", basePath, trimTargetURI)
		store := filestore.New(trimTargetURI)
		store.UseIn(composer)
	} else {
		return nil, ErrBadTargetPrefix
	}

	tConfig := tusd.Config{
		BasePath:           basePath,
		StoreComposer:      composer,
		Logger:             DebugOut,
		DisableDownload:    true, // TODO
		DisableTermination: true, // TODO
	}

	handler, err := tusd.NewHandler(tConfig)
	if err != nil {
		return nil, err
	}

	var t = TUS{
		handler: handler,
		config:  &tConfig,
	}
	return &t, nil
}
