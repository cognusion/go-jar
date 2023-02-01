

# cache
`import "github.com/cognusion/go-jar/cache"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [type BackFillFunc](#BackFillFunc)
* [type Config](#Config)
* [type Error](#Error)
  * [func (e Error) Error() string](#Error.Error)
* [type GroupCache](#GroupCache)
  * [func NewGroupCache(config Config, fillfunc BackFillFunc) (*GroupCache, error)](#NewGroupCache)
  * [func (gc *GroupCache) Add(config Config, fillfunc BackFillFunc) error](#GroupCache.Add)
  * [func (gc *GroupCache) Close() error](#GroupCache.Close)
  * [func (gc *GroupCache) Get(cacheName, key string) (value interface{}, ok bool)](#GroupCache.Get)
  * [func (gc *GroupCache) GetContext(ctx context.Context, cacheName, key string) (value interface{}, ok bool)](#GroupCache.GetContext)
  * [func (gc *GroupCache) Names() []string](#GroupCache.Names)
  * [func (gc *GroupCache) Remove(cacheName, key string) error](#GroupCache.Remove)
  * [func (gc *GroupCache) RemoveContext(ctx context.Context, cacheName, key string) error](#GroupCache.RemoveContext)
  * [func (gc *GroupCache) Set(cacheName, key string, value []byte) error](#GroupCache.Set)
  * [func (gc *GroupCache) SetContext(ctx context.Context, cacheName, key string, value []byte, expiration time.Time) error](#GroupCache.SetContext)
  * [func (gc *GroupCache) SetDebugOut(logger *log.Logger)](#GroupCache.SetDebugOut)
  * [func (gc *GroupCache) SetPeers(peers ...string)](#GroupCache.SetPeers)
  * [func (gc *GroupCache) SetToExpireAt(cacheName, key string, expireAt time.Time, value []byte) error](#GroupCache.SetToExpireAt)
  * [func (gc *GroupCache) Stats(w http.ResponseWriter, req *http.Request)](#GroupCache.Stats)


#### <a name="pkg-files">Package files</a>
[group.go](https://github.com/cognusion/go-jar/tree/master/cache/group.go) [misc.go](https://github.com/cognusion/go-jar/tree/master/cache/misc.go)


## <a name="pkg-constants">Constants</a>
``` go
const (
    // NilBackfillError is returned by the Getter if there there is no backfill func, in lieu of panicing
    NilBackfillError = Error("item not in cache and backfill func is nil")
    // ItemNotFoundError is a generic error returned by a BackFillFunc if the item is not found or findable
    ItemNotFoundError = Error("item not found")
    // CacheNotFoundError is an error returned if the cache requested is not found
    CacheNotFoundError = Error("cache not found")
    // NameRequiredError is returned when creating or adding a cache, and the Config.Name field is empty
    NameRequiredError = Error("name is required")
)
```




## <a name="BackFillFunc">type</a> [BackFillFunc](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=831:881#L27)
``` go
type BackFillFunc func(key string) ([]byte, error)
```
BackFillFunc is a function that can retrieve an uncached item to go into the cache










## <a name="Config">type</a> [Config](https://github.com/cognusion/go-jar/tree/master/cache/misc.go?s=261:771#L16)
``` go
type Config struct {
    Name           string        // For New and Add. Pass as ``cacheName`` to differentiate caches
    ListenAddress  string        // Only for New to set the listener
    PeerList       []string      // Only for New to establish the initial PeerList. May be reset with GroupCache.SetPeers()
    CacheSize      int64         // For New and Add to set the size in bytes of the cache
    ItemExpiration time.Duration // For New and Add to set the default expiration duration. Leave as empty for infinite.
}

```
Config is used to store configuration information to pass to a GroupCache.










## <a name="Error">type</a> [Error](https://github.com/cognusion/go-jar/tree/master/cache/misc.go?s=61:78#L8)
``` go
type Error string
```
Error is an error type










### <a name="Error.Error">func</a> (Error) [Error](https://github.com/cognusion/go-jar/tree/master/cache/misc.go?s=130:159#L11)
``` go
func (e Error) Error() string
```
Error returns the stringified version of Error




## <a name="GroupCache">type</a> [GroupCache](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=1087:1295#L31)
``` go
type GroupCache struct {
    // contains filtered or unexported fields
}

```
GroupCache is a distributed LRU cache where consistent hashing on keynames is used to cut out
"who's on first" nonsense, and backfills are linearly distributed to mitigate multiple-member requests.







### <a name="NewGroupCache">func</a> [NewGroupCache](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=1492:1569#L44)
``` go
func NewGroupCache(config Config, fillfunc BackFillFunc) (*GroupCache, error)
```
NewGroupCache creates a GroupCache from the Config. Only call this once. If you need
more caches use the .Add() function. fillfunc may be nil if caches will be added later
using .Add().





### <a name="GroupCache.Add">func</a> (\*GroupCache) [Add](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=2344:2413#L81)
``` go
func (gc *GroupCache) Add(config Config, fillfunc BackFillFunc) error
```
Add creates new caches in the cluster. Config.ListenAddress and Config.PeerList are ignored.




### <a name="GroupCache.Close">func</a> (\*GroupCache) [Close](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=3278:3313#L124)
``` go
func (gc *GroupCache) Close() error
```
Close calls the listener close function




### <a name="GroupCache.Get">func</a> (\*GroupCache) [Get](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=3450:3527#L130)
``` go
func (gc *GroupCache) Get(cacheName, key string) (value interface{}, ok bool)
```
Get will return the value of the cacheName'd key, asking other cache members or
backfilling as necessary.




### <a name="GroupCache.GetContext">func</a> (\*GroupCache) [GetContext](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=3743:3848#L136)
``` go
func (gc *GroupCache) GetContext(ctx context.Context, cacheName, key string) (value interface{}, ok bool)
```
GetContext will return the value of the cacheName'd key, asking other cache members or
backfilling as necessary, honoring the provided context.




### <a name="GroupCache.Names">func</a> (\*GroupCache) [Names](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=3032:3070#L110)
``` go
func (gc *GroupCache) Names() []string
```
Names returns the names of the current caches




### <a name="GroupCache.Remove">func</a> (\*GroupCache) [Remove](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=6157:6214#L187)
``` go
func (gc *GroupCache) Remove(cacheName, key string) error
```
Remove makes a best effort to remove an item from the cache




### <a name="GroupCache.RemoveContext">func</a> (\*GroupCache) [RemoveContext](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=6385:6470#L192)
``` go
func (gc *GroupCache) RemoveContext(ctx context.Context, cacheName, key string) error
```
RemoveContext makes a best effort to remove an item from the cache, honoring the provided context.




### <a name="GroupCache.Set">func</a> (\*GroupCache) [Set](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=4366:4434#L155)
``` go
func (gc *GroupCache) Set(cacheName, key string, value []byte) error
```
Set forces an item into the cache, following the configured expiration policy




### <a name="GroupCache.SetContext">func</a> (\*GroupCache) [SetContext](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=4721:4839#L161)
``` go
func (gc *GroupCache) SetContext(ctx context.Context, cacheName, key string, value []byte, expiration time.Time) error
```
SetContext forces an item into the cache, following the specified expiration (unless a zero Time is provided
then falling back to the configured expiration policy) honoring the provided context.




### <a name="GroupCache.SetDebugOut">func</a> (\*GroupCache) [SetDebugOut](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=6704:6757#L201)
``` go
func (gc *GroupCache) SetDebugOut(logger *log.Logger)
```
SetDebugOut wires in the debug logger to the specified logger




### <a name="GroupCache.SetPeers">func</a> (\*GroupCache) [SetPeers](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=6844:6891#L206)
``` go
func (gc *GroupCache) SetPeers(peers ...string)
```
SetPeers allows the dynamic [re]setting of the peerlist




### <a name="GroupCache.SetToExpireAt">func</a> (\*GroupCache) [SetToExpireAt](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=5841:5939#L181)
``` go
func (gc *GroupCache) SetToExpireAt(cacheName, key string, expireAt time.Time, value []byte) error
```
SetToExpireAt forces an item into the cache, to expire at a specific time regardless of the cache configuration. Use
SetContext if you need to set the expiration and a context.




### <a name="GroupCache.Stats">func</a> (\*GroupCache) [Stats](https://github.com/cognusion/go-jar/tree/master/cache/group.go?s=6993:7062#L211)
``` go
func (gc *GroupCache) Stats(w http.ResponseWriter, req *http.Request)
```
Stats is a request finisher that outputs the GroupCache stats as JSON








- - -
Generated by [godoc2md](http://godoc.org/github.com/cognusion/godoc2md)
