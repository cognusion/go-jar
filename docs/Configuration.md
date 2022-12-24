# Configuration

JAR configuration can occur via file or ENV. I recommend files, and this documentation assumes that (*YAML* specifically). JAR can read configuration in JSON, TOML, YAML, HCL, or Java properties -formats. The config variable *names* are case insensitive, but be careful with the *values* as most are strict.

This documentation reflects configuration possibilities as of the associated commit. Please extend this documentation, or ask questions, as they come up.

```bash
Usage of ./jar:
      --checkconfig     Run through the config load and then exit
      --config string   Config file to load
      --debug           Enable vociferous output
      --docs            Run through the config, build runtime docs, and then exit
      --dumpconfig      Load the config, dump it to stderr, and then exit
      --version         Print the version (3.24.4), and then exit
```

## General

### authoritativedomains: [list]

A list of domains or hostname suffixes this instance will handle. Anything not matching will return 400. This also impact **tls.httpredirects**.

### compression: [list]

A list of MIME types that are eligible for wire-time compression if the client requests it

```yaml
compression:
  - text/xml
  - text/css
  - text/javascript
  - application/javascript
```

### config: [configfile]

Only available via CLI or ENV, points to the configuration file to load. Configuration in JSON, TOML, YAML, HCL, or Java properties -formats is allowed.

### debug: [true|false]

**Default: false**
Also available via CLI as **--debug** this will enable debug logging (see **debuglog** for more) and possibly open additional diagnostic codepaths including spawning goroutines for monitoring and emitting

### debugrequests: [true|false]

**Default: false**
Only enabled if **debug: true** is set.
Enables debug logging of full requests (POST/PUT bodies omitted)

### debugresponses: [true|false]

**Default: false**
Only enabled if **debug: true** is set.
Enables debug logging of full responses (response bodies omitted, generally)

### debugtimings: [true|false] *(Caution)*

**Default: false**
Only enabled if **debug: true** is set.
Enables debug logging of various function timings. Very noisy. **May be removed at any time**.

### ec2: [true|false]

**Default: false**
Enables AWS EC2 awareness.

### fakexfflog: [true|false]

**Default: false**
Backfills the X-Forwarded-For field in the access log, if it isn't already occupied.

### forbiddenpaths: [list]

List of regular expressions (PCRE) that if matched against a URI path, will result in a 403. Barewords will match anywhere in the path. Remember to escape your fore-slashes. *Case-insensitive*.  If you want to limit these to only certain **Paths**, see **Path.ForbiddenPaths**

```yaml
forbiddenpaths:
  - ^\/supersecret\/
  - ^\/(?!admin\/).*healthcheck
  - omgnotthis
```

### handlers: [list]

Handlers listed here will be prepended to the **handlers** list of *all* Paths.

```yaml
handlers:
  - Recoverer
  - ErrorHandler
```

### headers: [list]

Headers listed here will be added to all outgoing responses, unless there is only a name and no value, in which case the header will be *removed* from all responses. Header names are assumed to be everything *left* of the first space. Values are assumed to be everything *right* of the first space. The **headers** list fully allows for **Macros** in the *values*.

```yaml
headers:
  # Set HSTS for TLS
  - "Strict-Transport-Security max-age=31536000"
  # Set the X-JAR header to a macro
  - "X-JAR %%FULLVERSION"
  # Strip the "Server" header from all responses
  - "Server"
```

### keepalivetimeout: [duration]

**Default: 5s**
Specifies the amount of time to allow a keptalive socket to linger.

### listen: [address]

**Default: :8080**
Specifies the *address:port* or *:port* to listen for HTTP requests on.

**NOTE:** Changing this value while running will create problems for graceful restarts, and should only be done during controlled stop/starts.

### macros: [key/value list]

In various configuration stanzas you may be allowed to use macros to replace text with dynamic values. Macros are **always** uppercase, and prefixed with two (2) percent-signs (%%) when used for expansion. See [Writing Handlers & Finishers](WritingHandlersAndFinishers.md) for information on how to use these in your own code.

Macros may be added or redefined at start time, via configuration. Macro values may contain other macros, but **only one level of macros will be expanded** (poor man's infinite-recursion avoidance).

As a rule, don't change a macro that defines a *value* to a macro containing another macro.

```yaml
macros:
  NAME: myserver
  CONFIGVERSION: 1.0
  SHORTVERSION: %%NAME/%%VERSION
  # The above works because all macros contained are defined
  # The below will not expand properly as it contains a macro that contains other macros
  FULLVERSION_BUSTED: %%SHORTVERSION %%CONFIGVERSION
  # The below works because all macros contained are defined
  FULLVERSION_WORKS: %%NAME/%%VERSION %%CONFIGVERSION
```

#### Defaults

- NAME: "JAR"
- VERSION: [JAR Version]
- SHORTVERSION: NAME/VERSION
- FULLVERSION: NAME/VERSION GoVersion MaxGoProcesses:NumberOfCPUs

### maxconnections: [number]

**Default: 0**
If 0, incoming connections to the listener are limited by the operating system, when it runs out of handles or ports. If > 0, provides a hard limit to the number of simultaneously accepted connections.

### requestidheadername: [header]

**Default: X-Requestid**
***Mandatory***
Sets the name of the request header to set with the request ID.

### slowrequestmax: [duration]

**Default: unset**
If set, requests taking longer than this will be logged to **slowlog**.

### tempfolder: [path]

**Default: /tmp**
Specifies the folder path where any temporary files should be stored. Temporary files will be put into sub-directories, and removed immediately after their useful life has ended, unless catastrophic failure occurs.

### timeout: [duration]

**Default: 0**
If 0, there is no timeout. *Ever*. If > 0, sets the global request/response timeout. A request lasting longer than this will be returned a **503 Service Unavailable**. It may be overridden on a per-Path basis via **Path.Timeout**.

### trustrequestidheader: [true|false]

**Default: false**
If set, will trust incoming request header **requestidheadername** to provide a proper requestId, otherwise will generate one and [over]write it.

### versionrequired: [version string]

**Default: [empty]**
Sets the minimum ``jar`` VERSION that the config is valid for. Fails on bootstrap if not met.

## Administration

### hotconfig: [true|false]

**Default: false**
If set, will trigger a graceful restart if the specified **config** file is modified.

### hotupdate: [true|false]

**Default: false**
If set, in conjunction with the **Update** Finisher, will allow a graceful restart onto new code after a successful update.

### updatepath: [url]

**Default: [empty]**
*WARNING:* Requires **ec2:true** or **keys** for AWS set in config/environment.
In conjunction with the **Update** Finisher, this specifies an *S3 URL* where a .zip containing an updated binary can be found (assumes running on an EC2 instance or in a credentialed environment with sufficient permission).

## TLS

JAR provides full support for TLS-fronted connections. Including transparently handling multiple certs/domains using SNI. As security standards have evolved, the minimum acceptable version is TLS v1.0 (RIP SSL).

### tls.certs [list of certs]

```yaml
 certs:
   "*.domain.com":
     keyfile: keys/domain.com/key.pem
     certfile: keys/domain.com/cert.pem
   "*.another.net":
     keyfile: keys/another.net/key.pem
     certfile: keys/anoter.net/cert.pem
```

### tls.ciphers [list]

**Default: [see example below for default]**
List of ciphersuites- ordered- that you want to support. The list of acceptable ciphersuites can be found in the Go [tls package documentation](https://golang.org/pkg/crypto/tls/#pkg-constants).

```yaml
 ciphers:
   - TLS_CHACHA20_POLY1305_SHA256
   - TLS_AES_256_GCM_SHA384
   - TLS_AES_128_GCM_SHA256
   - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
   - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
   - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
   - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
   - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
   - TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
   - TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
   - TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
   - TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
```

Available ciphers vary, but the superset is below (the provided hex ID is a reference to the RFC-assigned cipher suite ID)

```text
"TLS_RSA_WITH_RC4_128_SHA":                0x0005,
"TLS_RSA_WITH_3DES_EDE_CBC_SHA":           0x000a,
"TLS_RSA_WITH_AES_128_CBC_SHA":            0x002f,
"TLS_RSA_WITH_AES_256_CBC_SHA":            0x0035,
"TLS_RSA_WITH_AES_128_CBC_SHA256":         0x003c,
"TLS_RSA_WITH_AES_128_GCM_SHA256":         0x009c,
"TLS_RSA_WITH_AES_256_GCM_SHA384":         0x009d,
"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":        0xc007,
"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    0xc009,
"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    0xc00a,
"TLS_ECDHE_RSA_WITH_RC4_128_SHA":          0xc011,
"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":     0xc012,
"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      0xc013,
"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      0xc014,
"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256": 0xc023,
"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":   0xc027,
"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   0xc02f,
"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": 0xc02b,
"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   0xc030,
"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": 0xc02c,
"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    0xcca8,
"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  0xcca9,
"TLS_AES_128_GCM_SHA256":                  0x1301,
"TLS_AES_256_GCM_SHA384":                  0x1302,
"TLS_CHACHA20_POLY1305_SHA256":            0x1303,
```

### tls.enabled [true|false]

**Default: false**
If set, will process all of the TLS-related directives, building out a TLS config, and eventually using it. This will ignore **listen** in lieu of **tls.listen** (see **tls.httpredirects** below).

**NOTE:** Changing this value while running will create problems for graceful restarts, and should only be done during controlled stop/starts.

### tls.http2 [true|false] **EXPERIMENTAL**

**Default: false**
**EXPERIMENTAL**: If set, enables HTTP/2 protocol support, maybe. **EXPERIMENTAL**

### tls.httpredirects [true|false]

**Default: false**
Binds a simple HTTP listener to the **listen** address, and redirects all traffic there to the **tls.listen** address.

### tls.keepalivedisabled [true|false]

**Default: false**
If set, will disable keepalives when using TLS.

### tls.listen  [address]

**Default: :8443**
Specifies the *address:port* or *:port* to listen for HTTPS requests on.

**NOTE:** Changing this value while running will create problems for graceful restarts, and should only be done during controlled stop/starts.

### tls.maxversion [decimal version]

**Default: 1.3**
Acceptable values are (currently):

- 1.0
- 1.1
- 1.2
- 1.3

### tls.minversion [decimal version]

**Default: 1.2**
Acceptable values are (currently):

- 1.0
- 1.1
- 1.2
- 1.3

## Logging

### accesslog: [logfile]

**Default:  [standard out]**
Supplied **logfile** may be an absolute or relative file location to log structured "access log" information. The file need not exist, but the path must, and be writable by the executing user.

### commonlog: [logfile]

**Default: off**
Supplied **logfile** may be an absolute or relative file location to log structured "Apache common log" (combined-log style) information. The file need not exist, but the path must, and be writable by the executing user.

### debuglog: [logfile]

**Default:  [standard error]**
Only enabled if **debug: true** is set.
Supplied **logfile** may be an absolute or relative file location to log unstructured "debug log" information. The file need not exist, but the path must, and be writable by the executing user.

### errorlog: [logfile]

**Default: [standard error]**
Supplied **logfile** may be an absolute or relative file location to log unstructured "error log" information. The file need not exist, but the path must, and be writable by the executing user.

### logage: [days]

**Default: 28**
For any log logging to a file, specifies the rough maximum number of days to keep a rolled log.

### logbackups: [number]

**Default: 3**
For any log logging to a file, specifies the maximum number of rolled files to keep.

### logsize: [MBsize]

**Default: 100**
For any log logging to a file, specifies the rough maximum size (in megabytes) a log is allowed to get before rolling.

### slowlog: [logfile]

**Default: off**
Supplied **logfile** may be an absolute or relative file location to log requests that exceed **slowrequestmax**. The file need not exist, but the path must, and be writable by the executing user.

## Paths

A **Path** is a special thing to JAR. It is a simple structure that is akin to an Apache "Location", "VirtualHost", or "Dir" - all in one. Paths are *ordered* in the configuration, and when a request is being examined, the first matching Path is used to service the request.

```yaml
 paths:
   -
    Path: /nope
    Absolute: true
    Finisher: Forbidden
  -
    Path: /search
    Host: google.example.com
    Methods:
      - GET
      - POST
    Redirect: https://www.google.com%1
    Timeout: 60s
  -
    Path: /secure
    RateLimit: 100
    Allow: 192.168.0.0/16
    BasicAuthSource: files/secure.passwd
    BasicAuthRealm: Secure
    Pool: securePool
  -
    Path: /
    RateLimit: 100
    ForbiddenPaths:
      - ^\/secret
      - ^\/\d\d\d
    Handlers:
      - Recoverer
    Pool: default
```

### absolute: [true|false]

**Default: false**
If set, **path** is treated absolutely, instead of as a prefix.

### allow: [rules]

**Default: all**
Comma-delimited list of IP addresses/ranges to allow. Setting **allow** but not setting **deny** automatically flips **deny** to *all*. Any addresses in **allow** are evaluated before addresses in **deny**, except *all*.

```yaml
  -
    Path: /admin/restart
    Absolute: true
    Allow: 127.0.0.1
    # implicit Deny: all
    Finisher: Restart
```

### basicauthrealm: [string]

Name of the HTTP Auth Realm on this Path. Need not be unique. Should not be empty if basic auth is used

### basicauthsource: [url]

URL to specify where HTTP Basic Auth information should come from (file://). Setting this forces HTTP Basic Authentication for the Path.

### basicauthusers: [list]

**Default: all**
List of usernames allowed on this Path.

### bodybytelimit: [size in bytes]

If set, enables the BodyByteLimit handler on the path, enforcing any body is at most the specified size, or a *413 Request Entity too large* is returned.

### deny: [rules]

**Default: none**
Comma-delimited list of IP addresses/ranges to deny. Any addresses in **Allow** are evaluated before addresses in **Deny**.

```yaml
  -
    Path: /admin/update
    Deny: 137.143.0.0/16, 8.8.8.8
    Absolute: true
    Finisher: Update
```

### errorcode: [http code]

**Default: http.StatusOk (200)**
The HTTP response code that will be returned with ErrorMessage, IFF ErrorMessage is set.

### errormessage: [string]

**Default: none**
A static message to respond with, if this path is executed.

### finisher: [finisher name]

The name of a Finisher used to complete requests to this Path. Mutually exclusive to **pool** and **redirect**.

### forbiddenpaths: [list]

List of regular expressions (PCRE) that if matched against a URI path, will result in a 403. Barewords will match anywhere in the path. Remember to escape your fore-slashes. *Case-insensitive*. There is also the global **forbiddenpaths** config if that is more appropriate.

### handlers: [list]

This is a list of Handlers that will be applied to the Path, after any global handlers.

```yaml
Handlers:
  - Recoverer
  - Crashy
```

### headers: [list]

This is a list of request headers, in "name value" format, that will be used when matching the **Path**. They can be absolute, or use simple matching. Header names are assumed to be everything *left* of the first space. Values are assumed to be everything *right* of the first space.

### host: [hostname|pattern]

**Default: [any]**
This is a hostname or hostname pattern that will be used when matching the Path. It can be absolute, or use expressions.

```yaml
Host: www.domain.com
```

```yaml
 # Match all hosts starting with 'dev-'
 Host: "{_:|^dev-.*}.domain.com"
```

### hosts: [list of hostname|pattern]

This is a list of hostnames or hostname patterns that the Path is valid for. See **Path.Host** for examples.

### methods: [list]

**Default: [any]**
This is a list of HTTP methods that will be used when matching the Path.

```yaml
Methods:
  - GET
  - HEAD
```

### name: [path name]

The name of a Path is used only in debug output, and is completely optional. If it is not set, the debug output will contain the ordered index number of the Path instead.

### options: [PathOptions]

A free-form field in a Path, used by specific Handlers or Finishers to consume path-specific configuration. This is added to the request Context, so keep it light.

```yaml
Options:
   MirrorRequest.Mirrors:
     - Pool1
     - Pool2
```

### path: [path]

Path is a URI path, starting with a forward-slash (/) and possibly with more specificity thereafter. By default, the path is treated as a prefix, thus */he* would match */he*, */help*, */helloooooo*, etc. If this is undesirable, set **absolute**. Without other configuration, this path will match any hostname, any method, any request that contains the path.

### pool: [pool name]

The name of a Pool used to complete requests to this Path. Mutually exclusive to **Finisher** and **Redirect**.

### ratelimit: [decimal requests/second]

RateLimit sets the number of requests-per-second-per-source (IP:port) allowed.

```yaml
RateLimit: 100.175
```

### ratelimitcollectonly: [boolean]

**Default: false**
If set, ratelimit violations will only print to the debug log, and not respond differently.

### ratelimitpurge: [duration]

**Default: 1 hour**
This is how long a ratelimit bucket can exist before it is expired.

### redirect: [url]

A URL to redirect requests to this Path. Mutually exclusive to **finisher** and **pool**. A macro *%1* may be put on the URL to substitute the request path.

```yaml
Redirect: https://www.google.com%1
```

### redirectcode: [http code]

**Default: 301**
By default, **redirect** requests use *301 Permanent*, but that can be changed here, within reason.

### replacepath: [string]

Simple string replacement of this string, for the path string, before the request is proxied. No regexps.

```yaml
  -
    Path: /hello
    ReplacePath: /world
    Pool: default
```

### stripprefix: [prefix]

Removes a the specified string from the beginning of a URI path, before it is forwarded on. Useful when remapping e.g. */files/folder/thefile.html* to */folder/thefile.html*

### timeout: [duration]

**Default: 0**
If > 0, sets the request/response timeout for this Path, overriding the global **timeout**. A request lasting longer than this will be returned a **503 Service Unavailable**.

## Pools

Pools are containers for one or more service endpoints providing analogous services, that are proxied, load-balanced, etc. Pool members are proxied differently depending on their protocol scheme. Currently ``https://``, ``http://``, ``s3://``, and ``ws://`` are supported. Not all configuration options are supported by all pool types.

```yaml
pools.healthcheckinterval: 30s
stickycookie.aes.ttl: 5m
pools:
  api:
    Name: api
    Members:
      - http://192.168.0.10:8080
      - http://192.168.0.11:8080
  workers:
    Name: wsworkers
    Sticky: true
    StickyCookieName: ROUTEID
    StickyCookieType: aes
    Members:
      - ws://192.168.0.10:8081
      - ws://192.168.0.11:8081
  default:
    Name: www
    Sticky: true
    Members:
      - http://192.168.0.9
```

### pools.defaultconsistenthashload: [float]

**Default: 1.25**
From the library: Load is used to calculate average load. See the code, the paper and Google's blog post to learn about it.

### pools.defaultconsistenthashpartitions: [integer]

**Default: 7**
From the library: Keys are distributed among partitions. Prime numbers are good to distribute keys uniformly. Select a big PartitionCount if you have too many keys.

### pools.defaultconsistenthashreplicationfactor: [integer]

**Default: 20**
From the library: Members are replicated on consistent hash ring. This number controls the number each member is replicated on the ring.

### pools.defaultmembererrorstatus: [healthcheckstatus]

**Default: Warning**
The default HealthCheckStatus for members in an error state. One of Unknown, Ok, Warning, or Critical. Overridden per-Pool setting **HealthCheckErrorStatus**.

### pools.defaultmemberweight: [number]

**Default: 1**
The default weight for a Pool member.

### pools.healthcheckinterval: [interval]

**Default: 1 minute**
Global for all pools. If set to 0, disables automatic healthchecks, otherwise sets the frequency of the healthchecks. This is most accurately described as a "maximum interval", by default, unless **healthcheckshotgun** its set.
**NOTE:** Healthchecking many members doesn't work so well on small, single-CPU boxes. You've been warned.

### pools.healthcheckshotgun: [true/false]

**Default: false**
Global for all pools. If set to false (the default), healthchecks will be adaptively scheduled over the interval, so for an interval of 1 minute, if there are 60 healthchecks to run, one will be fired off every second throughout that minute. The order of scheduling is not maintained from interval-to-interval.
If set to true, all healthchecks will be run concurrently at the beginning of each interval.

### pools.localmemberweight: [number]

**Default: 1000**
The weight for a Pool member who is AZ-local to the JAR instance.

### pools.prematerialize: [true/false]

**Default: false**
If set, each Pool will be materialized during bootstrap, instead of as-requested. Pools generally materialize very quickly, but a materialized Pool takes up more
memory (and goros) than a husk, so unless all of your Pools are used all of the time, leaving this alone is just fine.

### stickycookie.aes.ttl: [duration] (*experimental*)

**Defalt: 0 (off)**
Global for all pools. If set, and a pool has **Sticky** set, and **StickyCookieType: aes**, will embed an expiration in the encrypted cookie and verify it upon receipt.
This *WILL NOT* set the cookie-level expiration. This *should* be used with **Buffered** and **BufferedFails** set to at least *2*.
**CAUTION:**  This option is tagged experimental and should be used with caution.

### stickycookie.httponly: [true/false]

**Default: false**
Global for all pools. If set, and a pool has **Sticky** set, the cookie will have the [HTTPOnly](https://www.owasp.org/index.php/HttpOnly) property set.

### stickycookie.secure: [true/false]

**Default: false**
Global for all pools. If set, and a pool has **Sticky** set, the cookie will have the [Secure](https://www.owasp.org/index.php/SecureFlag) property set.

### buffered: [true|false]

**Default: false**
If set, all outgoing requests will be buffered. In the event of a timeout or 500-class error, the request will be retried on other members up to **bufferedfails**-minus-1 times.
**NOTE:** Buffered requests *may* be retried on members that failed previously. Do not use this, if that is unacceptable.
**ANOTHER NOTE:** This is generally a bad idea unless you are confident in the idempotence of your requests (or rather the endpoint handling your requests).

### bufferedfails: [number]

If **buffered** is set, this is the number of times a request may fail before giving up.

### consistenthashing: [true|false]

**Default: false**
If set, consistent hashing will be used on the pool, ensuring consistency and uniform distribution across pool members.

### consistenthashname: [string]

If **consistenthashing** is set, this value will be the field whose value is used as a hash key.

### consistenthashsource: [header|cookie|request]

If **consistenthashing** is set, this value specifies where to pull the value, specified by **consistenthashname**, for the hash key. 
For "header" and "cookie", it is paired with **consistenthashname** to choose which key from those maps is used.
For "request" it is paired with **consistenthashname** to choose from one of "remoteaddr", "host", or "url".

### ec2affinity: [true|false]

**Default: false**
If set, and globally **ec2: true** then Pool member who are EC2 instances and in the same Availability Zone as the running JAR instance, will receive much higher
weight than other members.

### healthcheckdisabled: [true|false]

**Default: false**
If set, will disable healthchecks for members of this pool

### healthcheckerrorstatus: [healthcheckstatus]

**Default: pools.defaultmembererrorstatus**
The HealthCheckStatus for members in an error state. One of Unknown, Ok, Warning, or Critical.

### healthcheckuri: [uri]

**Default: "/"**
Set the URI used to healthcheck the member.

### members: [urls]

A list of URIs that will be added to the Pool. Pool members are proxied differently depending on their protocol scheme. Currently ``https://``, ``http://``, ``s3://``, and ``ws://`` are supported. The scheme of the first member listed determines the type of the Pool, and mixing membership types will generally not work.

```yaml
Members:
  - http://server1.example.com
  - http://server2.example.com
  - http://server3.example.com
```

### name: [name]

The unique name of the Pool. Will be referenced by Paths.

### prune: [true|false]

**Default: false**
If set, will remove members who are failing healthcheck, and add them back after they pass again.

### removeheaders: [list]

List of headers to specifically remove for this pool.

### replacepath: [path]

If set, and requested URI path to hit this pool, will be replaced with this.

### sticky: [true|false]

**Default: false**
If set, will try to pin the source to a specific member for subsequent requests, as long as the member remains available.
**NOTE:** Applications that require pinning are lazy, and are not in the best interest of your customers. There are numerous conditions when this cannot be honored and requests will go elsewhere, and your application should handle that with aplomb and not bomb.
**NOTE:** Set **stickycookietype: aes** and **keys.stickycookie** if you want "advanced obscurity". See **keys** below for more information.

### stickycookiename: [string]

**Default: "jar+Pool.Name"**
If set, will override the default sticky cookie name.

### stickycookietype: [plain|hash|aes]

**Default: plain**
If **sticky** is set, the following values are allowed:

- plain - Cookie values will be a cleartext representation of the pool member the requestor is pinned to, e.g. `http://127.0.0.1:8081/`
- hash - Cookie values will be hashed with **keys.stickycookie** as the salt. See **keys** below for more information.
- aes - Cookie values will be AES-encrypted with **keys.stickycookie**. See **keys** below for more information.

### stripprefix: [string]

**Default: none**
Removes the specified string from the front of a URL before processing. Dupes **Path.StripPrefix**

### timeout: [duration] *(stub)*

**Default: 0**
If set, will limit the amount of time a request to a pool member will be allowed to take. This will override any global **timeout** set.

## Workers

Workers are used by Handlers and Finishers, as well as some JAR subsystems (e.g. Pool member healthchecking). The number of Workers will automatically expand and contract based on the perceived amount of work, and the depth of the work queue. The defaults are quite sane, and it is not generally recommended to change them. Idle workers take up almost no CPU and very very little memory (stack), so the only reason to control the pool size is if you're encountering issues with too much Work being done simultaneously, e.g. on tiny instances.  It is also worth noting that Workers will not abandon work-in-progress, even if they've been asked to die off due to pool resizing.

```yaml
workers:
  queuesize: 100
  initialpoolsize: 10
  minpoolsize: 2
  maxpoolsize: 0
  resizeinterval: 30s
```

### initialpoolsize: [number]

**Default: 10**
Defines the number of Workers to start when the Workerpool is created. If **resizeinterval** is *0*, this is the permanent number of Workers in the pool.

### maxpoolsize: [number]

**Default: 0**
Defines the maximum number of Workers that are allowed. *0* allows indefinite growth. Generally speaking, this should not be set unless running on tiny systems and encountering other problems. If **resizeinterval** is *0* this has no purpose.

### minpoolsize: [number]

**Default: 2**
Defines the minimum number of Workers that are allowed. *0* disables shrinkage, such that once a Worker is created, it will never be organically terminated (*not recommended*). If **resizeinterval** is *0* this has no purpose.

### queuesize: [number]

**Default: 100**
This sets the size of the work queue. Too small of a number here, and Adding new work may be delayed causing Workers to be idle longer than necessary. The queue does consume a very small amount of memory, so too large is wasteful, albeit not debilitating (within reason).
Unless you're running out of queue slots, this should be left alone, and increased in multiples (e.g. try 200 next, then 400 if that's not enough, etc.).

### resizeinterval: [interval]

**Default: 30s**
This sets the interval at which the worker pool adjuster runs and possibly resizes the pool. Unless you're hitting memory issues, it's generally not advisable in making this too small, or you may suffer some inefficiencies from workers needing to fire up when Work arrives. Increasing the interval can be useful to curb yo-yoing due to infrequent, bursty loads. *0* disables automatic resizing.

## Handlers

Handlers form an ordered chain that inspects/manipulates requests. The term "middleware" is used interchangeably here. Handlers here are listed by their *Function name*, or *Type.Function name*. Parentheticals are used to denote when the configuration name deviates.

### AccessHandler

AccessHandler is a middleware that is automatically inserted towards the beginning of the chain, before the formal global handlers, if the **Path** has **Allow** or **Deny** set.
You cannot explicitly set or place this handler.

### AccessLogHandler

AccessLogHandler is a middleware that is automatically inserted towards the beginning of the chain, before the formal global handlers, that provides collation of request and subsequent response information, suitable for "access log" -style logging.
You cannot explicitly set or place this handler.

### BasicAuth.Handler

BasicAuth is a middleware that is automatically added to a chain (towards the beginning) if the **BasicAuthSource** Path property is used. The handler enforces HTTP Basic Authentication and possibly rate-limiting bad authentication requests.
You cannot explicitly set or place this handler.
**NOTE**: BasicAuth no longer supports DES-hashed passwords (classic UNIX crypt()) as of v3.0.0

### BodyByteLimit.Handler

BodyByteLimit is a middleware that is automatically added to a chain (towards the beginning) if the **BodyByteLimit** Path property is set (> 0). The handler enforces that the body is at most **BodyByteLimit** bytes in size.
You cannot explicitly set or place this handler.

### Compression.Handler

Compression.Handler is a middleware that is automatically inserted towards the beginning of the chain, before the formal global handlers,  if the **compression** configuration list has entries.
You cannot explicitly set or place this handler.

### CORS

CORS is a middleware that validates the provided *Origin* header against a configured whitelist (see **CORS.orgins**, below) and will reflect CORS data if there's a match.
It is automatically added after the global handlers, if **CORS** is defined in the config.
You cannot explicitly set or place this handler.

```yaml
CORS:
  origins:
    - "^https://.*.mydomain.(com|net)$"
    - "https://www.google.com"
  allowheaders: "origin,x-requested-with,content-type,accept,x-token,x-jarcluster"
  allowmethods: "GET, POST, PUT, DELETE, PATCH"
  allowcredentials: "false"
  maxage: "86400"
  privatenetwork: "false"
```

#### CORS.origins: [list]

This is a list of strings to test against. Regular expressions are fully supported.

#### CORS.allowheaders: [string]

A verbatim string put to in `Access-Control-Allow-Headers` and `Access-Control-Expose-Headers`

#### CORS.allowmethods: [string]

A verbatim string to put in `Access-Control-Allow-Methods`

#### CORS.allowcredentials: [string]

A verbatim string to put in `Access-Control-Allow-Credentials`

#### CORS.maxage: [string]

A verbatim string to put in `Access-Control-Max-Age`

#### CORS.privatenetwork: [string]

A verbative string to put in `Access-Control-Request-Private-Network`

### ErrorWrapper.Handler (ErrorHandler)

ErrorHandler is a middleware that should be at or near the beginning of a handler chain. Responses with *Request.StatusCode* >= 400 will be wrapped in the defined template.
**NOTE**: Currently, this handler will not wrap errors from any handler before it, including most of the access and authorization handlers. This is trivial to fix (hint: monkey ordering in ``BuildPaths()``) if it becomes an issue.

#### errorhandler.template: [filepath]

The path to an *html/template* template to render against. A **TemplateError** struct will be passed in:

```go
// TemplateError is a static structure to pass into error-wrapping templates
type TemplateError struct {
  // ErrorCode is the string value of the error
  ErrorCode string
  // ErrorMessage is an optional message the template may optionally render
  ErrorMessage string
  // RedirectURL is a URL the template is advised to redirect to
  RedirectURL string
  // RedirectSeconds is the number of seconds the template is advised to wait
  // before executing the RedirectURL
  RedirectSeconds int
}
```

### ForbiddenPaths.Handler

ForbiddenPaths.Handler is a middleware that is automatically added to a chain if the **forbiddenpaths** and/or **Path.forbiddenpaths** config directive/s is/are used. It ranges over pre-compiled PCREs to check the URI path and returns *403 Forbidden* if matched.
You cannot explicitly set or place this handler.

### OauthHandler (OAuth)

TODOC

### SwitchHandler

SwitchHandler is a middleware, best placed first or second in a chain, that adds mass-virtual-hosting information to the request Context, that subsequent handlers can use.

- Using **mapfiles**, discerns the "name", "id", and "endpoint" of the requested "name".

#### SwitchHandler.enforceorg: [true|false]

**Default: false**
If set, will stop requests for invalid "names" with a *400 Bad Request* response. If SwitchHandler is a global Handler, this is probably not a great idea, unless you only ever plan on handling org-specific requests (hint: you're not).

### PathReplacer.Handler

PathReplacer.Handler is a middleware that is automatically added to a chain, near the end, if the **ReplacePath** Path property is used. It does a simple string replacement to change the path of the request.
You cannot explicitly set or place this handler.

### RateLimiter.Handler

RateLimiter is a middleware that is automatically added to a chain (towards the beginning) if the **RateLimit** Path property is used.
You cannot explicitly set or place this handler.

### RealAddr

RealAddr is an always-embedded middleware that is *almost* always first. It checks the X-Forwarded-For header and replaces the request's *RemoteAddr* field with the believed client address. This is important for logging, and for access control.
You cannot explicitly set or place this handler.

#### disablerealaddr: [true|false]

**Default: false**
If set, the RealAddr middleware will not be embedded.
**NOTE**: This is not a supported or well-tested config, and the results of otherwise-working parts of the system may be unexpected.

### Recoverer

Recoverer is a middleware that protects against panics further down the chain, and replaces the unprofessionally defensive "server closed" with proper logging and passing a *500 Internal Server Error* error back to the caller.
**NOTE**: This does not change the stability of the running JAR instance, merely keeps an otherwise-crashed connection alive a bit longer to tell them there was a problem.

#### recoverer.logstacktraces: [true|false]

**Default: false**
If set, and a panic recovery is triggered, a stacktrace will be logged to the **errorlog**. Additionally, if set, all bootstrap panics will dump stack as well.

### ResponseHeaders

ResponseHeaders is an always-embedded middleware if the **headers** config is populated. Headers listed there are added to the response.
You cannot explicitly set or place this handler.

### SetupHandler

SetupHandler is an always-embedded middleware, that adds important information to the request Context, that subsequent handlers can use.

- Increments the  request counter
- Generates and sets the unique requestID
- Sets the context "ts" (timestamp) the request was seen

## Finishers

A single Finisher carries out the end request in lieu of a Pool

### EndpointDecider (urlswitch)

URLswitch checks the *endpoint* context setting, and if a **Pool** is defined with the same name, hands the request off to it, otherwise returns *400 Bad Request*, which is arguably the wrong code, but I'm arrogant enough to believe I have properly defined a pool for every cluster, and this won't happen unless the request is indeed "bad".

### Forbidden

Forbidden simply returns *403 Forbidden* when hit

### HealthCheck

HealthCheck collects and reports health and metrics information in the HealthCheck schema

### HTTPStatus*nnn*

Reflects the presented HTTP status code back to the caller, with the defined ``http.StatusText()`` for *nnn*. e.g. ``Finisher: HTTPStatus200`` is the same as ``Finisher: Ok`` below.

### Ok

Ok just returns *200 Ok* and "Ok".

### Restart

Restart causes a "USR2" signal to be sent to the process, gracefully restarting it.

### S3StreamProxy (s3proxy)

Handles a **multipart** form POST that has a file upload box, streaming the file to an S3 bucket.
Options are handled path-local, so different Paths may use different parameters.

```yaml
  -
    Path: /upload
    Finisher: s3proxy
    RateLimit: 10
    Options:
      s3proxy.name: upload
      s3proxy.bucket: somebucket
      s3proxy.prefix: someprefix/
      s3proxy.namefield: name
      s3proxy.emailfield: emailaddress
      s3proxy.tofield: sendto
      s3proxy.filefield: fileupload
      s3proxy.badfileexts: .scr .exe .bat .msi .dll
      s3proxy.wrapsuccess: true
```

TODO: Document better

### Stack

Stack returns a stackdump. Best not to let this out, umkay?

### TestingFinisher (Test)

TestingFinisher reflects request headers, cookie information, etc for debugging.

### TUS

The TUS finisher supports the [TUS](https://tus.io/) resumable upload protocol. Each **Path** needs a **tus.targeturi** set to a `file://` for local folder spooling or `s3://` for S3 spooling.

If you do "parallel" uploads > 1 on the client, there will be multiple "part files" left behind, in addition to the final file. It is recommended that your upload area be cleaned periodically of files old files. Yes, we could keep track of those parts, and after the final file is finished, delete the "part files" for you. We aren't.

```yaml
-
    Path: /tus/
    Options:
      tus.targeturi: file:///tmp/tus/
      tus.appendfilename: true
    Finisher: tus
-
    Path: /tus2/
    Options:
      tus.targeturi: s3://my-s3-bucket
      tus.appendfilename: true
    Finisher: tus
```

#### tus.targeturi: [file:// or s3:// URI for target]

Please note the `file://` URIs need an extra `/` for fully-qualified paths.

Please note that only root-level S3 buckets are supported at this time (no "folders").

#### tus.appendfilename: [true/false]

If `true` will append the orginal filename to the target filename, e.g. `hash-filename.ext`. **NOTE:** for S3 this is a COPY, DELETE and will incur additional charges. Also, for S3 this operation is **limited to files < 5GB in size**, as there is extra work required to "copy" files larger than 5GB within S3, and we're not doing that right now.

### Update

Update will download a - we hope - update .zip from the configured **updatepath**, and replace our executable with its contents. If **hotupdate** is set, a graceful restart will be requested to handle subsequent requests with the new code.

## Other

### keys: [key/value pairs]

#### keys.aws.region: [AWS Region]

#### keys.aws.access: [AWS AccessKeyID]

#### keys.aws.secret: [AWS Secret Key]

#### stickycookie: [base64-encoded key]

If present, sticky-session cookie values will be AES-GCM encoded using this key (Base64, standard-encoding assumed). Key size *must* be exactly one of 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.

Additionally used as-is as the salt for the *hash* sticky type. The previous rules do not apply.

### mapfiles: [key/value pairs]

### striprequestheaders: [list]

### zulip: [key/value pairs]

Enables messaging (and possibly logging) to a Zulip server.

#### url: [Base URL for Zulip Server]

#### username: [username to send as]

#### token: [user's token to authenticate with]

#### retrycount: [integer]

Sets the number of retries on failed send. Set to ``0`` to disable.

#### retryinterval: [duration]

Sets the duration to wait between retries.
