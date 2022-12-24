package tus

import "github.com/aws/aws-sdk-go/service/s3"

// Config encapsulates various options passable to New
type Config struct {
	// TargetURI is a `file://` or `s3://` URI to designate where the upload should go
	TargetURI string
	// AppendFilename renames (COPY,DELETE) the file after upload to append `-filename.ext`. This can result in
	// increased costs for paid storage services
	AppendFilename bool
	// S3Client is an s3.S3 to be used if TargetURI is an `s3://`
	S3Client *s3.S3
}
