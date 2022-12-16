package jar

import (
	"github.com/cognusion/go-jar/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"net/http"
	"strings"
)

func init() {
	aws.TimingOut = TimingOut
	aws.DebugOut = DebugOut
}

const (
	// ErrInvalidS3URL is returned when the relevant URL parts from the provided S3 URL cannot be derived
	ErrInvalidS3URL = Error("the S3 URL passed is invalid")
)

// S3Pool is an http.Handler that grabs a file from S3 and streams it back to the client
type S3Pool struct {
	session *aws.Session
	bucket  string
}

// NewS3Pool returns an S3Pool or an error
func NewS3Pool(s3url string) (*S3Pool, error) {

	// If the AWSSession isn't initialized, we cannot grab it
	if AWSSession == nil {
		return nil, ErrNoSession
	}

	bucket, _, _ := aws.S3urlToParts(s3url)
	if bucket == "" {
		return nil, ErrInvalidS3URL
	}

	return &S3Pool{
		session: AWSSession,
		bucket:  bucket,
	}, nil

}

// ServeHTTP is a proper http.Handler for authenticated S3 requests
func (s3p *S3Pool) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	filepath := r.URL.RequestURI()

	DebugOut.Printf("Request for %s %s", s3p.bucket, filepath)

	if strings.HasSuffix(filepath, "/") {
		// Request is for a folder/listing. Not allowed
		http.Error(w, ErrRequestError{r, "Folder listing not allowed"}.Error(), http.StatusForbidden)
		return
	}
	_, err := s3p.session.BucketToWriter(s3p.bucket, filepath, w)
	if err != nil {
		// TODO: don't scrape
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fallthrough
			case s3.ErrCodeNoSuchKey:
				http.Error(w, ErrRequestError{r, "Not found"}.Error(), http.StatusNotFound)
				return
			}
			// Anything else, we treat as a generic error
		}

		es := ErrRequestError{r, "Error during download"}.Error()
		ErrorOut.Printf("%s '%s' from bucket %s: %v", es, filepath, s3p.bucket, err)
		http.Error(w, es, http.StatusInternalServerError)
	}
}
