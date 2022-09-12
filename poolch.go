package jar

import (
	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
	"github.com/vulcand/oxy/roundrobin"

	"net/http"
	"net/url"
	"strings"
)

const (
	// ErrConsistentHashNextServerUnsupported is returned if NextServer is called
	ErrConsistentHashNextServerUnsupported = Error("Consistent Hash Pools don't support NextServer")

	// ErrConsistentHashInvalidSource is returned the source is not one of "request", "header", or "cookie"
	ErrConsistentHashInvalidSource = Error("the source provided is not valid")
)

const (
	invalidSource hashKeySource = iota
	requestSource
	headerSource
	cookieSource
)

type hashKeySource int

// ConsistentHashPool is a PoolManager that implements a consistent hash on a key to return
// the proper member consistently
type ConsistentHashPool struct {
	conhash       *consistent.Consistent
	hashKey       string
	hashKeySource hashKeySource
	pool          *Pool
	next          http.Handler
}

// NewConsistentHashPool returns a primed ConsistentHashPool
func NewConsistentHashPool(source, key string, pool *Pool, next http.Handler) (*ConsistentHashPool, error) {
	return NewConsistentHashPoolOpts(source, key, 7, 20, 1.25, pool, next)
}

// NewConsistentHashPoolOpts exposes some internal tunables, but still returns a ConsistentHashPool
func NewConsistentHashPoolOpts(source, key string, partitionCount, replicationFactor int, load float64, pool *Pool, next http.Handler) (*ConsistentHashPool, error) {

	hSource := stringToHashKeySource(source)
	if hSource == invalidSource {
		return nil, ErrConsistentHashInvalidSource
	}

	cfg := consistent.Config{
		PartitionCount:    partitionCount,
		ReplicationFactor: replicationFactor,
		Load:              load,
		Hasher:            hasher{},
	}
	c := consistent.New(nil, cfg)

	chp := ConsistentHashPool{
		conhash:       c,
		hashKey:       key,
		pool:          pool,
		next:          next,
		hashKeySource: hSource,
	}

	return &chp, nil
}

// String returns the Address of the Member
func (m *Member) String() string {
	return m.URL.String()
}

// Servers returns a list of member URLs
func (ch *ConsistentHashPool) Servers() []*url.URL {
	ml := ch.conhash.GetMembers()
	sl := make([]*url.URL, len(ml))
	for i, m := range ml {
		sl[i] = m.(*Member).URL
	}
	return sl
}

// ServeHTTP handles its part of the request
func (ch *ConsistentHashPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// make shallow copy of request before chaning anything to avoid side effects
	newReq := *req

	//hk := getHashKeyFromReq(ch.hashKey, &newReq)
	//m := ch.conhash.LocateKey(hk)
	m := ch.conhash.LocateKey(getHashKeyFromReq(ch.hashKey, ch.hashKeySource, &newReq))

	newReq.URL = m.(*Member).URL

	//DebugOut.Printf("CH Chose: %s for %s using %v!\n", newReq.URL.String(), req.URL.String(), hk)
	//DebugOut.Printf("\t%+v ... %+v\n", req.URL, newReq.URL)

	ch.next.ServeHTTP(w, &newReq)
}

// ServerWeight is a noop to implement PoolManager
func (ch *ConsistentHashPool) ServerWeight(u *url.URL) (int, bool) {
	return 0, false
}

// RemoveServer removes the specified member from the pool
func (ch *ConsistentHashPool) RemoveServer(u *url.URL) error {
	ch.conhash.Remove(u.String())
	return nil
}

// UpsertServer adds or updates the member to the pool
func (ch *ConsistentHashPool) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	var m *Member
	if ch.pool != nil {
		m = ch.pool.GetMember(u)
	} else {
		m = &Member{
			URL: u,
		}
	}
	ch.conhash.Add(m)
	return nil
}

// NextServer is an error-causing noop to implement PoolManager
func (ch *ConsistentHashPool) NextServer() (*url.URL, error) {
	return nil, ErrConsistentHashNextServerUnsupported
}

// Next returns the specified next Handler
func (ch *ConsistentHashPool) Next() http.Handler {
	return ch.next
}

func stringToHashKeySource(source string) hashKeySource {
	lsource := strings.ToLower(source)
	switch lsource {
	case "request":
		return requestSource
	case "header":
		return headerSource
	case "cookie":
		return cookieSource
	default:
		return invalidSource
	}
}

// getHashKeyFromReq follows the hashKey rules to return the proper []byte
func getHashKeyFromReq(key string, source hashKeySource, req *http.Request) []byte {
	lkey := strings.ToLower(key)

	switch source {
	case requestSource:
		if lkey == "remoteaddr" {
			return []byte(req.RemoteAddr)
		} else if lkey == "url" {
			return []byte(req.URL.String())
		} else if lkey == "host" {
			return []byte(req.Host)
		}
	case headerSource:
		return []byte(req.Header.Get(key))
	case cookieSource:
		cookie, err := req.Cookie(key)
		if err != nil {
			break
		}
		return []byte(cookie.Value)
	}

	return []byte("")
}

// consistent package doesn't provide a default hashing function.
// You should provide a proper one to distribute keys/members uniformly.
type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	// you should use a proper hash function for uniformity.
	return xxhash.Sum64(data)
}
