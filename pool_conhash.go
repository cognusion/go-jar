package jar

import (
	"bytes"

	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
	"github.com/vulcand/oxy/v2/roundrobin"

	"net/http"
	"net/url"
	"strings"
)

const (
	// ErrConsistentHashNextServerUnsupported is returned if NextServer is called
	ErrConsistentHashNextServerUnsupported = Error("Consistent Hash Pools don't support NextServer")

	// ErrConsistentHashInvalidSource is returned the source is not one of "request", "header", or "cookie"
	ErrConsistentHashInvalidSource = Error("the consistent hash source provided is not valid")

	// ErrConsistentHashSourceNameImbalance is returned when the configured lists are not of the same lengths
	ErrConsistentHashSourceNameImbalance = Error("ConsistentHashSources and ConsistentHashNames are not balanced lists")
)

const (
	invalidSource hashKeySource = iota
	requestSource
	headerSource
	cookieSource
)

func init() {
	ConfigAdditions[ConfigPoolsDefaultConsistentHashPartitions] = 7
	ConfigAdditions[ConfigPoolsDefaultConsistentHashReplicationFactor] = 20
	ConfigAdditions[ConfigPoolsDefaultConsistentHashLoad] = 1.25
}

type hashKeySource int

// materializeConsistent extends Pool to be able to create ConsistentHashPools
func (p *Pool) materializeConsistent(next http.Handler) (PoolManager, error) {
	hashSources, err := makeHashSources(p.Config.ConsistentHashSources, p.Config.ConsistentHashNames)
	if err != nil {
		return nil, err
	}

	DebugOut.Printf("\t\tConsistentHash with '%+v'\n", hashSources)

	// Set defaults
	partitions := Conf.GetInt(ConfigPoolsDefaultConsistentHashPartitions)
	replication := Conf.GetInt(ConfigPoolsDefaultConsistentHashReplicationFactor)
	load := Conf.GetFloat64(ConfigPoolsDefaultConsistentHashLoad)

	// Allow overrides via PoolOptions :(
	if v := p.Config.Options.GetInt(ConfigConsistentHashPartitions); v != -1 {
		partitions = v
	}

	if v := p.Config.Options.GetInt(ConfigConsistentHashReplications); v != -1 {
		replication = v
	}

	if v := p.Config.Options.GetFloat64(ConfigConsistentHashLoad); v != -1 {
		load = v
	}

	return NewConsistentHashPoolOpts(hashSources, partitions, replication, load, p, next)
}

type hashSource struct {
	Source hashKeySource
	Key    string
}

// makeHashSources takes lists of source and name strings, and returns a list of hashSources or an error
// if a source is not valid or the lists are imbalanced.
func makeHashSources(sources, names []string) ([]hashSource, error) {
	if len(sources) < 1 {
		return nil, ErrConsistentHashInvalidSource
	} else if len(sources) != len(names) {
		return nil, ErrConsistentHashSourceNameImbalance
	}

	var hs = make([]hashSource, len(sources))
	for i, t := range sources {
		s := stringToHashKeySource(strings.TrimSpace(t))
		if s == invalidSource {
			return nil, ErrConsistentHashInvalidSource
		}
		hs[i] = hashSource{s, strings.TrimSpace(names[i])}
	}
	return hs, nil
}

// ConsistentHashPool is a PoolManager that implements a consistent hash on a key to return
// the proper member consistently
type ConsistentHashPool struct {
	conhash *consistent.Consistent
	sources []hashSource
	pool    *Pool
	next    http.Handler
}

// NewConsistentHashPool returns a primed ConsistentHashPool
func NewConsistentHashPool(source, key string, pool *Pool, next http.Handler) (*ConsistentHashPool, error) {
	sourceKeys, err := makeHashSources([]string{source}, []string{key})
	if err != nil {
		return nil, err
	}
	return NewConsistentHashPoolOpts(sourceKeys, 7, 20, 1.25, pool, next)
}

// NewConsistentHashPoolOpts exposes some internal tunables, but still returns a ConsistentHashPool
func NewConsistentHashPoolOpts(sourceKeys []hashSource, partitionCount, replicationFactor int, load float64, pool *Pool, next http.Handler) (*ConsistentHashPool, error) {

	cfg := consistent.Config{
		PartitionCount:    partitionCount,
		ReplicationFactor: replicationFactor,
		Load:              load,
		Hasher:            hasher{},
	}
	c := consistent.New(nil, cfg)

	chp := ConsistentHashPool{
		conhash: c,
		sources: sourceKeys,
		pool:    pool,
		next:    next,
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
func (ch *ConsistentHashPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// make shallow copy of request
	newReq := *r

	b := getAllHashKeysFromReq(ch.sources, &newReq)
	DebugOut.Printf("CH: %s\n", string(b))
	m := ch.conhash.LocateKey(b)
	if m == nil {
		// frick, pool is probably empty
		RequestErrorResponse(r, w, "Pool faulted, and likely is empty", http.StatusServiceUnavailable)
		return
	}
	newReq.URL = m.(*Member).URL

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
		// We have a pool, so let it render a Member for us
		m = ch.pool.GetMember(u)
	} else {
		// Render a trivial Member
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

// getAllHashKeysFromReq concatenates all of the values to return the proper []byte
func getAllHashKeysFromReq(sourceKeys []hashSource, req *http.Request) []byte {
	var b bytes.Buffer
	for _, v := range sourceKeys {
		b.Write(getHashKeyFromReq(v.Key, v.Source, req))
	}
	return b.Bytes()
}

// getHashKeyFromReq follows the hashKey rules to return the proper []byte
func getHashKeyFromReq(key string, source hashKeySource, req *http.Request) []byte {
	lkey := strings.ToLower(key)
	switch source {
	case requestSource:
		if lkey == "remoteaddr" {
			return []byte(strings.SplitN(req.RemoteAddr, ":", 2)[0])
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
