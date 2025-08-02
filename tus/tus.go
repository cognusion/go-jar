package tus

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/memorylocker"
	"github.com/tus/tusd/v2/pkg/s3store"
	"golang.org/x/exp/slog"
)

const (
	// ErrBadTargetPrefix is returned by HandleFinisher if the requested TUS targetURL prefix does not exist
	ErrBadTargetPrefix = Error("requested targetURI prefix is not valid")
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
	// ErrorOut is a log.Logger for error messages
	ErrorOut = log.New(io.Discard, "[ERROR] ", 0)

	bofcACL = s3types.ObjectCannedACLBucketOwnerFullControl
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
	s3      *s3.Client
}

// New returns an initialized TUS. **WARNING:** Do not set `Config.S3Client`
// unless you're using S3 as the target.
func New(basePath string, config Config) (*TUS, error) {

	composer := tusd.NewStoreComposer()

	// Check the prefix
	if strings.HasPrefix(strings.ToLower(config.TargetURI), "s3://") {
		// Handle S3
		trimTargetURI := strings.TrimPrefix(config.TargetURI, "s3://")
		DebugOut.Printf("NewTUS S3: %s -> %s\n", basePath, trimTargetURI)
		store := s3store.New(trimTargetURI, config.S3Client)
		store.UseIn(composer)

		locker := memorylocker.New()
		locker.UseIn(composer)
	} else if strings.HasPrefix(strings.ToLower(config.TargetURI), "file://") {
		// Handle local file
		trimTargetURI := strings.TrimPrefix(config.TargetURI, "file://")
		DebugOut.Printf("NewTUS File: %s -> %s\n", basePath, trimTargetURI)
		store := filestore.New(trimTargetURI)
		store.UseIn(composer)
	} else {
		return nil, ErrBadTargetPrefix
	}

	tConfig := tusd.Config{
		BasePath:           basePath,
		StoreComposer:      composer,
		Logger:             slog.New(slog.NewTextHandler(DebugOut.Writer(), nil)),
		DisableDownload:    true, // TODO
		DisableTermination: true, // TODO
	}
	if config.AppendFilename {
		tConfig.NotifyCompleteUploads = true
	}

	handler, err := tusd.NewHandler(tConfig)
	if err != nil {
		return nil, err
	}

	var t = TUS{
		handler: handler,
		config:  &tConfig,
		s3:      config.S3Client,
	}

	if config.AppendFilename {
		DebugOut.Println("TUS appending filenames")
		if t.s3 != nil {
			// S3
			go t.eventHandlerS3()
		} else {
			// File
			go t.eventHandlerFile()
		}
	}

	return &t, nil
}

func (t *TUS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pass the request on to TUS
	DebugOut.Println("TUS Handler...")
	t.handler.ServeHTTP(w, r)
}

func (t *TUS) eventHandlerS3() {
	defer DebugOut.Println("TUS eventHandler exiting...")
	for {
		event := <-t.handler.CompleteUploads
		if event.Upload.IsFinal || !event.Upload.IsPartial {
			if event.Upload.MetaData["filename"] == "" {
				DebugOut.Print("TUS Upload finished, but no filename present\n")
			} else {
				DebugOut.Printf("TUS Upload of %s finished: %s/%s\n", event.Upload.MetaData["filename"], event.Upload.Storage["Bucket"], event.Upload.Storage["Key"])
				err := t.rename(event.Upload.Storage["Bucket"], event.Upload.Storage["Key"], fmt.Sprintf("%s-%s", event.Upload.Storage["Key"], event.Upload.MetaData["filename"]))
				if err != nil {
					ErrorOut.Printf("TUS Rename error %+v : %+v\n", err, event)
				}
			}
		}
	}
}

func (t *TUS) eventHandlerFile() {
	defer DebugOut.Println("TUS eventHandler exiting...")
	for {
		event := <-t.handler.CompleteUploads
		if event.Upload.IsFinal || !event.Upload.IsPartial {
			// File
			if event.Upload.MetaData["filename"] == "" {
				DebugOut.Print("TUS Upload finished, but no filename present\n")
			} else {
				DebugOut.Printf("TUS Upload of %s finished: %s\n", event.Upload.MetaData["filename"], event.Upload.Storage["Path"])
				err := t.rename("", event.Upload.Storage["Path"], fmt.Sprintf("%s-%s", event.Upload.Storage["Path"], event.Upload.MetaData["filename"]))
				if err != nil {
					ErrorOut.Printf("TUS Rename error %+v : %+v\n", err, event)
				}
			}
		}
	}
}

// rename is a copy followed by a delete
func (t *TUS) rename(bucket, old, new string) error {

	if t.s3 != nil {
		// S3
		cpCfg := s3.CopyObjectInput{
			ACL:        bofcACL,
			Bucket:     aws.String(bucket),
			CopySource: aws.String(fmt.Sprintf("%s/%s", bucket, old)),
			Key:        aws.String(new),
		}

		cpCfgInfo := s3.CopyObjectInput{
			ACL:        bofcACL,
			Bucket:     aws.String(bucket),
			CopySource: aws.String(fmt.Sprintf("%s/%s%s", bucket, old, ".info")),
			Key:        aws.String(fmt.Sprintf("%s%s", new, ".info")),
		}

		delCfg := s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(old),
		}

		delCfgInfo := s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fmt.Sprintf("%s%s", old, ".info")),
		}

		for i, c := range []s3.CopyObjectInput{cpCfg, cpCfgInfo} {
			if _, err := t.s3.CopyObject(context.Background(), &c); err != nil {
				return err
			} else if _, err = t.s3.DeleteObject(context.Background(), &[]s3.DeleteObjectInput{delCfg, delCfgInfo}[i]); err != nil {
				return err
			}
		}
	} else {
		// File
		oldInfo := fmt.Sprintf("%s%s", old, ".info")
		newInfo := fmt.Sprintf("%s%s", new, ".info")
		if err := os.Rename(old, new); err != nil {
			return err
		}
		if err := os.Rename(oldInfo, newInfo); err != nil {
			return err
		}
	}
	return nil
}
