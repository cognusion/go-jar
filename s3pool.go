package jar

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cognusion/go-jar/aws"

	"net/http"
	"net/url"
	"strings"
)

func init() {
	Materializers["s3"] = materializeS3

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

func materializeS3(p *Pool) (http.Handler, error) {

	// Define ListMembers
	p.ListMembers = func() []*url.URL {
		return nil
	}

	// Define AddMember
	p.AddMember = func(member string) error {
		return ErrPoolAddMemberNotSupported
	}

	// Define DeleteMember
	p.DeleteMember = func(member string) error {
		return ErrPoolDeleteMemberNotSupported
	}

	// Define RemoveMember
	p.RemoveMember = func(member string) error {
		return ErrPoolRemoveMemberNotSupported
	}

	// Add members
	if len(p.Config.Members) < 1 {
		return nil, ErrPoolNoMembersConfigured
	}

	// We only take the first.
	member := p.Config.Members[0]

	memberURL, err := url.Parse(member)
	if err != nil {
		return nil, err
	}

	DebugOut.Printf("\t\tAdding member '%s'\n", member)
	p.GetMember(memberURL)

	// Add it to
	pool, err := NewS3Pool(member)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
