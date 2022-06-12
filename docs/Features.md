# Features

## Working / Advanced WIP

### TLS

TLS listening with configurable options is fully supported. See [Configuration Docs](Configuration.md) for supported configuration options. Additionally, HTTP-to-HTTPS redirects can optionally be handled automatically.

JAR TLS configurations have been tested by SSLlabs and meet or exceed current ratings in all criteria against current TLS configurations with HAproxy.

### Graceful Restarts

Restarts are supersafe. A new instance is started up, and if it fails (e.g. chokes on a config problem) the old instance never relinquishes the listener and keeps running. If the new instance starts successfully, the handover of the listener socket is coordinated, the old instance continues to run until all existing requests are finished, but all new requests run through the new instance.

Below is an example of a graceful restart that had a fairly long-running request on the "old" instance. The emitters from both "old" (request count 65832) and new (4 and climbing) are outputting every 5 seconds, with the "old" request count fixed, and the "new" request count climbing until the last "old" request finishes, allowing it to close down.

```bash
[DEBUG] 2018/05/31 00:26:29 health.go:47: Rate: 2.8804/second Goros: 25 (2 / 0.00) Requests: 65832
[DEBUG] 2018/05/31 00:26:32 health.go:47: Rate: 0.8000/second Goros: 31 (10 / 0.00) Requests: 4
[DEBUG] 2018/05/31 00:26:34 health.go:47: Rate: 2.6501/second Goros: 21 (2 / 0.00) Requests: 65832
[DEBUG] 2018/05/31 00:26:37 health.go:47: Rate: 1.1998/second Goros: 31 (10 / 0.00) Requests: 33
[DEBUG] 2018/05/31 00:26:39 health.go:47: Rate: 2.4382/second Goros: 21 (2 / 0.00) Requests: 65832
[DEBUG] 2018/05/31 00:26:42 health.go:47: Rate: 1.2638/second Goros: 31 (10 / 0.00) Requests: 43
[DEBUG] 2018/05/31 00:26:44 health.go:47: Rate: 2.2433/second Goros: 21 (2 / 0.00) Requests: 65832
[DEBUG] 2018/05/31 00:26:47 health.go:47: Rate: 1.5465/second Goros: 31 (10 / 0.00) Requests: 67
[DEBUG] 2018/05/31 00:26:49 health.go:47: Rate: 2.0639/second Goros: 21 (2 / 0.00) Requests: 65832
[DEBUG] 2018/05/31 00:26:52 health.go:47: Rate: 1.4708/second Goros: 31 (10 / 0.00) Requests: 70
[DEBUG] 2018/05/31 00:26:54 health.go:47: Rate: 1.8989/second Goros: 21 (2 / 0.00) Requests: 65832
[DEBUG] 2018/05/31 00:26:57 health.go:47: Rate: 1.4012/second Goros: 31 (10 / 0.00) Requests: 73
[DEBUG] 2018/05/31 00:27:02 health.go:47: Rate: 1.4651/second Goros: 27 (2 / 0.00) Requests: 84
[DEBUG] 2018/05/31 00:27:07 health.go:47: Rate: 1.7477/second Goros: 27 (2 / 0.00) Requests: 109
[DEBUG] 2018/05/31 00:27:12 health.go:47: Rate: 1.8478/second Goros: 27 (2 / 0.00) Requests: 124
```

### Configuration

See [Configuration.md](Configuration.md) for supported configuration options.

* Reading configuration files in  JSON, TOML, YAML, HCL, or Java properties -formats
* Reading configuration from environment variables
* Reading configuration from network services such as etcd or Consul
* Front-loaded validation of *most* options to aid **Graceful Restarts**
* Intelligent (and safe) defaults

#### Configuration Monitoring

In concert with **Graceful Restarts**, the configuration file can be watched for changes, triggering a graceful restart to absorb the changes

### Workers

Elastic pool of Workers hanging around, easily accessible (*AddWork(&WorkType{})*), waiting to do Work (and interface type) on behalf of Handlers/Finishers.

* Healthchecking Pool members (Per-Pool Lifeguard)
* API requests
  
See the [Writing Handlers and Finishers docs](WritingHandlersAndFinishers.md) and/or the [worker docs](../workers/Readme.md) for more info.

### Paths and Pools

The heart of JAR are Paths- filters and options that describe containers for traffic, and Pools- options that describe containers for destinations.

#### Path

* Prefix or absolute URI
* Host-based filter
* HTTP method filter
* HTTP request header filter
* Browser/version filter
* Per-source rate-limiting
* Allow/Deny by IP address/range
* HTTP Basic Auth (vs file)
* Request handling and completion options
  * Handlers (middleware)
  * Finisher
  * Redirect
  * Pool
* Pre-proxy URI prefix stripping
* Pre-proxy URI replacement
* Timeout

#### Pool

* Dynamic membership
* Request buffering/retrying
* Sticky sessions (cookie-based)
  * Plain, hex-encoded, or AES-encrypted values
* Response-header stripping
* Healthchecks and membership management
* EC2 awareness/affinity

### AWS awareness (Ongoing)

See the [AWS docs](Aws.md) for more information.

### Self-updating

Can be told, via an API endpoint, to download and replace the existing binary, and optionally trigger a **Graceful Restart** to start a new instance running on it.

Unscheduled but achievable future work includes binary patching, and code-signing/verification.

### Injectable Handlers and Finishers

See the [Writing Handlers and Finishers docs](WritingHandlersAndFinishers.md) for more information and examples. Ordered lists of middleware, called Handlers, or request completers, called Finishers, are trivial to author and inject into JAR, exposing their existence and configuration to the config system.
Highlights include:

* Request/Response logging (including "access log")
* Request header mangling
* CORS
* Browser/Device Detection
* Configurable response headers
* Map tile token verification
* JWT verification
* Rate limiting
* Content compression for configured MIME-types
* Flexible (and pluggable) error wrapping
* "Admin" Finishers
  * Healthchecks
  * Triggered self-updates
  * Triggered restarts

### First-Class Metrics

Any executable code can hook into the Metrics registry (see *WritingHandlersAndFinishers.md* for more information) to create:

* Counters
* Gauges
* Histograms
* Meters

These are automatically integrated into the *metrics* section of the **healthcheck** system.

```json
{
   "overallStatus":"OK",
   "metrics":[
     {
       "name":"Requests_count",
       "value":3528651
     },
     {
       "name":"Requests_1m.rate",
       "value":2.026142615386795
     },
     {
       "name":"Requests_5m.rate",
       "value":2.1115041158513814
     },
     {
       "name":"Requests_15m.rate",
       "value":2.2938515348864628
     },
     {
       "name":"Requests_mean.rate",
       "value":2.304830642151365
     }]
 }
```

### First-Class Healthchecks

Keeping track of how things are operating is critical, as is reporting those states.

### Custom error pages

Local template (precompiled at runtime for speed) ~~or subrequested out to a service (deprecated)~~.

### Timeouts and tunables

Global **timeout** (overridable per-Path) to kill long-running requests/responses.

### Low resource usage

Using metrics and benchmarks, care is taken to ensure heap allocations and time spent are kept as low as possible.

Minimizing "allocs" keeps the garbage collector more idle, allowing for many thousands of connections per second to be created and destroyed, with minimal waste that needs to be cleaned up after the fact.
