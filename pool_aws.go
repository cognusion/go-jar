package jar

import (
	"net"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cognusion/go-jar/aws"
	"github.com/vulcand/oxy/v2/roundrobin"

	"net/http"
	"net/url"
	"strings"
)

func init() {
	// Set up the Materializers
	Materializers["s3"] = materializeS3

	// Set up the MemberBuilders
	MemberBuilders["http"] = append(MemberBuilders["http"], ec2HTTPMember)
	MemberBuilders["https"] = append(MemberBuilders["https"], ec2HTTPMember)
	MemberBuilders["ws"] = append(MemberBuilders["ws"], ec2HTTPMember)

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

func ec2HTTPMember(conf *PoolConfig, u *url.URL, m *Member) *Member {

	if m == nil {
		m = NewMember(u)
	}

	// If we're EC2-aware, and this Pool is using EC2Affinity, let's find out what the AZ is
	if AWSSession != nil && Conf.GetBool(ConfigEC2) && conf.EC2Affinity {

		if !conf.Prune {
			// Not forcing this, but chances are you don't know what you're doing.
			ErrorOut.Printf("WARNING!!! Pool %s is using EC2Affinity but not Prune. This may delay or prevent expected failover to non-local members in the event of a member failure.\n", conf.Name)
			DebugOut.Printf("WARNING!!! Pool %s is using EC2Affinity but not Prune. This may delay or prevent expected failover to non-local members in the event of a member failure.\n", conf.Name)
		}

		// If the Hostname isn't just digits and dots, it's a name and not a number, make it a number
		if ok, err := regexp.MatchString(`[^\d\.]`, u.Hostname()); err == nil && ok {
			// hostname is probably not an address
			addrs, err := net.LookupHost(u.Hostname())
			if err != nil {
				DebugOut.Printf("Error resolving hostname '%s': %s\n", u.Hostname(), err)
			} else if len(addrs) > 0 {
				// Take the first address
				m.Address = addrs[0]
			}
		}

		if az, azerr := AWSSession.GetInstanceAZByIP(m.Address); azerr != nil {
			ErrorOut.Printf("Error adding EC2-aware pool-member '%s' to %s: %s\n", m.Address, conf.Name, azerr)
		} else if az == "" {
			DebugOut.Printf("\t\t\tPool %s has member %s that has no AZ\n", conf.Name, m.Address)
		} else if az == AWSSession.Me.AvailabilityZone {
			DebugOut.Printf("\t\t\tPool %s has member %s that is AZ-local!\n", conf.Name, m.Address)
			m.Weight = roundrobin.Weight(LocalMemberWeight)
			m.AZ = az
		} else {
			DebugOut.Printf("\t\t\tPool %s has member %s that is not AZ-local (%s)\n", conf.Name, m.Address, az)
			m.AZ = az
		}
	}

	return m
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
