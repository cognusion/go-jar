package tus

import "github.com/tus/tusd/pkg/s3store"

// Config encapsulates various options passable to New
type Config struct {
	// TargetURI is a `file://` or `s3://` URI to designate where the upload should go
	TargetURI string
	// AppendExtension renames (COPY,DELETE) the file after upload. This can result in
	// increased costs for paid storage services
	AppendExtension bool
	// S3Client is an s3.S3 to be used if TargetURI is an `s3://`
	S3Client s3store.S3API
}
