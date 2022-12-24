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

// New returns an initialized TUS
func New(basePath string, config Config) (*TUS, error) {

	composer := tusd.NewStoreComposer()

	// Check the prefix
	if strings.HasPrefix(strings.ToLower(config.TargetURI), "s3://") {
		// Handle S3
		trimTargetURI := strings.TrimPrefix(config.TargetURI, "s3://")
		DebugOut.Printf("NewTUSwithS3: %s -> %s\n", basePath, trimTargetURI)
		store := s3store.New(trimTargetURI, config.S3Client)
		store.UseIn(composer)

		locker := memorylocker.New()
		locker.UseIn(composer)
	} else if strings.HasPrefix(strings.ToLower(config.TargetURI), "file://") {
		// Handle local file
		trimTargetURI := strings.TrimPrefix(config.TargetURI, "file://")
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
	if config.AppendExtension {
		tConfig.NotifyCompleteUploads = true
	}

	handler, err := tusd.NewHandler(tConfig)
	if err != nil {
		return nil, err
	}

	if config.AppendExtension {
		go func() {
			for {
				event := <-handler.CompleteUploads
				if event.Upload.IsFinal {
					// TODO: This is clearly not working
					DebugOut.Printf("TUS Upload of %s finished: %s/%s\n", event.Upload.MetaData["filename"], event.Upload.Storage["Bucket"], event.Upload.Storage["Key"])
				}
			}
		}()
	}

	var t = TUS{
		handler: handler,
		config:  &tConfig,
	}
	return &t, nil
}
