package jar

import (
	"github.com/cognusion/go-jar/cache"
	"github.com/cognusion/go-prw"
	"github.com/cognusion/go-recyclable"
	"github.com/cognusion/go-timings"
	"github.com/pquerna/cachecontrol/cacheobject"

	"encoding/gob"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// ConfigGroupCachePeers is a config for a list of peers, ala ``http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082``
	ConfigGroupCachePeers = ConfigKey("groupcache.peerlist")
	// ConfigGroupCacheAddr is a config key for my listening address, ala ``:8080``
	ConfigGroupCacheAddr = ConfigKey("groupcache.listenaddress")

	// ConfigCacheSizeMB is the size- in megabytes- the cache should be. Defaults to 16.
	ConfigCacheSizeMB = ConfigKey("cache.sizemb")
	// ConfigCacheMaxItemSizeB is the size - in bytes - at which an item will not be cached.
	ConfigCacheMaxItemSizeB = ConfigKey("cache.maxitemsize")
	// ConfigCacheExpiration is the duration in which an item should live in the cache. Default is "until evicted".
	ConfigCacheExpiration = ConfigKey("cache.expiration")
	// ConfigCacheControlHeader is the value to append to a `Cache-Control` response header. Default is to omit the header.
	ConfigCacheControlHeader = ConfigKey("cache.controlheader")

	// NoCacheDefinedError is returned when a Path has CacheName set, but no cache has been globally configured.
	NoCacheDefinedError = Error("path requests a cache, but no cache defined")
	// CacheAlreadyDefinedError is returned when a Path has CacheName set, but that CacheName has already been used.
	CacheAlreadyDefinedError = Error("path requests a new cache that already exists")
)

// CacheCluster is our internal representation of GroupCache
type CacheCluster struct {
	*cache.GroupCache

	config  cache.Config
	addLock sync.Mutex
}

// NewCacheCluster should be called at most once, and returns an initialized CacheCluster
func NewCacheCluster(address string, peers []string) *CacheCluster {
	// TODO Should we defend against multiple calls?
	DebugOut.Printf("Adding cache cluster. Listening on %s with peers: %+v\n", address, peers)
	conf := cache.Config{
		ListenAddress: address,
		PeerList:      peers,
	}
	c, _ := cache.NewGroupCache(conf, nil) // No error possible if fillfunc is nil
	return &CacheCluster{
		GroupCache: c,
		config:     conf,
	}
}

// PageCache is a cache that is specific to caching responses
type PageCache struct {
	Name               string
	CacheSize          int64
	MaxItemSize        int64 // MaxItemSize > 0 is *roughly* the largest "page" "body" that will be cached
	ItemExpiration     time.Duration
	cluster            *CacheCluster
	cacheControlHeader string
	syncCacheIt        bool
}

// NewPageCache should be called at most once per unique "name", and returns an initialized PageCache
func (cc *CacheCluster) NewPageCache(name string, cacheSize, maxItemSize int64, itemExpiration time.Duration, cacheControlHeader string) (*PageCache, error) {

	// defend against multiple calls to the same cache
	cc.addLock.Lock()
	defer cc.addLock.Unlock()
	for _, cacheName := range cc.GroupCache.Names() {
		if name == cacheName {
			return nil, CacheAlreadyDefinedError
		}
	}

	nc := cc.config
	nc.Name = name
	nc.CacheSize = cacheSize
	nc.ItemExpiration = itemExpiration
	DebugOut.Printf("Adding PageCache: %+v\n", nc)
	err := cc.Add(nc, nil)
	if err != nil {
		return nil, err
	}

	return &PageCache{
		Name:               name,
		CacheSize:          cacheSize,
		MaxItemSize:        maxItemSize,
		ItemExpiration:     itemExpiration,
		cluster:            cc,
		cacheControlHeader: cacheControlHeader,
	}, nil
}

// Handler is a JAR Handler that returns the cached response or waits until the response is
// returned and caches it if appropriate.
func (c *PageCache) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rw, _ := prw.NewPluggableResponseWriterIfNot(w)
		defer rw.Flush()

		var requestID string
		if rid := r.Context().Value(requestIDKey); rid != nil {
			requestID = rid.(string)
		}

		// Timings
		t := timings.Tracker{}
		t.Start()

		saneURL := r.Host + url.PathEscape(r.URL.String())

		//DebugOut.Printf("Cache %s Handler for %+v\n", c.Name, r)
		// Check for a cache hit
		if v, ok := c.cluster.Get(c.Name, saneURL); ok {
			defer TimingOut.Printf("{%s} CacheHandler %s (hit) took %s\n", requestID, c.Name, t.Since().String())

			hit := RecyclableBufferPool.Get()
			defer hit.Close()
			hit.Reset(v.([]byte))

			dec := gob.NewDecoder(hit)
			if err := dec.Decode(&rw); err != nil {
				ErrorOut.Printf("{%s} Requested cache %s gob decoding went sideways: %+v\n", requestID, c.Name, err)
			}
			if c.cacheControlHeader != "" {
				rw.Header().Add("Cache-Control", c.cacheControlHeader)
			}
			return
		}
		// Post: Cache Miss
		TimingOut.Printf("{%s} CacheHandler %s (miss) took %s\n", requestID, c.Name, t.Since().String())

		next.ServeHTTP(rw, r)
		//Post: All downlevel handlers, finishers etc. have completed, and we're bubbling back up the response chain

		// Is item larger than we want to cache?
		if c.MaxItemSize > 0 && int64(rw.Length()) >= c.MaxItemSize {
			return
		}

		// Is this something we want to cache?
		if isCacheable(r.Header, r.Method, rw.Header(), rw.Code()) {
			if c.cacheControlHeader != "" {
				rw.Header().Add("Cache-Control", c.cacheControlHeader)
			}

			// Cache it
			buff := RecyclableBufferPool.Get()
			buff.Reset([]byte{})
			enc := gob.NewEncoder(buff)
			if err := enc.Encode(rw); err != nil {
				ErrorOut.Printf("{%s} Requested cache %s gob encoding went sideways: %+v\n", requestID, c.Name, err)
				buff.Close()
				return
			}
			DebugOut.Printf("{%s} Caching %s\n", requestID, saneURL)

			// ... async
			rchan := cacheIt(c, requestID, saneURL, buff)
			if c.syncCacheIt {
				<-rchan
			}
		}
	}
	return http.HandlerFunc(fn)
}

func cacheIt(c *PageCache, requestID, key string, buff *recyclable.Buffer) <-chan bool {
	rchan := make(chan bool, 1)

	go func() {
		defer buff.Close()

		var err error
		if c.ItemExpiration == 0 {
			err = c.cluster.Set(c.Name, key, buff.Bytes())
		} else {
			err = c.cluster.SetToExpireAt(c.Name, key, time.Now().Add(c.ItemExpiration), buff.Bytes())
		}

		if err != nil {
			switch err {
			case cache.CacheNotFoundError:
				ErrorOut.Printf("{%s} Requested cache %s doesn't exist", requestID, c.Name)
			default:
				ErrorOut.Printf("{%s} Requested cache %s Set went sideways: %+v", requestID, c.Name, err)
			}
		}
		rchan <- true
	}()

	return rchan
}

// isCacheable is a simplifier for isCacheableReasons
func isCacheable(reqHeader http.Header, reqMethod string, resHeader http.Header, resStatusCode int) bool {
	ok, _ := isCacheableReasons(reqHeader, reqMethod, resHeader, resStatusCode)
	return ok
}

// isCacheableReasons parses request Headers and Method, response Headers and StatusCode, and determines if the response is cacheables given
// RFC-correctness. If `false` is returned, the *[]Reason will be non-nil
func isCacheableReasons(reqHeader http.Header, reqMethod string, resHeader http.Header, resStatusCode int) (bool, *[]cacheobject.Reason) {
	reqDir, _ := cacheobject.ParseRequestCacheControl(reqHeader.Get("Cache-Control"))
	resDir, _ := cacheobject.ParseResponseCacheControl(resHeader.Get("Cache-Control"))
	expiresHeader, _ := http.ParseTime(resHeader.Get("Expires"))
	dateHeader, _ := http.ParseTime(resHeader.Get("Date"))
	lastModifiedHeader, _ := http.ParseTime(resHeader.Get("Last-Modified"))

	obj := cacheobject.Object{
		RespDirectives:         resDir,
		RespHeaders:            resHeader,
		RespStatusCode:         resStatusCode,
		RespExpiresHeader:      expiresHeader,
		RespDateHeader:         dateHeader,
		RespLastModifiedHeader: lastModifiedHeader,

		ReqDirectives: reqDir,
		ReqHeaders:    reqHeader,
		ReqMethod:     reqMethod,

		NowUTC: time.Now().UTC(),
	}
	rv := cacheobject.ObjectResults{}

	cacheobject.CachableObject(&obj, &rv)
	cacheobject.ExpirationObject(&obj, &rv)

	// This used to be clever:
	// return len(rv.OutReasons) == 0, &rv.OutReasons
	// but leaking an empty array from inside rv seemed gross.
	// bench did not jitter at all
	if len(rv.OutReasons) == 0 {
		return true, nil
	}
	return false, &rv.OutReasons
}
