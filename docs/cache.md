# Cache

JAR includes a [groupcache-based](https://github.com/mailgun/groupcache/) content-caching mechanism, that allows groups of JAR servers to share the cache burden, using consistent-hashing and other intelligence to maintain high-availability of the cache and backfill accordingly. 

In the event one does not want a distributed cache, configuring oneself as the only peer is supported.

## Roadmap

**CAUTION:** This config format is not stable. Specifically the global `groupcache` configs are in flux, with an eye for minimizing differentiation of configuration between member instances.

## Configuration

```yaml
-
    Path: /pictures/
    CacheName: whatever
    hmacsigned: true
    Options:
      cache.sizemb: 128
      cache.maxitemsize: 10485760
      cache.expiration: 24h
      cache.controlheader: private
      hmac.key: abcdefghijk123
      hmac.expiration: 168h
    Pool: picservers 

groupcache:
  peerlist: http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082
  listenaddress: :8080
```

### Global

#### groupcache.listenaddress: [IP:PORT or :PORT]

Address on which groupcache should listen on.

#### groupcache.peerlist: [comma-delimited list of base URLs]

A list of peers ***the first of which being ourself!!!***, ala `http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082`

### Path

#### CacheName: [string]

Unique name for the cache of this path. 

At this time, you cannot share a cache between paths. That may be supported in the future.

### Per-Path Options

#### cache.controlheader: [string value]

The value to append to a `Cache-Control` response header. Default is to omit the header.

#### cache.expiration: [duration]

The duration in which an item should live in the cache. Default is "until evicted".

#### cache.maxitemsize: [size (bytes)]

The size - in bytes - at which an item will not be cached.

#### cache.sizemb: [size (megabytes)]

**Default: 16**
The size- in megabytes- the cache on this instance should be.
