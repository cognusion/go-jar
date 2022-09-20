

# jar
`import "github.com/cognusion/go-jar"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Subdirectories](#pkg-subdirectories)

## <a name="pkg-overview">Overview</a>
Package jar is a readily-embeddable feature-rich proxy-focused AWS-aware
distributed-oriented resiliency-enabling URL-driven superlative-laced
***elastic application link***. At its core, JAR is "just" a load-balancing
proxy taking cues from HAProxy (resiliency, zero-drop restarts, performance)
and Apache HTTPD (virtualize everything) while leveraging over 20 years
of systems engineering experience to provide robust features with exceptional
stability.

JAR has been in production use since 2018 and handles millions of connections a day
across heterogeneous application stacks.

Consumers will want to 'cd cmd/jard; go build; #enjoy'




## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func AccessLogHandler(next http.Handler) http.Handler](#AccessLogHandler)
* [func AddMetrics(m map[string]map[string]interface{}, hc *health.Check) *health.Check](#AddMetrics)
* [func AddStatuses(s *health.StatusRegistry, hc *health.Check) *health.Check](#AddStatuses)
* [func AuthoritativeDomainsHandler(next http.Handler) http.Handler](#AuthoritativeDomainsHandler)
* [func Bootstrap()](#Bootstrap)
* [func BootstrapChan(closer chan struct{})](#BootstrapChan)
* [func BuildPath(path Path, index int, router *mux.Router) (int, error)](#BuildPath)
* [func BuildPaths(router *mux.Router) error](#BuildPaths)
* [func ChanBootstrap() chan error](#ChanBootstrap)
* [func ConnectionCounterAdd()](#ConnectionCounterAdd)
* [func ConnectionCounterGet() int64](#ConnectionCounterGet)
* [func ConnectionCounterRemove()](#ConnectionCounterRemove)
* [func CopyHeaders(dst http.Header, src http.Header)](#CopyHeaders)
* [func CopyRequest(req *http.Request) *http.Request](#CopyRequest)
* [func CopyURL(i *url.URL) *url.URL](#CopyURL)
* [func DumpFinisher(w http.ResponseWriter, r *http.Request)](#DumpFinisher)
* [func DumpHandler(h http.Handler) http.Handler](#DumpHandler)
* [func ECBDecrypt(b64key string, eb64ciphertext string) (plaintext []byte, err error)](#ECBDecrypt)
* [func ECBEncrypt(b64key string, plaintext []byte) (b64ciphertext string, err error)](#ECBEncrypt)
* [func EndpointDecider(w http.ResponseWriter, r *http.Request)](#EndpointDecider)
* [func FileExists(filePath string) bool](#FileExists)
* [func FlashEncoding(src string) string](#FlashEncoding)
* [func FolderExists(filePath string) bool](#FolderExists)
* [func Forbidden(w http.ResponseWriter, r *http.Request)](#Forbidden)
* [func GetErrorLog(filename, prefix string, format, size, backups, age int) *log.Logger](#GetErrorLog)
* [func GetLog(filename, prefix string, format, size, backups, age int) *log.Logger](#GetLog)
* [func GetLogOrDiscard(filename, prefix string, format, size, backups, age int) *log.Logger](#GetLogOrDiscard)
* [func GetRequestID(ctx context.Context) string](#GetRequestID)
* [func GetSwitchName(request *http.Request) string](#GetSwitchName)
* [func HandleFinisher(handler string) (http.HandlerFunc, error)](#HandleFinisher)
* [func HandleGenericWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool](#HandleGenericWrapper)
* [func HandleHandler(handler string, hchain alice.Chain) (alice.Chain, error)](#HandleHandler)
* [func HandleReload(name string, mfiles map[string]string)](#HandleReload)
* [func HandleRemoteWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool](#HandleRemoteWrapper)
* [func HandleTemplateWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool](#HandleTemplateWrapper)
* [func InitConfig() *viper.Viper](#InitConfig)
* [func LoadConfig(configFilename string, v *viper.Viper) error](#LoadConfig)
* [func LogInit() error](#LogInit)
* [func MinuteDelayer(w http.ResponseWriter, r *http.Request)](#MinuteDelayer)
* [func MinuteStreamer(w http.ResponseWriter, r *http.Request)](#MinuteStreamer)
* [func NewECBDecrypter(b cipher.Block) cipher.BlockMode](#NewECBDecrypter)
* [func NewECBEncrypter(b cipher.Block) cipher.BlockMode](#NewECBEncrypter)
* [func OkFinisher(w http.ResponseWriter, r *http.Request)](#OkFinisher)
* [func PoolLister(w http.ResponseWriter, r *http.Request)](#PoolLister)
* [func PoolMemberAdder(w http.ResponseWriter, r *http.Request)](#PoolMemberAdder)
* [func PoolMemberLister(w http.ResponseWriter, r *http.Request)](#PoolMemberLister)
* [func PoolMemberLoser(w http.ResponseWriter, r *http.Request)](#PoolMemberLoser)
* [func PrettyPrint(v interface{}) string](#PrettyPrint)
* [func ReaderToString(r io.Reader) string](#ReaderToString)
* [func RealAddr(h http.Handler) http.Handler](#RealAddr)
* [func Recoverer(next http.Handler) http.Handler](#Recoverer)
* [func ReplaceURI(r *http.Request, urlPath, requestURI string)](#ReplaceURI)
* [func RequestErrorResponse(r *http.Request, w http.ResponseWriter, Message string, code int)](#RequestErrorResponse)
* [func RequestErrorString(Request *http.Request, Message string) string](#RequestErrorString)
* [func ResponseHeaders(next http.Handler) http.Handler](#ResponseHeaders)
* [func Restart(w http.ResponseWriter, r *http.Request)](#Restart)
* [func RouteIDInspectionHandler(next http.Handler) http.Handler](#RouteIDInspectionHandler)
* [func S3StreamProxyFinisher(w http.ResponseWriter, r *http.Request)](#S3StreamProxyFinisher)
* [func SetupHandler(next http.Handler) http.Handler](#SetupHandler)
* [func Stack(w http.ResponseWriter, r *http.Request)](#Stack)
* [func StringIfCtx(r *http.Request, name interface{}) string](#StringIfCtx)
* [func SwitchHandler(next http.Handler) http.Handler](#SwitchHandler)
* [func TestFinisher(w http.ResponseWriter, r *http.Request)](#TestFinisher)
* [func TrimPrefixURI(r *http.Request, prefix string)](#TrimPrefixURI)
* [func URLCaptureHandler(next http.Handler) http.Handler](#URLCaptureHandler)
* [func Unzip(src, dest string) error](#Unzip)
* [func Update(w http.ResponseWriter, r *http.Request)](#Update)
* [func ValidateExtras() []error](#ValidateExtras)
* [func WithRqID(ctx context.Context, requestID string) context.Context](#WithRqID)
* [func WithSessionID(ctx context.Context, sessionID string) context.Context](#WithSessionID)
* [type Access](#Access)
  * [func NewAccessFromStrings(allow, deny string) (*Access, error)](#NewAccessFromStrings)
  * [func (a *Access) AccessHandler(next http.Handler) http.Handler](#Access.AccessHandler)
  * [func (a *Access) AddAddress(address string, allow bool) error](#Access.AddAddress)
  * [func (a *Access) Validate(address string) bool](#Access.Validate)
* [type AccessLog](#AccessLog)
* [type BasicAuth](#BasicAuth)
  * [func NewBasicAuth(source, realm string, users []string) *BasicAuth](#NewBasicAuth)
  * [func NewVerifiedBasicAuth(source, realm string, users []string) (*BasicAuth, error)](#NewVerifiedBasicAuth)
  * [func (b *BasicAuth) Authenticate(username, password, realm string) bool](#BasicAuth.Authenticate)
  * [func (b *BasicAuth) Load() error](#BasicAuth.Load)
  * [func (b *BasicAuth) VerifySource() error](#BasicAuth.VerifySource)
* [type BodyByteLimit](#BodyByteLimit)
  * [func NewBodyByteLimit(limit int64) BodyByteLimit](#NewBodyByteLimit)
  * [func (b *BodyByteLimit) Handler(next http.Handler) http.Handler](#BodyByteLimit.Handler)
* [type CORS](#CORS)
  * [func NewCORS() *CORS](#NewCORS)
  * [func NewCORSFromConfig(origins []string, conf map[string]string) (*CORS, error)](#NewCORSFromConfig)
  * [func (c *CORS) AddOrigin(origins []string) error](#CORS.AddOrigin)
  * [func (c *CORS) Handler(next http.Handler) http.Handler](#CORS.Handler)
  * [func (c *CORS) ResponseModifier(resp *http.Response) error](#CORS.ResponseModifier)
  * [func (c *CORS) String() string](#CORS.String)
* [type Cert](#Cert)
* [type Compression](#Compression)
  * [func NewCompression(contentTypes []string) *Compression](#NewCompression)
  * [func (c *Compression) Handler(next http.Handler) http.Handler](#Compression.Handler)
* [type ConfigKey](#ConfigKey)
* [type ConsistentHashPool](#ConsistentHashPool)
  * [func NewConsistentHashPool(source, key string, pool *Pool, next http.Handler) (*ConsistentHashPool, error)](#NewConsistentHashPool)
  * [func NewConsistentHashPoolOpts(source, key string, partitionCount, replicationFactor int, load float64, pool *Pool, next http.Handler) (*ConsistentHashPool, error)](#NewConsistentHashPoolOpts)
  * [func (ch *ConsistentHashPool) Next() http.Handler](#ConsistentHashPool.Next)
  * [func (ch *ConsistentHashPool) NextServer() (*url.URL, error)](#ConsistentHashPool.NextServer)
  * [func (ch *ConsistentHashPool) RemoveServer(u *url.URL) error](#ConsistentHashPool.RemoveServer)
  * [func (ch *ConsistentHashPool) ServeHTTP(w http.ResponseWriter, req *http.Request)](#ConsistentHashPool.ServeHTTP)
  * [func (ch *ConsistentHashPool) ServerWeight(u *url.URL) (int, bool)](#ConsistentHashPool.ServerWeight)
  * [func (ch *ConsistentHashPool) Servers() []*url.URL](#ConsistentHashPool.Servers)
  * [func (ch *ConsistentHashPool) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error](#ConsistentHashPool.UpsertServer)
* [type CorsString](#CorsString)
* [type DebugTrip](#DebugTrip)
  * [func (d *DebugTrip) RoundTrip(r *http.Request) (*http.Response, error)](#DebugTrip.RoundTrip)
* [type ErrConfigurationError](#ErrConfigurationError)
  * [func (e ErrConfigurationError) Error() string](#ErrConfigurationError.Error)
* [type ErrRequestError](#ErrRequestError)
  * [func (e ErrRequestError) Bytes() []byte](#ErrRequestError.Bytes)
  * [func (e ErrRequestError) Error() string](#ErrRequestError.Error)
  * [func (e ErrRequestError) String() string](#ErrRequestError.String)
  * [func (e ErrRequestError) WrappedResponse(code int, w http.ResponseWriter)](#ErrRequestError.WrappedResponse)
* [type Error](#Error)
  * [func (e Error) Error() string](#Error.Error)
* [type ErrorWrapper](#ErrorWrapper)
  * [func (e *ErrorWrapper) Handler(next http.Handler) http.Handler](#ErrorWrapper.Handler)
* [type FinisherMap](#FinisherMap)
  * [func (h *FinisherMap) List() []string](#FinisherMap.List)
* [type ForbiddenPaths](#ForbiddenPaths)
  * [func NewForbiddenPaths(paths []string) (*ForbiddenPaths, error)](#NewForbiddenPaths)
  * [func (f *ForbiddenPaths) Handler(next http.Handler) http.Handler](#ForbiddenPaths.Handler)
* [type GenericResponse](#GenericResponse)
  * [func (gr *GenericResponse) Finisher(w http.ResponseWriter, r *http.Request)](#GenericResponse.Finisher)
* [type HTTPWork](#HTTPWork)
  * [func (h *HTTPWork) Return(rthing interface{})](#HTTPWork.Return)
  * [func (h *HTTPWork) Work() interface{}](#HTTPWork.Work)
* [type HandlerMap](#HandlerMap)
  * [func (h *HandlerMap) List() []string](#HandlerMap.List)
* [type HealthCheckError](#HealthCheckError)
  * [func (h *HealthCheckError) Error() string](#HealthCheckError.Error)
* [type HealthCheckResult](#HealthCheckResult)
* [type HealthCheckStatus](#HealthCheckStatus)
  * [func StringToHealthCheckStatus(hc string) (HealthCheckStatus, error)](#StringToHealthCheckStatus)
  * [func (i HealthCheckStatus) String() string](#HealthCheckStatus.String)
* [type HealthCheckWork](#HealthCheckWork)
  * [func (h *HealthCheckWork) Return(rthing interface{})](#HealthCheckWork.Return)
  * [func (h *HealthCheckWork) Work() interface{}](#HealthCheckWork.Work)
* [type JSONAccessLog](#JSONAccessLog)
  * [func (a *JSONAccessLog) CommonLogFormat(combined bool) string](#JSONAccessLog.CommonLogFormat)
  * [func (a *JSONAccessLog) RequestFiller(r *http.Request)](#JSONAccessLog.RequestFiller)
  * [func (a *JSONAccessLog) Reset()](#JSONAccessLog.Reset)
  * [func (a *JSONAccessLog) ResponseFiller(endtime time.Time, duration time.Duration, responseCode int, responseLength int)](#JSONAccessLog.ResponseFiller)
* [type Member](#Member)
  * [func (m *Member) String() string](#Member.String)
* [type NoopResponseWriter](#NoopResponseWriter)
  * [func NewNoopResponseWriter() NoopResponseWriter](#NewNoopResponseWriter)
  * [func (n *NoopResponseWriter) Header() http.Header](#NoopResponseWriter.Header)
  * [func (n *NoopResponseWriter) Write(bytes []byte) (int, error)](#NoopResponseWriter.Write)
  * [func (n *NoopResponseWriter) WriteHeader(statusCode int)](#NoopResponseWriter.WriteHeader)
* [type Path](#Path)
* [type PathHandler](#PathHandler)
  * [func (p *PathHandler) Handler(next http.Handler) http.Handler](#PathHandler.Handler)
* [type PathOptions](#PathOptions)
  * [func (p *PathOptions) Get(key string) interface{}](#PathOptions.Get)
  * [func (p *PathOptions) GetBool(key string) bool](#PathOptions.GetBool)
  * [func (p *PathOptions) GetString(key string) string](#PathOptions.GetString)
  * [func (p *PathOptions) GetStringSlice(key string) []string](#PathOptions.GetStringSlice)
* [type PathReplacer](#PathReplacer)
  * [func (p *PathReplacer) Handler(next http.Handler) http.Handler](#PathReplacer.Handler)
* [type PathStripper](#PathStripper)
  * [func (p *PathStripper) Handler(next http.Handler) http.Handler](#PathStripper.Handler)
* [type Pool](#Pool)
  * [func (p *Pool) GetMember(u *url.URL) *Member](#Pool.GetMember)
  * [func (p *Pool) GetPool() (http.Handler, error)](#Pool.GetPool)
  * [func (p *Pool) IsMaterialized() bool](#Pool.IsMaterialized)
  * [func (p *Pool) Materialize() (http.Handler, error)](#Pool.Materialize)
* [type PoolConfig](#PoolConfig)
* [type PoolID](#PoolID)
  * [func (p *PoolID) Handler(next http.Handler) http.Handler](#PoolID.Handler)
* [type PoolManager](#PoolManager)
* [type PoolOptions](#PoolOptions)
  * [func (p *PoolOptions) Get(key string) interface{}](#PoolOptions.Get)
  * [func (p *PoolOptions) GetBool(key string) bool](#PoolOptions.GetBool)
  * [func (p *PoolOptions) GetFloat64(key string) float64](#PoolOptions.GetFloat64)
  * [func (p *PoolOptions) GetInt(key string) int](#PoolOptions.GetInt)
  * [func (p *PoolOptions) GetString(key string) string](#PoolOptions.GetString)
  * [func (p *PoolOptions) GetStringSlice(key string) []string](#PoolOptions.GetStringSlice)
* [type Pools](#Pools)
  * [func BuildPools() (*Pools, bool)](#BuildPools)
  * [func NewPools(poolConfigs map[string]*PoolConfig, interval time.Duration) *Pools](#NewPools)
  * [func (p *Pools) Exists(name string) bool](#Pools.Exists)
  * [func (p *Pools) Get(name string) (*Pool, bool)](#Pools.Get)
  * [func (p *Pools) List() []string](#Pools.List)
  * [func (p *Pools) Merge(pools map[string]*Pool)](#Pools.Merge)
  * [func (p *Pools) Replace(pools map[string]*Pool)](#Pools.Replace)
  * [func (p *Pools) Set(name string, pool *Pool)](#Pools.Set)
* [type ProcessInfo](#ProcessInfo)
  * [func NewProcessInfo(pid int32) *ProcessInfo](#NewProcessInfo)
  * [func (p *ProcessInfo) CPU() float64](#ProcessInfo.CPU)
  * [func (p *ProcessInfo) Memory() float64](#ProcessInfo.Memory)
  * [func (p *ProcessInfo) SetInterval(i time.Duration)](#ProcessInfo.SetInterval)
  * [func (p *ProcessInfo) UpdateCPU()](#ProcessInfo.UpdateCPU)
* [type ProxyResponseModifier](#ProxyResponseModifier)
* [type ProxyResponseModifierChain](#ProxyResponseModifierChain)
  * [func (p *ProxyResponseModifierChain) Add(prm ProxyResponseModifier)](#ProxyResponseModifierChain.Add)
  * [func (p *ProxyResponseModifierChain) ToProxyResponseModifier() ProxyResponseModifier](#ProxyResponseModifierChain.ToProxyResponseModifier)
* [type PruneFunc](#PruneFunc)
* [type RateLimiter](#RateLimiter)
  * [func NewRateLimiter(max float64, purgeDuration time.Duration) RateLimiter](#NewRateLimiter)
  * [func NewRateLimiterCollector(max float64, purgeDuration time.Duration) RateLimiter](#NewRateLimiterCollector)
  * [func (rl *RateLimiter) Handler(next http.Handler) http.Handler](#RateLimiter.Handler)
* [type Redirect](#Redirect)
  * [func (rd *Redirect) Finisher(w http.ResponseWriter, r *http.Request)](#Redirect.Finisher)
* [type S3Pool](#S3Pool)
  * [func NewS3Pool(s3url string) (*S3Pool, error)](#NewS3Pool)
  * [func (s3p *S3Pool) ServeHTTP(w http.ResponseWriter, r *http.Request)](#S3Pool.ServeHTTP)
* [type StatusFinisher](#StatusFinisher)
  * [func (sf StatusFinisher) Finisher(w http.ResponseWriter, r *http.Request)](#StatusFinisher.Finisher)
* [type SuiteMap](#SuiteMap)
  * [func NewSuiteMapFromCipherSuites(cipherSuites []*tls.CipherSuite) SuiteMap](#NewSuiteMapFromCipherSuites)
  * [func (s *SuiteMap) AllSuites() []uint16](#SuiteMap.AllSuites)
  * [func (s *SuiteMap) CipherListToSuites(list []string) ([]uint16, error)](#SuiteMap.CipherListToSuites)
  * [func (s *SuiteMap) List() []string](#SuiteMap.List)
  * [func (s *SuiteMap) Suite(number uint16) string](#SuiteMap.Suite)
* [type TemplateError](#TemplateError)
* [type Timeout](#Timeout)
  * [func (t *Timeout) Handler(next http.Handler) http.Handler](#Timeout.Handler)
* [type ZulipWork](#ZulipWork)
  * [func (z *ZulipWork) Return(rthing interface{})](#ZulipWork.Return)
  * [func (z *ZulipWork) Work() interface{}](#ZulipWork.Work)


#### <a name="pkg-files">Package files</a>
[a_common.go](https://github.com/cognusion/go-jar/tree/master/a_common.go) [access.go](https://github.com/cognusion/go-jar/tree/master/access.go) [basicauth.go](https://github.com/cognusion/go-jar/tree/master/basicauth.go) [compression.go](https://github.com/cognusion/go-jar/tree/master/compression.go) [config.go](https://github.com/cognusion/go-jar/tree/master/config.go) [cors.go](https://github.com/cognusion/go-jar/tree/master/cors.go) [crypto.go](https://github.com/cognusion/go-jar/tree/master/crypto.go) [debug.go](https://github.com/cognusion/go-jar/tree/master/debug.go) [errors.go](https://github.com/cognusion/go-jar/tree/master/errors.go) [finishers.go](https://github.com/cognusion/go-jar/tree/master/finishers.go) [handlers.go](https://github.com/cognusion/go-jar/tree/master/handlers.go) [health.go](https://github.com/cognusion/go-jar/tree/master/health.go) [healthprocess.go](https://github.com/cognusion/go-jar/tree/master/healthprocess.go) [helpers.go](https://github.com/cognusion/go-jar/tree/master/helpers.go) [log.go](https://github.com/cognusion/go-jar/tree/master/log.go) [macros.go](https://github.com/cognusion/go-jar/tree/master/macros.go) [paths.go](https://github.com/cognusion/go-jar/tree/master/paths.go) [pool.go](https://github.com/cognusion/go-jar/tree/master/pool.go) [poolch.go](https://github.com/cognusion/go-jar/tree/master/poolch.go) [poolconfig.go](https://github.com/cognusion/go-jar/tree/master/poolconfig.go) [pools.go](https://github.com/cognusion/go-jar/tree/master/pools.go) [proxyresponsemodifier.go](https://github.com/cognusion/go-jar/tree/master/proxyresponsemodifier.go) [s3pool.go](https://github.com/cognusion/go-jar/tree/master/s3pool.go) [s3proxy.go](https://github.com/cognusion/go-jar/tree/master/s3proxy.go) [taskscheduler.go](https://github.com/cognusion/go-jar/tree/master/taskscheduler.go) [update.go](https://github.com/cognusion/go-jar/tree/master/update.go) [urlswitch.go](https://github.com/cognusion/go-jar/tree/master/urlswitch.go) [version.go](https://github.com/cognusion/go-jar/tree/master/version.go) [worker-zulip.go](https://github.com/cognusion/go-jar/tree/master/worker-zulip.go) [workers.go](https://github.com/cognusion/go-jar/tree/master/workers.go) [z_zMustBeLast.go](https://github.com/cognusion/go-jar/tree/master/z_zMustBeLast.go)


## <a name="pkg-constants">Constants</a>
``` go
const (
    // ErrBootstrapDone should not be treated as a proper error, as it is returned if Bootstrap
    // is complete (e.g. checkconfig or doc output), and won't continue for non-error reasons
    ErrBootstrapDone = Error("Bootstrap() is done. This is not necessarily an error")

    // ErrPoolBuild is a panic in bootstrap() if BuildPools fails
    ErrPoolBuild = Error("error building pools")

    // ErrValidateExtras is a panic in bootstrap() if there are any errors in ValidateExtras.
    // Preceding error output may provide more specific information.
    ErrValidateExtras = Error("error validating extras")

    // ErrVersion is a panic in bootstrap() if versionrequired is set in the config, but is less
    // than the VERSION constant of the compiled binary reading the config.
    ErrVersion = Error("version requirement not met")
)
```
``` go
const (
    // ErrSourceVerificationFailed is an error returned when an authentication source cannot be verified
    ErrSourceVerificationFailed = Error("cannot verify the provided source")

    // ErrSourceNotSupported is an error returned when the authentication source scheme is not supported [yet]
    ErrSourceNotSupported = Error("specified source scheme is not supported")
)
```
``` go
const (
    ConfigAccessLog            = ConfigKey("accesslog")
    ConfigAuthPool             = ConfigKey("authpool")
    ConfigCheckConfig          = ConfigKey("checkconfig")
    ConfigCommonLog            = ConfigKey("commonlog")
    ConfigDebug                = ConfigKey("debug")
    ConfigDebugLog             = ConfigKey("debuglog")
    ConfigDebugRequests        = ConfigKey("debugrequests")
    ConfigDebugResponses       = ConfigKey("debugresponses")
    ConfigDebugTimings         = ConfigKey("debugtimings")
    ConfigDocs                 = ConfigKey("docs")
    ConfigDumpConfig           = ConfigKey("dumpconfig")
    ConfigEC2                  = ConfigKey("ec2")
    ConfigErrorLog             = ConfigKey("errorlog")
    ConfigHandlers             = ConfigKey("handlers")
    ConfigHotConfig            = ConfigKey("hotconfig")
    ConfigHotUpdate            = ConfigKey("hotupdate")
    ConfigKeepaliveTimeout     = ConfigKey("keepalivetimeout")
    ConfigKeys                 = ConfigKey("keys")
    ConfigKeysAwsRegion        = ConfigKey("keys.aws.region")
    ConfigKeysAwsAccessKey     = ConfigKey("keys.aws.access")
    ConfigKeysAwsSecretKey     = ConfigKey("keys.aws.secret")
    ConfigListen               = ConfigKey("listen")
    ConfigLogAge               = ConfigKey("logage")
    ConfigLogBackups           = ConfigKey("logbackups")
    ConfigLogSize              = ConfigKey("logsize")
    ConfigMaxConnections       = ConfigKey("maxconnections")
    ConfigPaths                = ConfigKey("paths")
    ConfigPools                = ConfigKey("pools")
    ConfigRequestIDHeaderName  = ConfigKey("requestidheadername")
    ConfigSlowLog              = ConfigKey("slowlog")
    ConfigSlowRequestMax       = ConfigKey("slowrequestmax")
    ConfigStripRequestHeaders  = ConfigKey("striprequestheaders")
    ConfigTempFolder           = ConfigKey("tempfolder")
    ConfigTimeout              = ConfigKey("timeout")
    ConfigTrustRequestIDHeader = ConfigKey("trustrequestidheader")
    ConfigUpdatePath           = ConfigKey("updatepath")
    ConfigAuthoritativeDomains = ConfigKey("authoritativedomains")
    ConfigVersionRequired      = ConfigKey("versionrequired")
    ConfigLogFakeXFF           = ConfigKey("fakexfflog")

    ConfigURLRouteHeaders            = ConfigKey("urlroute.enableheaders")
    ConfigURLRouteIDHeaderName       = ConfigKey("urlroute.idheadername")
    ConfigURLRouteEndpointHeaderName = ConfigKey("urlroute.endpointheadername")
    ConfigURLRouteNameHeaderName     = ConfigKey("urlroute.nameheadername")
    ConfigPoolHeaderName             = ConfigKey("urlroute.poolheadername")
    ConfigPoolMemberHeaderName       = ConfigKey("urlroute.poolmemberheadername")
)
```
Constants for configuration key strings

``` go
const (
    ConfigCORSAllowCredentials = ConfigKey("CORS.allowcredentials")
    ConfigCORSAllowHeaders     = ConfigKey("CORS.allowheaders")
    ConfigCORSAllowMethods     = ConfigKey("CORS.allowmethods")
    ConfigCORSOrigins          = ConfigKey("CORS.origins")
    ConfigCORSMaxAge           = ConfigKey("CORS.maxage")

    CORSAllowOrigin      = CorsString("Access-Control-Allow-Origin")
    CORSAllowCredentials = CorsString("Access-Control-Allow-Credentials")
    CORSExposeHeaders    = CorsString("Access-Control-Expose-Headers")
    CORSAllowMethods     = CorsString("Access-Control-Allow-Methods")
    CORSAllowHeaders     = CorsString("Access-Control-Allow-Headers")
    CORSMaxAge           = CorsString("Access-Control-Max-Age")
)
```
Constants for configuration key strings

``` go
const (
    // ErrCiphertextTooShort is returned when the ciphertext is too damn short
    ErrCiphertextTooShort = Error("ciphertext too short")

    // ErrCiphertextIrregular is returned when the ciphertext is not a multiple of the block size
    ErrCiphertextIrregular = Error("ciphertext is not a multiple of the block size")
)
```
``` go
const (
    ConfigTLSCerts             = ConfigKey("tls.certs")
    ConfigTLSCiphers           = ConfigKey("tls.ciphers")
    ConfigTLSEnabled           = ConfigKey("tls.enabled")
    ConfigTLSHTTPRedirects     = ConfigKey("tls.httpredirects")
    ConfigTLSKeepaliveDisabled = ConfigKey("tls.keepalivedisabled")
    ConfigTLSListen            = ConfigKey("tls.listen")
    ConfigTLSMaxVersion        = ConfigKey("tls.maxversion")
    ConfigTLSMinVersion        = ConfigKey("tls.minversion")
    ConfigTLSHTTP2             = ConfigKey("tls.http2")
)
```
Constants for configuration key strings

``` go
const (

    // ErrPreviousError is exclusive to the HandleSuddenEviction scheme,
    // and announces that original error should be the one returned to the caller
    ErrPreviousError = Error("previous error stands")

    // ErrAuthError is used to communicate authentication errors. More detail
    // will be in the error log, but let's not leak that, shall we?
    ErrAuthError = Error("an error occurred during AAA")

    // ErrForbiddenError is used to communicate resource access denial
    ErrForbiddenError = Error("you do not have access to this resource")

    // ErrNoSession is called when an AWS feature is called, but there is no initialized AWS session
    ErrNoSession = Error("there is no initialized AWS session")

    // ErrUnknownError is returned when an error occurs for undefined-yet-anticipated reasons
    ErrUnknownError = Error("unknown error")
)
```
``` go
const (
    ConfigErrorHandlerTemplate = ConfigKey("errorhandler.template")
    ConfigErrorHandlerURI      = ConfigKey("errorhandler.uri")
)
```
Constants for configuration key strings

``` go
const (

    // PathOptionsKey is a keyid for setting/getting PathOptions to/from a Context
    PathOptionsKey commonIDKey

    // ErrAborted is only used during panic recovery, if http.ErrAbortHandler was called
    ErrAborted = Error("client aborted connection, or connection closed")
)
```
``` go
const (
    ConfigHeaders                 = ConfigKey("headers")
    ConfigRecovererLogStackTraces = ConfigKey("Recoverer.logstacktraces")
)
```
Constants for configuration key strings

``` go
const (
    // ErrNoSuchEntryError is returned by the Status Registry when no status exists for the requested thing
    ErrNoSuchEntryError = Error("no such element exists")

    // ErrNoSuchHealthCheckStatus is returned when a string-based status has been used, but no corresponding HealthCheckStatus exists
    ErrNoSuchHealthCheckStatus = Error("no such HealthCheckStatus exists")
)
```
``` go
const (
    ConfigCompression     = ConfigKey("compression")
    ConfigDisableRealAddr = ConfigKey("disablerealaddr")
    ConfigForbiddenPaths  = ConfigKey("forbiddenpaths")
)
```
Constants for configuration key strings

``` go
const (
    // ErrPoolsConfigdefaultmembererrorstatusInvalid is returned when the pools.defaultmembererrorstatus is set improperly
    ErrPoolsConfigdefaultmembererrorstatusInvalid = Error("pools.defaultmembererrorstatus is set to an invalid HealthCheckStatus")

    // ErrPoolsConfigdefaultmembererrorstatusEmpty is returned when the pools.defaultmembererrorstatus is empty
    ErrPoolsConfigdefaultmembererrorstatusEmpty = Error("pools.defaultmembererrorstatus is empty")

    // ErrPoolStickyAESNoKey is returned when materializing a Pool with StickyCookieType set to 'aes' but 'keys.stickycookie' is not set
    ErrPoolStickyAESNoKey = Error("Pool.StickyCookieType set to 'aes' but no keys.stickycookie set")

    // ErrPoolAddMemberNotSupported is returned when Pool.AddMember is called on a Pool that doesn't support the operation
    ErrPoolAddMemberNotSupported = Error("this Pool does not support dynamic adding of members")

    // ErrPoolDeleteMemberNotSupported is returned when Pool.DeleteMember is called on a Pool that doesn't support the operation
    ErrPoolDeleteMemberNotSupported = Error("this Pool does not support dynamic deletion of members")

    // ErrPoolRemoveMemberNotSupported is returned when Pool.RemoveMember is called on a Pool that doesn't support the operation
    ErrPoolRemoveMemberNotSupported = Error("this Pool does not support dynamic removing of members")

    // ErrPoolNoMembersConfigured is returned when a non-dynamic Pool type (e.g. S3) has no members configured
    ErrPoolNoMembersConfigured = Error("no members configured for a non-dynamic Pool")

    // ErrPoolConfigConsistentAndSticky is returned when a Pool has both Sticky and ConsistentHashing set
    ErrPoolConfigConsistentAndSticky = Error("a Pool cannot have Sticky and ConsistentHashing set")
)
```
``` go
const (
    ConfigKeysStickyCookie           = ConfigKey("keys.stickycookie")
    ConfigStickyCookieAESTTL         = ConfigKey("stickycookie.aes.ttl")
    ConfigStickyCookieHTTPOnly       = ConfigKey("stickycookie.httponly")
    ConfigStickyCookieSecure         = ConfigKey("stickycookie.secure")
    ConfigConsistentHashPartitions   = ConfigKey("consistenthash.partitions")
    ConfigConsistentHashReplications = ConfigKey("consistenthash.replfactor")
    ConfigConsistentHashLoad         = ConfigKey("consistenthash.load")
)
```
Constants for configuration key strings

``` go
const (
    // ErrConsistentHashNextServerUnsupported is returned if NextServer is called
    ErrConsistentHashNextServerUnsupported = Error("Consistent Hash Pools don't support NextServer")

    // ErrConsistentHashInvalidSource is returned the source is not one of "request", "header", or "cookie"
    ErrConsistentHashInvalidSource = Error("the source provided is not valid")
)
```
``` go
const (
    ConfigPoolsDefaultMemberErrorStatus               = ConfigKey("pools.defaultmembererrorstatus")
    ConfigPoolsDefaultMemberWeight                    = ConfigKey("pools.defaultmemberweight")
    ConfigPoolsHealthcheckInterval                    = ConfigKey("pools.healthcheckinterval")
    ConfigPoolsLocalMemberWeight                      = ConfigKey("pools.localmemberweight")
    ConfigPoolsPreMaterialize                         = ConfigKey("pools.prematerialize")
    ConfigPoolsDefaultConsistentHashPartitions        = ConfigKey("pools.defaultconsistenthashpartitions")
    ConfigPoolsDefaultConsistentHashReplicationFactor = ConfigKey("pools.defaultconsistenthashreplicationfactor")
    ConfigPoolsDefaultConsistentHashLoad              = ConfigKey("pools.defaultconsistenthashload")
)
```
Constants for configuration key strings

``` go
const (
    ConfigS3StreamProxyName              = ConfigKey("s3proxy.name")
    ConfigS3StreamProxyBucket            = ConfigKey("s3proxy.bucket")
    ConfigS3StreamProxyPrefix            = ConfigKey("s3proxy.prefix")
    ConfigS3StreamProxyRedirectURL       = ConfigKey("s3proxy.redirecturl")
    ConfigS3StreamProxyFormNameField     = ConfigKey("s3proxy.namefield")
    ConfigS3StreamProxyFormEmailField    = ConfigKey("s3proxy.emailfield")
    ConfigS3StreamProxyFormToField       = ConfigKey("s3proxy.tofield")
    ConfigS3StreamProxyFormFileField     = ConfigKey("s3proxy.filefield")
    ConfigS3StreamProxyBadFileExtensions = ConfigKey("s3proxy.badfileexts")
    ConfigS3StreamProxyWrapSuccess       = ConfigKey("s3proxy.wrapsuccess")
    ConfigS3StreamProxyZulipStream       = ConfigKey("s3proxy.zulipstream")
    ConfigS3StreamProxyZulipTopic        = ConfigKey("s3proxy.zuliptopic")
)
```
Constants for configuration key strings

``` go
const (
    // ErrUpdateConfigS3NoEC2 is returned when the s3 updatepath is set, but ec2 is not.
    ErrUpdateConfigS3NoEC2 = Error("s3 updatepath set, but ec2 is false")

    // ErrUpdateConfigEmptyURL is returned when the updatepath is empty
    ErrUpdateConfigEmptyURL = Error("update url is empty, not updating")
)
```
``` go
const (
    ConfigMapFiles                 = ConfigKey("mapfiles")
    ConfigMapIDMap                 = ConfigKey("urlroute.idmap")
    ConfigMapEndpointMap           = ConfigKey("urlroute.endpointmap")
    ConfigSwitchHandlerEnforce     = ConfigKey("SwitchHandler.enforce")
    ConfigSwitchHandlerStripPrefix = ConfigKey("SwitchHandler.stripprefix") // e.g. xzy strips ^xyz.*-
)
```
Constants for configuration key strings

``` go
const (
    ConfigZulipBaseURL       = ConfigKey("zulip.url")
    ConfigZulipUsername      = ConfigKey("zulip.username")
    ConfigZulipToken         = ConfigKey("zulip.token")
    ConfigZulipRetryCount    = ConfigKey("zulip.retrycount")
    ConfigZulipRetryInterval = ConfigKey("zulip.retryinterval")
)
```
Constants for configuration key strings

``` go
const (
    ConfigWorkersInitialPoolSize = ConfigKey("workers.initialpoolsize")
    ConfigWorkersMaxPoolSize     = ConfigKey("workers.maxpoolsize")
    ConfigWorkersMinPoolSize     = ConfigKey("workers.minpoolsize")
    ConfigWorkersQueueSize       = ConfigKey("workers.queuesize")
    ConfigWorkersResizeInterval  = ConfigKey("workers.resizeinterval")
)
```
Constants for configuration key strings

``` go
const (
    ConfigMacros = ConfigKey("macros")
)
```
Constants for configuration key strings

``` go
const (
    // ErrFinisher404 returned by HandleFinisher if the requested finisher doesn't exist. Other errors should be treated as failures
    ErrFinisher404 = Error("requested finisher does not exist")
)
```
``` go
const (
    // ErrInvalidS3URL is returned when the relevant URL parts from the provided S3 URL cannot be derived
    ErrInvalidS3URL = Error("the S3 URL passed is invalid")
)
```
``` go
const (
    // ErrNoSuchMemberError is returned if the member doesn't exist or has been removed from a Pool
    ErrNoSuchMemberError = Error("member no longer exists in pool")
)
```
``` go
const (
    // ErrNoZulipClient is returned by a worker when there is no Zulip client defined
    ErrNoZulipClient = Error("no Zulip client defined")
)
```
``` go
const (
    // ErrS3ProxyConfigNoEC2 is returned when the s3proxy is used, but ec2 is not.
    ErrS3ProxyConfigNoEC2 = Error("s3proxy used, but ec2 is false")
)
```
``` go
const (
    // VERSION is the internal code revision number
    VERSION string = "1.1.0+git"
)
```

## <a name="pkg-variables">Variables</a>
``` go
var (
    // Conf is the config struct
    Conf *viper.Viper

    // StopFuncs is an aggregator for functions that needs to be called during graceful shutdowns.
    // Can only be called once!
    StopFuncs = funcregistry.NewFuncRegistry(true)

    // StrainFuncs is an aggregator for functions that can be called when JAR is under resource pressure.
    StrainFuncs = funcregistry.NewFuncRegistry(false)

    // InitFuncs are called in the early phases of Bootstrap()
    InitFuncs = funcregistry.NewFuncRegistry(true)

    // FileWatcher is an abstracted mechanism for calling WatchHandlerFuncs when a file is changed
    FileWatcher *watcher.Watcher

    // LoadBalancers are Pools
    LoadBalancers *Pools

    // Seq is a Sequence used for request ids
    Seq = sequence.New(1)

    // APPENDIX is an array of functions that have something to contribute to the docs appendix
    APPENDIX = make([]func(int), 0)

    // Ec2Session is an aws.Session for use in various places
    Ec2Session *aws.Session

    // Hostname is a local cache of os.Hostname
    Hostname string
)
```
``` go
var (
    // ConfigValidations is used to wire in func()error to be run, validating distributed configs
    ConfigValidations = make(map[string]func() error)
    // ConfigAdditions is used to wire in additional default configurations
    ConfigAdditions = make(configDefaultSetter)
)
```
``` go
var (
    // Metrics is a Registry for metrics, to be reported in the healthcheck
    Metrics = metrics.NewRegistry()
    // Status is a Registry for statuses, to be reported in the healthcheck
    Status = health.NewStatusRegistry()

    // NUMCPU is the number of CPUs at starttime
    NUMCPU = runtime.NumCPU()
    // GOVERSION is the version of Go
    GOVERSION = runtime.Version()

    // Counter is the clicker to a request counter.
    Counter func()

    // ThisProcess is updated information about this process
    ThisProcess *ProcessInfo

    // ConnectionCounter is used for tracking the current number of connections served
    ConnectionCounter int64

    // CurrentHealthCheck is a cache of the current state, refreshed periodically
    CurrentHealthCheck atomic.Value

    // HealthCheck is a Finisher that writes the healthcheck
    HealthCheck = healthCheckAsync

    // TerseHealthCheck is a Finisher that writes the terse healthcheck
    TerseHealthCheck = terseHealthCheckAsync
)
```
``` go
var (

    // CheckAuthoritative compares domain suffixes in the "authoritativedomains" against the requested URL.Hostname()
    // and returns true if it matches or if "authoritativedomains" is not used
    CheckAuthoritative = func(*http.Request) bool { return true }

    // RecyclableBufferPool is a sync.Pool of RecyclableBuffers that are safe to Get() and use (after a reset), and then
    // Close() them when you're done, to put them back in the Pool
    RecyclableBufferPool sync.Pool
)
```
``` go
var (
    // OutFormat is a log.Logger format used by default
    OutFormat = log.Ldate | log.Ltime | log.Lshortfile
    // DebugOut is a log.Logger for debug messages
    DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
    // TimingOut is a log.Logger for timing-related debug messages. DEPRECATED
    TimingOut = log.New(io.Discard, "[TIMING] ", 0)
    // ErrorOut is a log.Logger for error messages
    ErrorOut = log.New(os.Stderr, "", OutFormat)
    // AccessOut is a log.Logger for access logging. PLEASE DO NOT USE THIS DIRECTLY
    AccessOut = log.New(os.Stdout, "", 0)
    // CommonOut is a log.Logger for Apache "common log format" logging. PLEASE DO NOT USE THIS DIRECTLY
    CommonOut = log.New(io.Discard, "", 0)
    // DocsOut is a log.Logger for documentation output
    DocsOut = log.New(io.Discard, "", 0)
    // SlowOut is a log.Logger for slow request information
    SlowOut = log.New(io.Discard, "", 0)

    // RequestTimer is a function to allow Durations to be added to the Timer Metric
    RequestTimer func(time.Duration)

    // SlowRequests is the slow request log Duration
    SlowRequests time.Duration

    // LogPool is a Pool of AccessLogs
    LogPool sync.Pool
)
```
``` go
var (
    // DefaultTrip should be used instead of the http.DefaultTransport, for pools/etc.
    DefaultTrip http.RoundTripper

    // DefaultClient should be used instead of using http.DefaultClient, for pools/etc.
    DefaultClient *http.Client

    // DefaultMemberWeight is the weight added to each member by default
    DefaultMemberWeight int
    // LocalMemberWeight is the weight assigned to each member that is AZ-local
    LocalMemberWeight int

    // ResponseModiferChain is a ProxyResponseModifierChain to handle sequences of modifications
    // use ``ResponseModiferChain.Add()`` to add your own
    ResponseModiferChain ProxyResponseModifierChain
)
```
``` go
var (
    // RestartSelf is a niladic that will trigger a graceful restart of this process
    RestartSelf func()
    // IntSelf is a niladic that will trigger an interrupt of this process
    IntSelf func()
    // KillSelf is a niladic that will trigger a graceful shutdown of this process
    KillSelf func()
)
```
``` go
var (
    // Workers are a pool of workers
    Workers *workers.WorkerPool
    // AddWork queues up some work for workers
    AddWork func(workers.Work)
)
```
``` go
var (
    // CorsHandler is the global handler for CORS
    CorsHandler func(next http.Handler) http.Handler
)
```
``` go
var (
    // ErrorTemplate is an HTML template for returning errors
    ErrorTemplate *template.Template
)
```
``` go
var (
    // Finishers is a map of available HandlerFuncs
    Finishers = make(FinisherMap)
)
```
``` go
var (
    // Handlers is a map of available Handlers (middlewares)
    Handlers = make(HandlerMap)
)
```
``` go
var (
    // MacroDictionary is a Dictionary for doing mcro
    MacroDictionary dictionary.Resolver
)
```
``` go
var (
    // SwitchMaps are maps of URLs parts and their IDs and/or endpoints
    SwitchMaps = mapmap.NewMapMap()
)
```
``` go
var (
    // TaskRegistry is for wrangling scheduled tasks
    TaskRegistry cronzilla.Wrangler
)
```
``` go
var ZulipClient *zulip.Zulip
```
ZulipClient is a global Zulip client to use for messaging, or nil if not



## <a name="AccessLogHandler">func</a> [AccessLogHandler](https://github.com/cognusion/go-jar/tree/master/log.go?s=8817:8870#L283)
``` go
func AccessLogHandler(next http.Handler) http.Handler
```
AccessLogHandler is a middleware that times how long requests takes, assembled an AccessLog, and logs accordingly



## <a name="AddMetrics">func</a> [AddMetrics](https://github.com/cognusion/go-jar/tree/master/health.go?s=8124:8208#L320)
``` go
func AddMetrics(m map[string]map[string]interface{}, hc *health.Check) *health.Check
```
AddMetrics ranges over the supplied map, adding each as a Metric to the supplied Check



## <a name="AddStatuses">func</a> [AddStatuses](https://github.com/cognusion/go-jar/tree/master/health.go?s=7833:7907#L309)
``` go
func AddStatuses(s *health.StatusRegistry, hc *health.Check) *health.Check
```
AddStatuses ranges over the supplied StatusRegistry, adding each as a Service to the supplied Check



## <a name="AuthoritativeDomainsHandler">func</a> [AuthoritativeDomainsHandler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=2209:2273#L85)
``` go
func AuthoritativeDomainsHandler(next http.Handler) http.Handler
```
AuthoritativeDomainsHandler declines to handle requests that are not listed in "authoritativedomains" config



## <a name="Bootstrap">func</a> [Bootstrap](https://github.com/cognusion/go-jar/tree/master/a_common.go?s=6329:6345#L235)
``` go
func Bootstrap()
```
Bootstrap assumes that the Conf object is all set, for now at least,
builds the necessary subsystems and starts running.

Bootstrap doesn't return unless the server exits



## <a name="BootstrapChan">func</a> [BootstrapChan](https://github.com/cognusion/go-jar/tree/master/a_common.go?s=4336:4376#L142)
``` go
func BootstrapChan(closer chan struct{})
```
BootstrapChan assumes that the Conf object is all set, for now at least,
builds the necessary subsystems and starts running.

BootstrapChan doesn't return unless the server exits or the passed chan is closed



## <a name="BuildPath">func</a> [BuildPath](https://github.com/cognusion/go-jar/tree/master/paths.go?s=5990:6059#L180)
``` go
func BuildPath(path Path, index int, router *mux.Router) (int, error)
```
BuildPath does the heavy lifting to build a single path (which may result in multiple paths, but that's just bookkeeping)



## <a name="BuildPaths">func</a> [BuildPaths](https://github.com/cognusion/go-jar/tree/master/paths.go?s=5279:5320#L154)
``` go
func BuildPaths(router *mux.Router) error
```
BuildPaths unmarshalls the paths config, creates handler chains, and updates the mux



## <a name="ChanBootstrap">func</a> [ChanBootstrap](https://github.com/cognusion/go-jar/tree/master/a_common.go?s=5381:5412#L185)
``` go
func ChanBootstrap() chan error
```
ChanBootstrap assumes that the Conf object is all set, for now at least,
builds the necessary subsystems and starts running.

ChanBootstrap returns quickly, and should be assumed running unless an error
is received on the returned chan. ErrBootstrapDone should not be treated as a
proper error, as it is returned if Bootstrap is complete (e.g. checkconfig or doc output)



## <a name="ConnectionCounterAdd">func</a> [ConnectionCounterAdd](https://github.com/cognusion/go-jar/tree/master/health.go?s=5382:5409#L206)
``` go
func ConnectionCounterAdd()
```
ConnectionCounterAdd atomically adds 1 to the ConnectionCounter



## <a name="ConnectionCounterGet">func</a> [ConnectionCounterGet](https://github.com/cognusion/go-jar/tree/master/health.go?s=5689:5722#L216)
``` go
func ConnectionCounterGet() int64
```
ConnectionCounterGet atomically returns the current value of the ConnectionCounter



## <a name="ConnectionCounterRemove">func</a> [ConnectionCounterRemove](https://github.com/cognusion/go-jar/tree/master/health.go?s=5526:5556#L211)
``` go
func ConnectionCounterRemove()
```
ConnectionCounterRemove atomically adds -1 to the ConnectionCounter



## <a name="CopyHeaders">func</a> [CopyHeaders](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=3371:3421#L129)
``` go
func CopyHeaders(dst http.Header, src http.Header)
```
CopyHeaders copies http headers from source to destination, it
does not overide, but adds multiple headers



## <a name="CopyRequest">func</a> [CopyRequest](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=2858:2907#L108)
``` go
func CopyRequest(req *http.Request) *http.Request
```
CopyRequest provides a safe copy of a bodyless request into a new request



## <a name="CopyURL">func</a> [CopyURL](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=3139:3172#L118)
``` go
func CopyURL(i *url.URL) *url.URL
```
CopyURL provides update safe copy by avoiding shallow copying User field



## <a name="DumpFinisher">func</a> [DumpFinisher](https://github.com/cognusion/go-jar/tree/master/debug.go?s=3842:3899#L147)
``` go
func DumpFinisher(w http.ResponseWriter, r *http.Request)
```
DumpFinisher is a special finisher that reflects a ton of request output



## <a name="DumpHandler">func</a> [DumpHandler](https://github.com/cognusion/go-jar/tree/master/debug.go?s=3489:3534#L137)
``` go
func DumpHandler(h http.Handler) http.Handler
```
DumpHandler is a special handler that ships a ton of request output to DebugLog



## <a name="ECBDecrypt">func</a> [ECBDecrypt](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=5931:6014#L186)
``` go
func ECBDecrypt(b64key string, eb64ciphertext string) (plaintext []byte, err error)
```
ECBDecrypt takes a base64-encoded key and RawURLencoded-base64 ciphertext to decrypt, and returns the plaintext or an error.
PKCS5 padding is trimmed as needed



## <a name="ECBEncrypt">func</a> [ECBEncrypt](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=6701:6783#L216)
``` go
func ECBEncrypt(b64key string, plaintext []byte) (b64ciphertext string, err error)
```
ECBEncrypt takes a base64-encoded key and a []byte, and returns the base64-encdoded ciphertext or an error.
PKCS5 padding is added as needed



## <a name="EndpointDecider">func</a> [EndpointDecider](https://github.com/cognusion/go-jar/tree/master/urlswitch.go?s=4426:4486#L143)
``` go
func EndpointDecider(w http.ResponseWriter, r *http.Request)
```
EndpointDecider is a Finisher that inspects the ``switchEndpointKey`` context to determine which materialized
Pool should get the request.
Requests for clusters that are not materialized, or not having the ``clustername`` context value set
will result in unrecoverable errors



## <a name="FileExists">func</a> [FileExists](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=2232:2269#L84)
``` go
func FileExists(filePath string) bool
```
FileExists returns true if the provided path exists, and is not a directory



## <a name="FlashEncoding">func</a> [FlashEncoding](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=4764:4801#L165)
``` go
func FlashEncoding(src string) string
```
FlashEncoding returns a URL-encoded version of the provided string,
with "+" additionally converted to "%2B"



## <a name="FolderExists">func</a> [FolderExists](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=2439:2478#L92)
``` go
func FolderExists(filePath string) bool
```
FolderExists returns true if the provided path exists, and is a directory



## <a name="Forbidden">func</a> [Forbidden](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=2542:2596#L94)
``` go
func Forbidden(w http.ResponseWriter, r *http.Request)
```
Forbidden is a Finisher that returns 403 for the requested Path



## <a name="GetErrorLog">func</a> [GetErrorLog](https://github.com/cognusion/go-jar/tree/master/log.go?s=3584:3669#L110)
``` go
func GetErrorLog(filename, prefix string, format, size, backups, age int) *log.Logger
```
GetErrorLog gets an error-type log



## <a name="GetLog">func</a> [GetLog](https://github.com/cognusion/go-jar/tree/master/log.go?s=3150:3230#L98)
``` go
func GetLog(filename, prefix string, format, size, backups, age int) *log.Logger
```
GetLog gets a standard-type log



## <a name="GetLogOrDiscard">func</a> [GetLogOrDiscard](https://github.com/cognusion/go-jar/tree/master/log.go?s=3377:3466#L104)
``` go
func GetLogOrDiscard(filename, prefix string, format, size, backups, age int) *log.Logger
```
GetLogOrDiscard gets a standard-type log, or discards the output



## <a name="GetRequestID">func</a> [GetRequestID](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=2645:2690#L100)
``` go
func GetRequestID(ctx context.Context) string
```
GetRequestID is returns a requestID from a context, or the empty string



## <a name="GetSwitchName">func</a> [GetSwitchName](https://github.com/cognusion/go-jar/tree/master/urlswitch.go?s=6275:6323#L199)
``` go
func GetSwitchName(request *http.Request) string
```
GetSwitchName is a function to return the switch name in a request's context, if present



## <a name="HandleFinisher">func</a> [HandleFinisher](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=1127:1188#L48)
``` go
func HandleFinisher(handler string) (http.HandlerFunc, error)
```
HandleFinisher takes a Finisher HandlerFunc name, and returns the function for it and nil, or nil and and error



## <a name="HandleGenericWrapper">func</a> [HandleGenericWrapper](https://github.com/cognusion/go-jar/tree/master/errors.go?s=6778:6881#L217)
``` go
func HandleGenericWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool
```
HandleGenericWrapper is essentially a noop for when no tempate or remote errorhandler is defined



## <a name="HandleHandler">func</a> [HandleHandler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=1791:1866#L73)
``` go
func HandleHandler(handler string, hchain alice.Chain) (alice.Chain, error)
```
HandleHandler takes a handler name, and an existing chain, and returns a new chain or an error



## <a name="HandleReload">func</a> [HandleReload](https://github.com/cognusion/go-jar/tree/master/urlswitch.go?s=6514:6570#L207)
``` go
func HandleReload(name string, mfiles map[string]string)
```
HandleReload waits 5 seconds after being called, and then rebuilds the SwitchMaps



## <a name="HandleRemoteWrapper">func</a> [HandleRemoteWrapper](https://github.com/cognusion/go-jar/tree/master/errors.go?s=7612:7714#L243)
``` go
func HandleRemoteWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool
```
HandleRemoteWrapper wraps errors (HTTP codes >= 400) in a pretty wrapper for client presentation,
using a Worker to make a subrequest to an error-wrapping API



## <a name="HandleTemplateWrapper">func</a> [HandleTemplateWrapper](https://github.com/cognusion/go-jar/tree/master/errors.go?s=7061:7165#L225)
``` go
func HandleTemplateWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool
```
HandleTemplateWrapper wraps errors (HTTP codes >= 400) in a pretty wrapper for client presentation,
using a template



## <a name="InitConfig">func</a> [InitConfig](https://github.com/cognusion/go-jar/tree/master/config.go?s=3170:3200#L72)
``` go
func InitConfig() *viper.Viper
```
InitConfig creates an config, initialized with defaults and environment-set values, and returns it



## <a name="LoadConfig">func</a> [LoadConfig](https://github.com/cognusion/go-jar/tree/master/config.go?s=3526:3586#L92)
``` go
func LoadConfig(configFilename string, v *viper.Viper) error
```
LoadConfig read the config file and returns a config object or an error



## <a name="LogInit">func</a> [LogInit](https://github.com/cognusion/go-jar/tree/master/log.go?s=1744:1764#L62)
``` go
func LogInit() error
```
LogInit initializes all of the loggers based on Conf settings



## <a name="MinuteDelayer">func</a> [MinuteDelayer](https://github.com/cognusion/go-jar/tree/master/debug.go?s=4109:4167#L152)
``` go
func MinuteDelayer(w http.ResponseWriter, r *http.Request)
```
MinuteDelayer is a special finisher that waits for 60s before returning



## <a name="MinuteStreamer">func</a> [MinuteStreamer](https://github.com/cognusion/go-jar/tree/master/debug.go?s=4488:4547#L163)
``` go
func MinuteStreamer(w http.ResponseWriter, r *http.Request)
```
MinuteStreamer is a special finisher that writes the next number, once a secondish, for 60 iterations



## <a name="NewECBDecrypter">func</a> [NewECBDecrypter](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=8072:8125#L275)
``` go
func NewECBDecrypter(b cipher.Block) cipher.BlockMode
```
NewECBDecrypter should never be used unless you know what you're doing



## <a name="NewECBEncrypter">func</a> [NewECBEncrypter](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=7356:7409#L248)
``` go
func NewECBEncrypter(b cipher.Block) cipher.BlockMode
```
NewECBEncrypter should never be used unless you know what you're doing



## <a name="OkFinisher">func</a> [OkFinisher](https://github.com/cognusion/go-jar/tree/master/debug.go?s=4299:4354#L158)
``` go
func OkFinisher(w http.ResponseWriter, r *http.Request)
```
OkFinisher is a Finisher that simply returns "Ok", for throughput testing.



## <a name="PoolLister">func</a> [PoolLister](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=3971:4026#L141)
``` go
func PoolLister(w http.ResponseWriter, r *http.Request)
```
PoolLister is a finisher to list the pools



## <a name="PoolMemberAdder">func</a> [PoolMemberAdder](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=5153:5213#L189)
``` go
func PoolMemberAdder(w http.ResponseWriter, r *http.Request)
```
PoolMemberAdder is a finisher to add a member to an existing pool



## <a name="PoolMemberLister">func</a> [PoolMemberLister](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=4210:4271#L151)
``` go
func PoolMemberLister(w http.ResponseWriter, r *http.Request)
```
PoolMemberLister is a finisher to list the members of an existing pool



## <a name="PoolMemberLoser">func</a> [PoolMemberLoser](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=6650:6710#L244)
``` go
func PoolMemberLoser(w http.ResponseWriter, r *http.Request)
```
PoolMemberLoser is a finisher to remove a member from an existing pool



## <a name="PrettyPrint">func</a> [PrettyPrint](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=4583:4621#L159)
``` go
func PrettyPrint(v interface{}) string
```
PrettyPrint returns the a JSONified version of the string, or %+v if that's not possible



## <a name="ReaderToString">func</a> [ReaderToString](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=5112:5151#L180)
``` go
func ReaderToString(r io.Reader) string
```
ReaderToString reads from a Reader into a Buffer, and then returns the string value of that



## <a name="RealAddr">func</a> [RealAddr](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=3964:4006#L133)
``` go
func RealAddr(h http.Handler) http.Handler
```
RealAddr is a special handler to grab the most probable "real" client address



## <a name="Recoverer">func</a> [Recoverer](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=6974:7020#L232)
``` go
func Recoverer(next http.Handler) http.Handler
```
Recoverer is a wrapping handler to make panic-capable handlers safer



## <a name="ReplaceURI">func</a> [ReplaceURI](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=3889:3949#L141)
``` go
func ReplaceURI(r *http.Request, urlPath, requestURI string)
```
ReplaceURI standardizes the replacement of the Request.URL.Path and Request.RequestURI, which are squirrely at best.



## <a name="RequestErrorResponse">func</a> [RequestErrorResponse](https://github.com/cognusion/go-jar/tree/master/errors.go?s=3240:3331#L109)
``` go
func RequestErrorResponse(r *http.Request, w http.ResponseWriter, Message string, code int)
```
RequestErrorResponse is the functional equivalent of ErrRequestError .WrappedResponse(..)



## <a name="RequestErrorString">func</a> [RequestErrorString](https://github.com/cognusion/go-jar/tree/master/errors.go?s=3021:3090#L104)
``` go
func RequestErrorString(Request *http.Request, Message string) string
```
RequestErrorString is the functional equivalent of ErrRequestError .String()



## <a name="ResponseHeaders">func</a> [ResponseHeaders](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=10394:10446#L335)
``` go
func ResponseHeaders(next http.Handler) http.Handler
```
ResponseHeaders is a simple piece of middleware that sets configured headers



## <a name="Restart">func</a> [Restart](https://github.com/cognusion/go-jar/tree/master/update.go?s=2576:2628#L100)
``` go
func Restart(w http.ResponseWriter, r *http.Request)
```
Restart signals the server to restart itself



## <a name="RouteIDInspectionHandler">func</a> [RouteIDInspectionHandler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=4592:4653#L153)
``` go
func RouteIDInspectionHandler(next http.Handler) http.Handler
```
RouteIDInspectionHandler checks the Query params for a ROUTEID and shoves it into a cookie



## <a name="S3StreamProxyFinisher">func</a> [S3StreamProxyFinisher](https://github.com/cognusion/go-jar/tree/master/s3proxy.go?s=1841:1907#L75)
``` go
func S3StreamProxyFinisher(w http.ResponseWriter, r *http.Request)
```
S3StreamProxyFinisher is a finisher that streams a POSTd file to an S3 bucket



## <a name="SetupHandler">func</a> [SetupHandler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=5302:5351#L177)
``` go
func SetupHandler(next http.Handler) http.Handler
```
SetupHandler adds the RequestID and various other informatives to a request context



## <a name="Stack">func</a> [Stack](https://github.com/cognusion/go-jar/tree/master/health.go?s=5188:5238#L196)
``` go
func Stack(w http.ResponseWriter, r *http.Request)
```
Stack is a Finisher that dumps the current stack to the request



## <a name="StringIfCtx">func</a> [StringIfCtx](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=2010:2068#L76)
``` go
func StringIfCtx(r *http.Request, name interface{}) string
```
StringIfCtx will return a non-empty string if the suppled Request
has a Context.WithValue() of the specified name



## <a name="SwitchHandler">func</a> [SwitchHandler](https://github.com/cognusion/go-jar/tree/master/urlswitch.go?s=2044:2094#L84)
``` go
func SwitchHandler(next http.Handler) http.Handler
```
SwitchHandler adds URL switching information to the request context



## <a name="TestFinisher">func</a> [TestFinisher](https://github.com/cognusion/go-jar/tree/master/debug.go?s=1347:1404#L65)
``` go
func TestFinisher(w http.ResponseWriter, r *http.Request)
```
TestFinisher is a special finisher that outputs some detectables



## <a name="TrimPrefixURI">func</a> [TrimPrefixURI](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=3616:3666#L136)
``` go
func TrimPrefixURI(r *http.Request, prefix string)
```
TrimPrefixURI standardizes the prefix trimming of the Request.URL.Path and Request.RequestURI, which are squirrely at best.



## <a name="URLCaptureHandler">func</a> [URLCaptureHandler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=2990:3044#L104)
``` go
func URLCaptureHandler(next http.Handler) http.Handler
```
URLCaptureHandler is an unchainable handler that captures the Hostname of the Pool Member servicing a request



## <a name="Unzip">func</a> [Unzip](https://github.com/cognusion/go-jar/tree/master/update.go?s=4694:4728#L181)
``` go
func Unzip(src, dest string) error
```
Unzip takes a source zip, and a destination folder, and unzips source into dest,
returning an error if appropriate



## <a name="Update">func</a> [Update](https://github.com/cognusion/go-jar/tree/master/update.go?s=2260:2311#L89)
``` go
func Update(w http.ResponseWriter, r *http.Request)
```
Update signals the updater to update itself



## <a name="ValidateExtras">func</a> [ValidateExtras](https://github.com/cognusion/go-jar/tree/master/config.go?s=8894:8923#L166)
``` go
func ValidateExtras() []error
```
ValidateExtras runs through a list of referenced functions, and returns any errors they return.
All functions will be run, so an array of errors may be returned



## <a name="WithRqID">func</a> [WithRqID](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=16676:16744#L543)
``` go
func WithRqID(ctx context.Context, requestID string) context.Context
```
WithRqID returns a context which knows its request ID



## <a name="WithSessionID">func</a> [WithSessionID](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=16868:16941#L548)
``` go
func WithSessionID(ctx context.Context, sessionID string) context.Context
```
WithSessionID returns a context which knows its session ID




## <a name="Access">type</a> [Access](https://github.com/cognusion/go-jar/tree/master/access.go?s=217:391#L15)
``` go
type Access struct {
    // contains filtered or unexported fields
}

```
Access is a type to provide binary validation of
addresses, based on the contents of "Allow/Deny" rules.







### <a name="NewAccessFromStrings">func</a> [NewAccessFromStrings](https://github.com/cognusion/go-jar/tree/master/access.go?s=585:647#L27)
``` go
func NewAccessFromStrings(allow, deny string) (*Access, error)
```
NewAccessFromStrings is the safest way to create a safe, valid Access
type. The supplied "allow" and "deny" strings may be comma-delimited
lists of IP addresses and/or CIDR networks.





### <a name="Access.AccessHandler">func</a> (\*Access) [AccessHandler](https://github.com/cognusion/go-jar/tree/master/access.go?s=2400:2462#L107)
``` go
func (a *Access) AccessHandler(next http.Handler) http.Handler
```
AccessHandler is a handler that consults r.RemoteAddr and validates
it against the Access type.




### <a name="Access.AddAddress">func</a> (\*Access) [AddAddress](https://github.com/cognusion/go-jar/tree/master/access.go?s=1388:1449#L54)
``` go
func (a *Access) AddAddress(address string, allow bool) error
```
AddAddress adds the supplied address to either the allow or deny
lists, depending on the value of the suppled boolean. An error is
returned if the supplied address cannot be parsed.




### <a name="Access.Validate">func</a> (\*Access) [Validate](https://github.com/cognusion/go-jar/tree/master/access.go?s=3158:3204#L128)
``` go
func (a *Access) Validate(address string) bool
```
Validate tests the supplied address against the Access type,
returning boolean




## <a name="AccessLog">type</a> [AccessLog](https://github.com/cognusion/go-jar/tree/master/log.go?s=4835:5395#L157)
``` go
type AccessLog interface {
    // CommonLogFormat will return the contents as a CLF-compatible string. If combined is set, a "combined" CLF is included (adds referer and user-agent)
    CommonLogFormat(combined bool) string
    // Reset will empty out the contents of the access log
    Reset()
    // ResponseFiller adds response information to the AccessLog entry
    ResponseFiller(responseTime time.Time, responseDuration time.Duration, responseCode int, responseLength int)
    // RequestFiller adds request information to the AccessLog entry
    RequestFiller(r *http.Request)
}
```
AccessLog is an interface providing base logging, but allowing addons to extent it easily










## <a name="BasicAuth">type</a> [BasicAuth](https://github.com/cognusion/go-jar/tree/master/basicauth.go?s=1063:1208#L40)
``` go
type BasicAuth struct {
    // List of allowed users
    Users []string
    // contains filtered or unexported fields
}

```
BasicAuth wraps a handler requiring HTTP basic auth







### <a name="NewBasicAuth">func</a> [NewBasicAuth](https://github.com/cognusion/go-jar/tree/master/basicauth.go?s=1304:1370#L50)
``` go
func NewBasicAuth(source, realm string, users []string) *BasicAuth
```
NewBasicAuth takes a source, realm, and list of users, returning an initialized *BasicAuth


### <a name="NewVerifiedBasicAuth">func</a> [NewVerifiedBasicAuth](https://github.com/cognusion/go-jar/tree/master/basicauth.go?s=1807:1890#L71)
``` go
func NewVerifiedBasicAuth(source, realm string, users []string) (*BasicAuth, error)
```
NewVerifiedBasicAuth takes a source, realm, and list of users, verifies the auth source, and returns an initialized *BasicAuth or an error





### <a name="BasicAuth.Authenticate">func</a> (\*BasicAuth) [Authenticate](https://github.com/cognusion/go-jar/tree/master/basicauth.go?s=2815:2886#L113)
``` go
func (b *BasicAuth) Authenticate(username, password, realm string) bool
```
Authenticate takes a username, password, realm, and return bool if the authentication is positive




### <a name="BasicAuth.Load">func</a> (\*BasicAuth) [Load](https://github.com/cognusion/go-jar/tree/master/basicauth.go?s=5710:5742#L203)
``` go
func (b *BasicAuth) Load() error
```
Load prepares any pre-auth dancing, caching, etc. necessary




### <a name="BasicAuth.VerifySource">func</a> (\*BasicAuth) [VerifySource](https://github.com/cognusion/go-jar/tree/master/basicauth.go?s=2413:2453#L98)
``` go
func (b *BasicAuth) VerifySource() error
```
VerifySource checks that the requested authentication source is valid, and accessible




## <a name="BodyByteLimit">type</a> [BodyByteLimit](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=15588:15634#L507)
``` go
type BodyByteLimit struct {
    // contains filtered or unexported fields
}

```
BodyByteLimit is a Request.Body size limiter







### <a name="NewBodyByteLimit">func</a> [NewBodyByteLimit](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=15693:15741#L512)
``` go
func NewBodyByteLimit(limit int64) BodyByteLimit
```
NewBodyByteLimit returns an initialized BodyByteLimit





### <a name="BodyByteLimit.Handler">func</a> (\*BodyByteLimit) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=15819:15882#L517)
``` go
func (b *BodyByteLimit) Handler(next http.Handler) http.Handler
```
Handler limits the size of Request.Body




## <a name="CORS">type</a> [CORS](https://github.com/cognusion/go-jar/tree/master/cors.go?s=2529:2736#L66)
``` go
type CORS struct {
    AllowCredentials string
    AllowMethods     string
    AllowHeaders     string
    MaxAge           string
    // contains filtered or unexported fields
}

```
CORS is an abstraction to handle CORS header nonsense.
In order to keep origin comparisons as fast as possible, the expressions are pre-compiled,
and thus need to either be added via AddOrigins() or supplied to NewCORSFromConfig().







### <a name="NewCORS">func</a> [NewCORS](https://github.com/cognusion/go-jar/tree/master/cors.go?s=2785:2805#L77)
``` go
func NewCORS() *CORS
```
NewCORS returns an initialized CORS struct.


### <a name="NewCORSFromConfig">func</a> [NewCORSFromConfig](https://github.com/cognusion/go-jar/tree/master/cors.go?s=2971:3050#L85)
``` go
func NewCORSFromConfig(origins []string, conf map[string]string) (*CORS, error)
```
NewCORSFromConfig returns an initialized CORS struct from a list of origins and a config map





### <a name="CORS.AddOrigin">func</a> (\*CORS) [AddOrigin](https://github.com/cognusion/go-jar/tree/master/cors.go?s=3365:3413#L99)
``` go
func (c *CORS) AddOrigin(origins []string) error
```
AddOrigin adds an origin expression to the CORS struct




### <a name="CORS.Handler">func</a> (\*CORS) [Handler](https://github.com/cognusion/go-jar/tree/master/cors.go?s=3856:3910#L115)
``` go
func (c *CORS) Handler(next http.Handler) http.Handler
```
Handler is a middleware that validates Origin request headers against
a whitelist of expressions, and may change the response headers accordingly




### <a name="CORS.ResponseModifier">func</a> (\*CORS) [ResponseModifier](https://github.com/cognusion/go-jar/tree/master/cors.go?s=5379:5437#L167)
``` go
func (c *CORS) ResponseModifier(resp *http.Response) error
```
ResponseModifier is an oxy/forward opsetter to remove CORS headers from responses




### <a name="CORS.String">func</a> (\*CORS) [String](https://github.com/cognusion/go-jar/tree/master/cors.go?s=5017:5047#L157)
``` go
func (c *CORS) String() string
```



## <a name="Cert">type</a> [Cert](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=3958:4029#L115)
``` go
type Cert struct {
    Domain   string
    Keyfile  string
    Certfile string
}

```
Cert encapsulated a Domain, the Keyfile, and a Certfile










## <a name="Compression">type</a> [Compression](https://github.com/cognusion/go-jar/tree/master/compression.go?s=225:275#L14)
``` go
type Compression struct {
    // contains filtered or unexported fields
}

```
Compression is used to support GZIP compression of data en route to a client







### <a name="NewCompression">func</a> [NewCompression](https://github.com/cognusion/go-jar/tree/master/compression.go?s=376:431#L19)
``` go
func NewCompression(contentTypes []string) *Compression
```
NewCompression returns a pointer to a Compression struct with the specified MIME-types baked in





### <a name="Compression.Handler">func</a> (\*Compression) [Handler](https://github.com/cognusion/go-jar/tree/master/compression.go?s=553:614#L24)
``` go
func (c *Compression) Handler(next http.Handler) http.Handler
```
Handler is a middleware to potentially GZIP-compress outgoing response bodies




## <a name="ConfigKey">type</a> [ConfigKey](https://github.com/cognusion/go-jar/tree/master/config.go?s=3043:3066#L69)
``` go
type ConfigKey = string
```
ConfigKey is a string type for static config key name consistency










## <a name="ConsistentHashPool">type</a> [ConsistentHashPool](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=772:944#L32)
``` go
type ConsistentHashPool struct {
    // contains filtered or unexported fields
}

```
ConsistentHashPool is a PoolManager that implements a consistent hash on a key to return
the proper member consistently







### <a name="NewConsistentHashPool">func</a> [NewConsistentHashPool](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=1007:1113#L41)
``` go
func NewConsistentHashPool(source, key string, pool *Pool, next http.Handler) (*ConsistentHashPool, error)
```
NewConsistentHashPool returns a primed ConsistentHashPool


### <a name="NewConsistentHashPoolOpts">func</a> [NewConsistentHashPoolOpts](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=1291:1454#L46)
``` go
func NewConsistentHashPoolOpts(source, key string, partitionCount, replicationFactor int, load float64, pool *Pool, next http.Handler) (*ConsistentHashPool, error)
```
NewConsistentHashPoolOpts exposes some internal tunables, but still returns a ConsistentHashPool





### <a name="ConsistentHashPool.Next">func</a> (\*ConsistentHashPool) [Next](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=3715:3764#L135)
``` go
func (ch *ConsistentHashPool) Next() http.Handler
```
Next returns the specified next Handler




### <a name="ConsistentHashPool.NextServer">func</a> (\*ConsistentHashPool) [NextServer](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=3554:3614#L130)
``` go
func (ch *ConsistentHashPool) NextServer() (*url.URL, error)
```
NextServer is an error-causing noop to implement PoolManager




### <a name="ConsistentHashPool.RemoveServer">func</a> (\*ConsistentHashPool) [RemoveServer](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=3087:3147#L110)
``` go
func (ch *ConsistentHashPool) RemoveServer(u *url.URL) error
```
RemoveServer removes the specified member from the pool




### <a name="ConsistentHashPool.ServeHTTP">func</a> (\*ConsistentHashPool) [ServeHTTP](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=2319:2400#L88)
``` go
func (ch *ConsistentHashPool) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
ServeHTTP handles its part of the request




### <a name="ConsistentHashPool.ServerWeight">func</a> (\*ConsistentHashPool) [ServerWeight](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=2939:3005#L105)
``` go
func (ch *ConsistentHashPool) ServerWeight(u *url.URL) (int, bool)
```
ServerWeight is a noop to implement PoolManager




### <a name="ConsistentHashPool.Servers">func</a> (\*ConsistentHashPool) [Servers](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=2090:2140#L78)
``` go
func (ch *ConsistentHashPool) Servers() []*url.URL
```
Servers returns a list of member URLs




### <a name="ConsistentHashPool.UpsertServer">func</a> (\*ConsistentHashPool) [UpsertServer](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=3251:3347#L116)
``` go
func (ch *ConsistentHashPool) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
```
UpsertServer adds or updates the member to the pool




## <a name="CorsString">type</a> [CorsString](https://github.com/cognusion/go-jar/tree/master/cors.go?s=1051:1075#L33)
``` go
type CorsString = string
```
CorsString is a string type for static string consistency










## <a name="DebugTrip">type</a> [DebugTrip](https://github.com/cognusion/go-jar/tree/master/debug.go?s=4929:5135#L182)
``` go
type DebugTrip struct {
    // RTFunc is executed when RoundTrip() is called on a request.
    // It can be changed at any point to aid in changing conditions
    RTFunc func(*http.Request) (*http.Response, error)
}

```
DebugTrip is an http.RoundTripper with a pluggable core func to aid in debugging










### <a name="DebugTrip.RoundTrip">func</a> (\*DebugTrip) [RoundTrip](https://github.com/cognusion/go-jar/tree/master/debug.go?s=5174:5244#L189)
``` go
func (d *DebugTrip) RoundTrip(r *http.Request) (*http.Response, error)
```
RoundTrip is the Request executor




## <a name="ErrConfigurationError">type</a> [ErrConfigurationError](https://github.com/cognusion/go-jar/tree/master/errors.go?s=2455:2508#L84)
``` go
type ErrConfigurationError struct {
    // contains filtered or unexported fields
}

```
ErrConfigurationError is returned when a debilitating configuration error
occurs. If this is the initial configuration load, the program should exit.
If this is a reload, the reload should abort and the known-working configuration
should persist










### <a name="ErrConfigurationError.Error">func</a> (ErrConfigurationError) [Error](https://github.com/cognusion/go-jar/tree/master/errors.go?s=2510:2555#L88)
``` go
func (e ErrConfigurationError) Error() string
```



## <a name="ErrRequestError">type</a> [ErrRequestError](https://github.com/cognusion/go-jar/tree/master/errors.go?s=2869:2939#L98)
``` go
type ErrRequestError struct {
    Request *http.Request
    Message string
}

```
ErrRequestError should be returned whenever an error is returned to
a requestor. Care should be taken not to expose dynamic information inside
the message. The request id will be automatically added to the message










### <a name="ErrRequestError.Bytes">func</a> (ErrRequestError) [Bytes](https://github.com/cognusion/go-jar/tree/master/errors.go?s=3430:3469#L114)
``` go
func (e ErrRequestError) Bytes() []byte
```
Bytes returns a []byte of the error




### <a name="ErrRequestError.Error">func</a> (ErrRequestError) [Error](https://github.com/cognusion/go-jar/tree/master/errors.go?s=3644:3683#L124)
``` go
func (e ErrRequestError) Error() string
```
Error returns a string of the error




### <a name="ErrRequestError.String">func</a> (ErrRequestError) [String](https://github.com/cognusion/go-jar/tree/master/errors.go?s=3541:3581#L119)
``` go
func (e ErrRequestError) String() string
```
String returns a string of the error




### <a name="ErrRequestError.WrappedResponse">func</a> (ErrRequestError) [WrappedResponse](https://github.com/cognusion/go-jar/tree/master/errors.go?s=4034:4107#L137)
``` go
func (e ErrRequestError) WrappedResponse(code int, w http.ResponseWriter)
```
WrappedResponse writes the templatized version of the error to a PRW




## <a name="Error">type</a> [Error](https://github.com/cognusion/go-jar/tree/master/errors.go?s=1245:1262#L44)
``` go
type Error string
```
Error is an error type










### <a name="Error.Error">func</a> (Error) [Error](https://github.com/cognusion/go-jar/tree/master/errors.go?s=1314:1343#L47)
``` go
func (e Error) Error() string
```
Error returns the stringified version of Error




## <a name="ErrorWrapper">type</a> [ErrorWrapper](https://github.com/cognusion/go-jar/tree/master/errors.go?s=4775:5089#L158)
``` go
type ErrorWrapper struct {
    // E takes the error code, request, a PluggableResponseWriter, and the original body,
    // and returns boolean true IFF rw has been written to. E should not change
    // headers as they may be ignored.
    E func(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool
}

```
An ErrorWrapper is a struct to abstract error wrapping










### <a name="ErrorWrapper.Handler">func</a> (\*ErrorWrapper) [Handler](https://github.com/cognusion/go-jar/tree/master/errors.go?s=5152:5214#L166)
``` go
func (e *ErrorWrapper) Handler(next http.Handler) http.Handler
```
Handler is the chainable handler that will wrap the error




## <a name="FinisherMap">type</a> [FinisherMap](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=784:828#L34)
``` go
type FinisherMap map[string]http.HandlerFunc
```
FinisherMap maps Finisher names to their HandlerFuncs










### <a name="FinisherMap.List">func</a> (\*FinisherMap) [List](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=880:917#L37)
``` go
func (h *FinisherMap) List() []string
```
List returns the names of all of the Finishers




## <a name="ForbiddenPaths">type</a> [ForbiddenPaths](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=11718:11825#L376)
``` go
type ForbiddenPaths struct {
    // Paths is a list of compiled Regexps, because speed
    Paths []*pcre.Regexp
}

```
ForbiddenPaths is a struct to assist in the expedient resolution of determining if a Request is destined to a forbidden path







### <a name="NewForbiddenPaths">func</a> [NewForbiddenPaths](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=12004:12067#L383)
``` go
func NewForbiddenPaths(paths []string) (*ForbiddenPaths, error)
```
NewForbiddenPaths takes a list of regexp-compatible strings, and returns the analogous ForbiddenPaths with compiled regexps,
or an error if a regexp could not be compiled





### <a name="ForbiddenPaths.Handler">func</a> (\*ForbiddenPaths) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=12487:12551#L401)
``` go
func (f *ForbiddenPaths) Handler(next http.Handler) http.Handler
```
Handler is a middleware that checks the request URI against regexps and 403's if match




## <a name="GenericResponse">type</a> [GenericResponse](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=3628:3688#L127)
``` go
type GenericResponse struct {
    Message string
    Code    int
}

```
GenericResponse is a Finisher that returns a possibly-wrapped response










### <a name="GenericResponse.Finisher">func</a> (\*GenericResponse) [Finisher](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=3757:3832#L133)
``` go
func (gr *GenericResponse) Finisher(w http.ResponseWriter, r *http.Request)
```
Finisher is a ... Finisher for the instantiated GenericResponse




## <a name="HTTPWork">type</a> [HTTPWork](https://github.com/cognusion/go-jar/tree/master/workers.go?s=1597:1989#L55)
``` go
type HTTPWork struct {
    Client       *http.Client
    Request      *http.Request
    ResponseChan chan interface{}
    // RetryCount is the number of times to retry Request if there is an error
    RetryCount int
    //RetryInterval is the duration between retries
    RetryInterval time.Duration
    //RetryHTTPErrors, if set, classifies HTTP responses >= 500 as errors for retry purposes
    RetryHTTPErrors bool
}

```
HTTPWork is a generic Work that can make HTTP requests










### <a name="HTTPWork.Return">func</a> (\*HTTPWork) [Return](https://github.com/cognusion/go-jar/tree/master/workers.go?s=2516:2561#L93)
``` go
func (h *HTTPWork) Return(rthing interface{})
```
Return is called response with results




### <a name="HTTPWork.Work">func</a> (\*HTTPWork) [Work](https://github.com/cognusion/go-jar/tree/master/workers.go?s=2020:2057#L68)
``` go
func (h *HTTPWork) Work() interface{}
```
Work is called to do work




## <a name="HandlerMap">type</a> [HandlerMap](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=1453:1511#L59)
``` go
type HandlerMap map[string]func(http.Handler) http.Handler
```
HandlerMap maps handler names to their funcs










### <a name="HandlerMap.List">func</a> (\*HandlerMap) [List](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=1562:1598#L62)
``` go
func (h *HandlerMap) List() []string
```
List returns the names of all of the Handlers




## <a name="HealthCheckError">type</a> [HealthCheckError](https://github.com/cognusion/go-jar/tree/master/pools.go?s=10244:10447#L360)
``` go
type HealthCheckError struct {
    PoolName    string
    URL         string
    StatusCode  int
    Prune       bool
    ErrorStatus HealthCheckStatus
    Add         PruneFunc
    Remove      PruneFunc
    Err         error
}

```
HealthCheckError is an error returned through the HealthCheck system










### <a name="HealthCheckError.Error">func</a> (\*HealthCheckError) [Error](https://github.com/cognusion/go-jar/tree/master/pools.go?s=10503:10544#L372)
``` go
func (h *HealthCheckError) Error() string
```
Error returns the stringified version of the error




## <a name="HealthCheckResult">type</a> [HealthCheckResult](https://github.com/cognusion/go-jar/tree/master/pools.go?s=10691:10839#L377)
``` go
type HealthCheckResult struct {
    PoolName   string
    URL        string
    StatusCode int
    Prune      bool
    Add        PruneFunc
    Remove     PruneFunc
}

```
HealthCheckResult is a non-error returned through the HealthCheck system










## <a name="HealthCheckStatus">type</a> [HealthCheckStatus](https://github.com/cognusion/go-jar/tree/master/health.go?s=3286:3312#L117)
``` go
type HealthCheckStatus int
```
HealthCheckStatus is a specific int for HealthCheckStatus consts


``` go
const (
    Unknown HealthCheckStatus = iota
    Ok
    Warning
    Critical
)
```
Constants for HealthCheckStatuses







### <a name="StringToHealthCheckStatus">func</a> [StringToHealthCheckStatus](https://github.com/cognusion/go-jar/tree/master/health.go?s=3844:3912#L142)
``` go
func StringToHealthCheckStatus(hc string) (HealthCheckStatus, error)
```
StringToHealthCheckStatus takes a string HealthCheckStatus and returns the HealthCheckStatus or ErrNoSuchHealthCheckStatus





### <a name="HealthCheckStatus.String">func</a> (HealthCheckStatus) [String](https://github.com/cognusion/go-jar/tree/master/health.go?s=3314:3356#L119)
``` go
func (i HealthCheckStatus) String() string
```



## <a name="HealthCheckWork">type</a> [HealthCheckWork](https://github.com/cognusion/go-jar/tree/master/pools.go?s=10889:11150#L387)
``` go
type HealthCheckWork struct {
    PoolName    string
    Member      string
    URL         string
    Prune       bool
    ErrorStatus HealthCheckStatus
    Add         PruneFunc
    Remove      PruneFunc
    // Return is an error, or the StatusCode int
    ReturnChan chan interface{}
}

```
HealthCheckWork is Work to run a HealthCheck










### <a name="HealthCheckWork.Return">func</a> (\*HealthCheckWork) [Return](https://github.com/cognusion/go-jar/tree/master/pools.go?s=12474:12526#L455)
``` go
func (h *HealthCheckWork) Return(rthing interface{})
```
Return consumes a Work result and slides it downthe return channel




### <a name="HealthCheckWork.Work">func</a> (\*HealthCheckWork) [Work](https://github.com/cognusion/go-jar/tree/master/pools.go?s=11235:11279#L400)
``` go
func (h *HealthCheckWork) Work() interface{}
```
Work executes the HealthCheck and returns HealthCheckResult or HealthCheckError




## <a name="JSONAccessLog">type</a> [JSONAccessLog](https://github.com/cognusion/go-jar/tree/master/log.go?s=5465:6194#L169)
``` go
type JSONAccessLog struct {
    Timestamp     string `json:"timestamp"`
    Hostname      string `json:"hostname"`
    RemoteAddress string `json:"remoteaddr"`
    User          string `json:"user"`
    XForwardedFor string `json:"x-forwarded-for"`
    ClientIP      string `json:"clientip"`
    Method        string `json:"method"`
    Request       string `json:"request"`
    Status        string `json:"status"`
    Bytes         string `json:"bytes"`
    UserAgent     string `json:"user-agent"`
    Duration      string `json:"duration"`
    Referer       string `json:"referer"`
    Message       string `json:"message"`
    RequestID     string `json:"requestid"`
    Proto         string `json:"proto"`
    TLSVersion    string `json:"tlsversion"`
    // contains filtered or unexported fields
}

```
JSONAccessLog is an AccessLog uberstruct for JSONifying log data










### <a name="JSONAccessLog.CommonLogFormat">func</a> (\*JSONAccessLog) [CommonLogFormat](https://github.com/cognusion/go-jar/tree/master/log.go?s=6347:6408#L191)
``` go
func (a *JSONAccessLog) CommonLogFormat(combined bool) string
```
CommonLogFormat will return the contents as a CLF-compatible string. If combined is set, a "combined" CLF is included (adds referer and user-agent)




### <a name="JSONAccessLog.RequestFiller">func</a> (\*JSONAccessLog) [RequestFiller](https://github.com/cognusion/go-jar/tree/master/log.go?s=7924:7978#L251)
``` go
func (a *JSONAccessLog) RequestFiller(r *http.Request)
```
RequestFiller adds request information to the AccessLog entry




### <a name="JSONAccessLog.Reset">func</a> (\*JSONAccessLog) [Reset](https://github.com/cognusion/go-jar/tree/master/log.go?s=7070:7101#L218)
``` go
func (a *JSONAccessLog) Reset()
```
Reset will empty out the contents of the access log




### <a name="JSONAccessLog.ResponseFiller">func</a> (\*JSONAccessLog) [ResponseFiller](https://github.com/cognusion/go-jar/tree/master/log.go?s=7482:7601#L240)
``` go
func (a *JSONAccessLog) ResponseFiller(endtime time.Time, duration time.Duration, responseCode int, responseLength int)
```
ResponseFiller adds response information to the AccessLog entry




## <a name="Member">type</a> [Member](https://github.com/cognusion/go-jar/tree/master/pool.go?s=5154:5259#L141)
``` go
type Member struct {
    URL     *url.URL
    Address string
    AZ      string
    // contains filtered or unexported fields
}

```
Member is an attribute struct to describe a Pool Member










### <a name="Member.String">func</a> (\*Member) [String](https://github.com/cognusion/go-jar/tree/master/poolch.go?s=1988:2020#L73)
``` go
func (m *Member) String() string
```
String returns the Address of the Member




## <a name="NoopResponseWriter">type</a> [NoopResponseWriter](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=5456:5522#L192)
``` go
type NoopResponseWriter struct {
    // contains filtered or unexported fields
}

```
NoopResponseWriter is a hack to support a Response with a status and headers,
but no body. This is almost never what you want. Really.







### <a name="NewNoopResponseWriter">func</a> [NewNoopResponseWriter](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=5628:5675#L199)
``` go
func NewNoopResponseWriter() NoopResponseWriter
```
NewNoopResponseWriter returns a NoopResponseWriter that you almost definitely
do not want to use.





### <a name="NoopResponseWriter.Header">func</a> (\*NoopResponseWriter) [Header](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=5774:5823#L206)
``` go
func (n *NoopResponseWriter) Header() http.Header
```
Header returns an http.Header




### <a name="NoopResponseWriter.Write">func</a> (\*NoopResponseWriter) [Write](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=5980:6041#L212)
``` go
func (n *NoopResponseWriter) Write(bytes []byte) (int, error)
```
Write completely ignores whatever you've written, but lies and
returns the size of whatever you wrote to it, and never an error




### <a name="NoopResponseWriter.WriteHeader">func</a> (\*NoopResponseWriter) [WriteHeader](https://github.com/cognusion/go-jar/tree/master/helpers.go?s=6164:6220#L220)
``` go
func (n *NoopResponseWriter) WriteHeader(statusCode int)
```
WriteHeader changes the response code




## <a name="Path">type</a> [Path](https://github.com/cognusion/go-jar/tree/master/paths.go?s=482:3898#L24)
``` go
type Path struct {
    // Name is an optional "name" for the path. Will be output in some logs. If not set, will use an index number
    Name string
    // Path is a URI prefix to match
    Path string
    // Absolute declares if Path should be absolute instead of as a prefix
    Absolute bool
    // Allow
    Allow string
    // Deny
    Deny string
    // Host is a hostname or hostname-pattern to restrict this Path too
    Host string
    // Hosts is a list of hostnames or hostname-patterns to restrict this Path too.
    // Will result in one actual Path per entry, which is almost always fine.
    Hosts []string
    // Methods is a list of HTTP methods to restrict this path to
    Methods []string
    // Headers is a list of HTTP Request headers to restrict this path to
    Headers []string
    // Handlers is an ordered list of http.Handlers to apply
    Handlers []string
    // Pool is an actual Pool to handle the proxying. Mutually exclusive with Finisher
    Pool string
    // Finisher is the final handler. Mutually exclusive with Pool
    Finisher string
    // RateLimit each IP to these many requests/second. Also must have the "RateLimiter" handler, or it will be appended to the chain
    RateLimit float64
    // RateLimitPurge is a duration where a limit gets dumped
    RateLimitPurge time.Duration
    // RateLimitCollectOnly sets if the ratelimiter should only collect and log, versus enforce
    RateLimitCollectOnly bool
    // BodyByteLimit is the maximum number of bytes a Request.Body is allowed to be. It is poor form to set this unless the Path is terminated by
    // a finisher that will otherwise consume the Request.Body and possibly OOM and/or overuse disk space.
    BodyByteLimit int64
    // Redirect is a special Finisher. "%1" may be used to optionally denote the request path.
    // e.g. Redirect http://somewhereelse.com%1
    Redirect string
    // RedirectCode is an optional code to send as the redirect status
    RedirectCode int
    // RedirectHostMatch is a Perl-Compatible Regular Expression with grouping to apply to the Hostname, replacing $1,$2, etc. in ``Redirect``
    RedirectHostMatch string
    // ReplacePath is used to replace the requested path with the target path
    ReplacePath string
    // StripPrefix is used to replace the requested path with one sans prefix
    StripPrefix string
    // BrowserExclusions is a list of browsers disallowed down this path, based on best-effort analysis of request headers
    BrowserExclusions []string
    // ForbiddenPaths is a list of path prefixes that will result in a 403, while traversing this path
    ForbiddenPaths []string
    // Timeout is a path-specific override of how long a request and response may take on this path
    Timeout time.Duration
    // BasicAuthRealm is the name of the HTTP Auth Realm on this Path. Need not be unique. Should not be empty.
    BasicAuthRealm string
    // BasicAuthSource is a URL to specify where HTTP Basic Auth information should come from (file://). Setting this forces auth
    BasicAuthSource string
    // BasicAuthUsers is a list of usernames allowed on this Path. Default is "all"
    BasicAuthUsers []string
    // ErrorMessage is a static message to respond with, if this path is executed
    ErrorMessage string
    // ErrorCode is the HTTP response code that will be returned with ErrorMessage, IFF ErrorMessage is set. Defaults to StatusOK
    ErrorCode int
    // Options is a horrible, brittle map[string]interface{} that some handlers or finishers
    // use for per-path configuration. Avoid if possible.
    Options PathOptions
}

```
Path is an extensible struct, detailing its configuration










## <a name="PathHandler">type</a> [PathHandler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=14512:14576#L469)
``` go
type PathHandler struct {
    Path    string
    Options PathOptions
}

```
PathHandler is a wrapping struct to inject the Path name, and any PathOptions into the Context










### <a name="PathHandler.Handler">func</a> (\*PathHandler) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=14670:14731#L475)
``` go
func (p *PathHandler) Handler(next http.Handler) http.Handler
```
Handler is a middleware that injects the Path name, and any PathOptions into the Context




## <a name="PathOptions">type</a> [PathOptions](https://github.com/cognusion/go-jar/tree/master/paths.go?s=3953:3992#L92)
``` go
type PathOptions map[string]interface{}
```
PathOptions is an MSI with a case-agnostic getter










### <a name="PathOptions.Get">func</a> (\*PathOptions) [Get](https://github.com/cognusion/go-jar/tree/master/paths.go?s=4056:4105#L95)
``` go
func (p *PathOptions) Get(key string) interface{}
```
Get returns an interface{} if *key* matches, otherwise nil




### <a name="PathOptions.GetBool">func</a> (\*PathOptions) [GetBool](https://github.com/cognusion/go-jar/tree/master/paths.go?s=4622:4668#L124)
``` go
func (p *PathOptions) GetBool(key string) bool
```
GetBool returns a bool value if *key* matches, otherwise false




### <a name="PathOptions.GetString">func</a> (\*PathOptions) [GetString](https://github.com/cognusion/go-jar/tree/master/paths.go?s=4334:4384#L109)
``` go
func (p *PathOptions) GetString(key string) string
```
GetString returns a string if *key* matches, otherwise empty string




### <a name="PathOptions.GetStringSlice">func</a> (\*PathOptions) [GetStringSlice](https://github.com/cognusion/go-jar/tree/master/paths.go?s=4927:4984#L139)
``` go
func (p *PathOptions) GetStringSlice(key string) []string
```
GetStringSlice returns a []string if *key* matches, otherwise an empty []string




## <a name="PathReplacer">type</a> [PathReplacer](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=13888:13942#L448)
``` go
type PathReplacer struct {
    From string
    To   string
}

```
PathReplacer is a wrapping struct to replace the Request path










### <a name="PathReplacer.Handler">func</a> (\*PathReplacer) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=14002:14064#L454)
``` go
func (p *PathReplacer) Handler(next http.Handler) http.Handler
```
Handler is a middleware that replaces the Request path




## <a name="PathStripper">type</a> [PathStripper](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=13389:13432#L428)
``` go
type PathStripper struct {
    Prefix string
}

```
PathStripper is a wrapping struct to remove the prefix from the Request path










### <a name="PathStripper.Handler">func</a> (\*PathStripper) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=13492:13554#L433)
``` go
func (p *PathStripper) Handler(next http.Handler) http.Handler
```
Handler is a middleware that replaces the Request path




## <a name="Pool">type</a> [Pool](https://github.com/cognusion/go-jar/tree/master/pool.go?s=5307:6297#L149)
``` go
type Pool struct {
    Config *PoolConfig

    // AddMember adds a URI to the loadbalancer. An error is returned if the URI doesn't parse properly
    AddMember func(string) error
    // RemoveMember removes a URI from the loadbalancer, but not from the member cache.
    // ErrNoSuchMemberError is returned if the requested member doesn't exist,
    // or another error if the URI provided doesn't parse properly.
    RemoveMember func(string) error
    // DeleteMember removes a URI from the entire Pool construct,
    // ErrNoSuchMemberError is returned if the requested member doesn't exist,
    // or another error if the URI provided doesn't parse properly.
    DeleteMember func(string) error
    // ListMembers returns a list of URIs for existing members
    ListMembers func() []*url.URL
    // contains filtered or unexported fields
}

```
Pool is a list of like-minded destinations










### <a name="Pool.GetMember">func</a> (\*Pool) [GetMember](https://github.com/cognusion/go-jar/tree/master/pool.go?s=6961:7005#L194)
``` go
func (p *Pool) GetMember(u *url.URL) *Member
```
GetMember interacts with an internal cache, returning a Member from the cache or crafting a new one (and adding it to the cache)




### <a name="Pool.GetPool">func</a> (\*Pool) [GetPool](https://github.com/cognusion/go-jar/tree/master/pool.go?s=6555:6601#L182)
``` go
func (p *Pool) GetPool() (http.Handler, error)
```
GetPool returns the materialized pool or an error. If the Pool has not been
materialized, it does that.




### <a name="Pool.IsMaterialized">func</a> (\*Pool) [IsMaterialized](https://github.com/cognusion/go-jar/tree/master/pool.go?s=6381:6417#L176)
``` go
func (p *Pool) IsMaterialized() bool
```
IsMaterialized return boolean on whether the pool has been materialized or not




### <a name="Pool.Materialize">func</a> (\*Pool) [Materialize](https://github.com/cognusion/go-jar/tree/master/pool.go?s=12993:13043#L365)
``` go
func (p *Pool) Materialize() (http.Handler, error)
```
Materialize returns a Handler that can represent the Pool.

Generally, you should call Pool.GetPool instead, so you can receive
a pointer to the exist materialized pool if it exists, or it will
Materialize it for you.




## <a name="PoolConfig">type</a> [PoolConfig](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=219:2621#L14)
``` go
type PoolConfig struct {
    // Name is what you'd like to call this Pool
    Name string
    // Members is a list of URIs you'd like in the pool
    Members []string
    // Buffered refers to whether you'd like buffer all the requests, to possibly retry them in the even of a Member failure
    Buffered bool
    // BufferedFails is the number of failures to accept before giving up
    BufferedFails int
    // RemoveHeaders is a list of pool-specific headers to remove
    RemoveHeaders []string
    // ConsistentHashing is mutually exclusive to Sticky, and enables automatic distributions
    ConsistentHashing bool
    // ConsistentHashSource is one of "header", "cookie", or "request".
    // For "header" and "cookie", it is paired with ConsistentHashName to choose which key from those maps is used.
    // For "request" it is paired with ConsistentHashName to choose from one of "remoteaddr", "host", and "url".
    ConsistentHashSource string
    // ConsistentHashName sets the request part, header, or cookie name to pull the value from
    ConsistentHashName string
    // Sticky is mutually exclusive to ConsistentHashing, and enables cookie-based session routing
    Sticky bool
    // StickyCookieName overrides the name of the cookie used to handle sticky sessions
    StickyCookieName string
    // StickyCookieType allows for the setting of cookie values to "plain", "hex"-encoded, or "aes"-encrypted
    StickyCookieType string
    // StripPrefix removes the specified string from the front of a URL before processing. Dupes Path.StripPrefix
    StripPrefix string
    // HealthCheckDisabled determines whether or not to healthcheck the members.
    HealthCheckDisabled bool
    // Healthcheck is a URI to check for health. Anything other than a 200 is bad.
    HealthCheckURI string
    // HealthCheckShotgun will disable the adaptive healthcheck scheduler, and fire all of them every interval
    HealthCheckShotgun bool
    // HealthCheckErrorStatus is a string mapping to a const HealthCheckStatus
    HealthCheckErrorStatus string
    // ReplacePath is used to replace the requested path with the target path
    ReplacePath string
    // Prune removes members that fail healthcheck, until they pass again
    Prune bool
    // EC2Affinity specifies whether an EC2-aware JAR should prefer a same-AZ member if available
    EC2Affinity bool
    // Options is a horrible, brittle map[string]interface{} that some PoolManagers
    // use for per-pool configuration. Avoid if possible.
    Options PoolOptions
}

```
PoolConfig is type exposing expected configuration for a pool, abstracted for passing around










## <a name="PoolID">type</a> [PoolID](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=15101:15136#L489)
``` go
type PoolID struct {
    Pool string
}

```
PoolID is a wrapping struct to inject the Pool name into the Context










### <a name="PoolID.Handler">func</a> (\*PoolID) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=15188:15244#L494)
``` go
func (p *PoolID) Handler(next http.Handler) http.Handler
```
Handler injects the Pool name into the Context




## <a name="PoolManager">type</a> [PoolManager](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=4545:4841#L153)
``` go
type PoolManager interface {
    Servers() []*url.URL
    ServeHTTP(w http.ResponseWriter, req *http.Request)
    ServerWeight(u *url.URL) (int, bool)
    RemoveServer(u *url.URL) error
    UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
    NextServer() (*url.URL, error)
    Next() http.Handler
}
```
PoolManager is an interface to encompass oxy/roundrobin and our chpool










## <a name="PoolOptions">type</a> [PoolOptions](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=2676:2715#L61)
``` go
type PoolOptions map[string]interface{}
```
PoolOptions is an MSI with a case-agnostic getter










### <a name="PoolOptions.Get">func</a> (\*PoolOptions) [Get](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=2779:2828#L64)
``` go
func (p *PoolOptions) Get(key string) interface{}
```
Get returns an interface{} if *key* matches, otherwise nil




### <a name="PoolOptions.GetBool">func</a> (\*PoolOptions) [GetBool](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=3902:3948#L123)
``` go
func (p *PoolOptions) GetBool(key string) bool
```
GetBool returns a bool value if *key* matches, otherwise false




### <a name="PoolOptions.GetFloat64">func</a> (\*PoolOptions) [GetFloat64](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=3611:3663#L108)
``` go
func (p *PoolOptions) GetFloat64(key string) float64
```
GetFloat64 returns a float64 if *key* matches, otherwise -1




### <a name="PoolOptions.GetInt">func</a> (\*PoolOptions) [GetInt](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=3335:3379#L93)
``` go
func (p *PoolOptions) GetInt(key string) int
```
GetInt returns an int if *key* matches, otherwise -1




### <a name="PoolOptions.GetString">func</a> (\*PoolOptions) [GetString](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=3057:3107#L78)
``` go
func (p *PoolOptions) GetString(key string) string
```
GetString returns a string if *key* matches, otherwise empty string




### <a name="PoolOptions.GetStringSlice">func</a> (\*PoolOptions) [GetStringSlice](https://github.com/cognusion/go-jar/tree/master/poolconfig.go?s=4207:4264#L138)
``` go
func (p *PoolOptions) GetStringSlice(key string) []string
```
GetStringSlice returns a []string if *key* matches, otherwise an empty []string




## <a name="Pools">type</a> [Pools](https://github.com/cognusion/go-jar/tree/master/pools.go?s=1453:1787#L43)
``` go
type Pools struct {
    sync.RWMutex // Readers must RLock/RUnlock. Writers must Lock/Unlock

    // StopWatch will stop the monitoring of the pool members.
    StopWatch func()
    // contains filtered or unexported fields
}

```
Pools is a goro-safe map of Pool objects, and if interval > 0, will also
healthcheck pool members, managing them accordingly.







### <a name="BuildPools">func</a> [BuildPools](https://github.com/cognusion/go-jar/tree/master/pools.go?s=8778:8810#L307)
``` go
func BuildPools() (*Pools, bool)
```
BuildPools unmarshalls the pools config, creates them, and updates the pool list
ConfigPoolsHealthcheckInterval will set the healthcheck interval for pool members.
Set to 0 to disable.


### <a name="NewPools">func</a> [NewPools](https://github.com/cognusion/go-jar/tree/master/pools.go?s=1942:2022#L54)
``` go
func NewPools(poolConfigs map[string]*PoolConfig, interval time.Duration) *Pools
```
NewPools creates a functioning Pools struct, initialized with the pools, and a healthcheck interval.
Set the interval to 0 to disable healthchecks





### <a name="Pools.Exists">func</a> (\*Pools) [Exists](https://github.com/cognusion/go-jar/tree/master/pools.go?s=4814:4854#L165)
``` go
func (p *Pools) Exists(name string) bool
```
Exists returns bool if the named Pool exists




### <a name="Pools.Get">func</a> (\*Pools) [Get](https://github.com/cognusion/go-jar/tree/master/pools.go?s=4983:5029#L174)
``` go
func (p *Pools) Get(name string) (*Pool, bool)
```
Get returns a Pool and a bool, given a name, from the Pools




### <a name="Pools.List">func</a> (\*Pools) [List](https://github.com/cognusion/go-jar/tree/master/pools.go?s=4590:4621#L151)
``` go
func (p *Pools) List() []string
```
List returns list of Pool names




### <a name="Pools.Merge">func</a> (\*Pools) [Merge](https://github.com/cognusion/go-jar/tree/master/pools.go?s=5330:5375#L193)
``` go
func (p *Pools) Merge(pools map[string]*Pool)
```
Merge adds-or-replaces the specified pools




### <a name="Pools.Replace">func</a> (\*Pools) [Replace](https://github.com/cognusion/go-jar/tree/master/pools.go?s=5524:5571#L203)
``` go
func (p *Pools) Replace(pools map[string]*Pool)
```
Replace does exactly that on the entire map of Pool




### <a name="Pools.Set">func</a> (\*Pools) [Set](https://github.com/cognusion/go-jar/tree/master/pools.go?s=5183:5227#L185)
``` go
func (p *Pools) Set(name string, pool *Pool)
```
Set adds-or-replaces the named pool




## <a name="ProcessInfo">type</a> [ProcessInfo](https://github.com/cognusion/go-jar/tree/master/healthprocess.go?s=225:397#L14)
``` go
type ProcessInfo struct {
    Ctx context.Context
    // contains filtered or unexported fields
}

```
ProcessInfo is used to track information about ourselves.
All member functions are safe to use across goros







### <a name="NewProcessInfo">func</a> [NewProcessInfo](https://github.com/cognusion/go-jar/tree/master/healthprocess.go?s=551:594#L25)
``` go
func NewProcessInfo(pid int32) *ProcessInfo
```
NewProcessInfo returns an intialized ProcessInfo that has an interval set to 1 minute.
Supply 0 as the pid to autodetect the running process' pid





### <a name="ProcessInfo.CPU">func</a> (\*ProcessInfo) [CPU](https://github.com/cognusion/go-jar/tree/master/healthprocess.go?s=1242:1277#L54)
``` go
func (p *ProcessInfo) CPU() float64
```
CPU returns the current value of the CPU tracker, as a percent of total




### <a name="ProcessInfo.Memory">func</a> (\*ProcessInfo) [Memory](https://github.com/cognusion/go-jar/tree/master/healthprocess.go?s=1102:1140#L49)
``` go
func (p *ProcessInfo) Memory() float64
```
Memory returns the current value of the process memory, as a percent of total




### <a name="ProcessInfo.SetInterval">func</a> (\*ProcessInfo) [SetInterval](https://github.com/cognusion/go-jar/tree/master/healthprocess.go?s=949:999#L44)
``` go
func (p *ProcessInfo) SetInterval(i time.Duration)
```
SetInterval changes(?) the interval at which CPU slices are taken for comparison.




### <a name="ProcessInfo.UpdateCPU">func</a> (\*ProcessInfo) [UpdateCPU](https://github.com/cognusion/go-jar/tree/master/healthprocess.go?s=1462:1495#L60)
``` go
func (p *ProcessInfo) UpdateCPU()
```
UpdateCPU loops while Ctx is valid, sampling our CPU usage every interval.
This should generally only be called once, unless you know what you're doing




## <a name="ProxyResponseModifier">type</a> [ProxyResponseModifier](https://github.com/cognusion/go-jar/tree/master/proxyresponsemodifier.go?s=357:415#L10)
``` go
type ProxyResponseModifier func(resp *http.Response) error
```
ProxyResponseModifier is a type interface compatible with oxy/forward, to allow the proxied response
to be modified at proxy-time, before the Handlers will see the response. This is of special importance
for responses which need absolute mangling before a response is completed e.g. streaming/chunked responses










## <a name="ProxyResponseModifierChain">type</a> [ProxyResponseModifierChain](https://github.com/cognusion/go-jar/tree/master/proxyresponsemodifier.go?s=580:652#L14)
``` go
type ProxyResponseModifierChain struct {
    // contains filtered or unexported fields
}

```
ProxyResponseModifierChain is an encapsulating type to chain multiple ProxyResponseModifier funcs for
sequential execution as a single ProxyResponseModifier










### <a name="ProxyResponseModifierChain.Add">func</a> (\*ProxyResponseModifierChain) [Add](https://github.com/cognusion/go-jar/tree/master/proxyresponsemodifier.go?s=738:805#L19)
``` go
func (p *ProxyResponseModifierChain) Add(prm ProxyResponseModifier)
```
Add appends the provided ProxyResponseModifier to the ProxyResponseModifierChain




### <a name="ProxyResponseModifierChain.ToProxyResponseModifier">func</a> (\*ProxyResponseModifierChain) [ToProxyResponseModifier](https://github.com/cognusion/go-jar/tree/master/proxyresponsemodifier.go?s=1046:1130#L25)
``` go
func (p *ProxyResponseModifierChain) ToProxyResponseModifier() ProxyResponseModifier
```
ToProxyResponseModifier returns a closure ProxyResponseModifier that will sequentially execute each
encapsulated ProxyResponseModifier, discontinuing and returning an error as soon as one is noticed




## <a name="PruneFunc">type</a> [PruneFunc](https://github.com/cognusion/go-jar/tree/master/pools.go?s=10137:10170#L357)
``` go
type PruneFunc func(string) error
```
PruneFunc is a func that may add or remove Pool members










## <a name="RateLimiter">type</a> [RateLimiter](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=8476:8539#L269)
``` go
type RateLimiter struct {
    *limiter.Limiter
    // contains filtered or unexported fields
}

```
RateLimiter is a wrapper around limiter.Limiter







### <a name="NewRateLimiter">func</a> [NewRateLimiter](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=8630:8703#L275)
``` go
func NewRateLimiter(max float64, purgeDuration time.Duration) RateLimiter
```
NewRateLimiter returns a RateLimiter based on the specified max rps and purgeDuration


### <a name="NewRateLimiterCollector">func</a> [NewRateLimiterCollector](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=9158:9240#L296)
``` go
func NewRateLimiterCollector(max float64, purgeDuration time.Duration) RateLimiter
```
NewRateLimiterCollector returns a RateLimiter based on the specified max rps and purgeDuration





### <a name="RateLimiter.Handler">func</a> (\*RateLimiter) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=9372:9434#L304)
``` go
func (rl *RateLimiter) Handler(next http.Handler) http.Handler
```
Handler is the middleware for the RateLimiter




## <a name="Redirect">type</a> [Redirect](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=2839:2905#L100)
``` go
type Redirect struct {
    URL  string
    Code int
    PCRE *pcre.Regexp
}

```
Redirect is a Finisher that returns 301 for the requested Path










### <a name="Redirect.Finisher">func</a> (\*Redirect) [Finisher](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=2967:3035#L107)
``` go
func (rd *Redirect) Finisher(w http.ResponseWriter, r *http.Request)
```
Finisher is a ... Finisher for the instantiated Redirect




## <a name="S3Pool">type</a> [S3Pool](https://github.com/cognusion/go-jar/tree/master/s3pool.go?s=494:554#L24)
``` go
type S3Pool struct {
    // contains filtered or unexported fields
}

```
S3Pool is an http.Handler that grabs a file from S3 and streams it back to the client







### <a name="NewS3Pool">func</a> [NewS3Pool](https://github.com/cognusion/go-jar/tree/master/s3pool.go?s=599:644#L30)
``` go
func NewS3Pool(s3url string) (*S3Pool, error)
```
NewS3Pool returns an S3Pool or an error





### <a name="S3Pool.ServeHTTP">func</a> (\*S3Pool) [ServeHTTP](https://github.com/cognusion/go-jar/tree/master/s3pool.go?s=995:1063#L50)
``` go
func (s3p *S3Pool) ServeHTTP(w http.ResponseWriter, r *http.Request)
```
ServeHTTP is a proper http.Handler for authenticated S3 requests




## <a name="StatusFinisher">type</a> [StatusFinisher](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=2229:2252#L84)
``` go
type StatusFinisher int
```
StatusFinisher is an abstracted type to dynamically provide Finishers of standard HTTP status codes










### <a name="StatusFinisher.Finisher">func</a> (StatusFinisher) [Finisher](https://github.com/cognusion/go-jar/tree/master/finishers.go?s=2321:2394#L87)
``` go
func (sf StatusFinisher) Finisher(w http.ResponseWriter, r *http.Request)
```
Finisher writes a response of the set HTTP status code and text




## <a name="SuiteMap">type</a> [SuiteMap](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=4092:4123#L122)
``` go
type SuiteMap map[string]uint16
```
SuiteMap is a map of TLS cipher suites, to their hex code


``` go
var (
    // Ciphers is a map of ciphers from crypto/tls
    Ciphers SuiteMap

    // SslVersions is a map of SSL/TLS versions, mapped locally
    SslVersions = SuiteMap{
        "VersionSSL30": 0x0300,
        "VersionTLS10": 0x0301,
        "VersionTLS11": 0x0302,
        "VersionTLS12": 0x0303,
        "VersionTLS13": 0x0304,
    }
)
```






### <a name="NewSuiteMapFromCipherSuites">func</a> [NewSuiteMapFromCipherSuites](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=4210:4284#L125)
``` go
func NewSuiteMapFromCipherSuites(cipherSuites []*tls.CipherSuite) SuiteMap
```
NewSuiteMapFromCipherSuites takes a []*CipherSuite and creates a SuiteMap from it





### <a name="SuiteMap.AllSuites">func</a> (\*SuiteMap) [AllSuites](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=5073:5112#L155)
``` go
func (s *SuiteMap) AllSuites() []uint16
```
AllSuites returns the hex codes for all of the cipher suites in an untrustable order




### <a name="SuiteMap.CipherListToSuites">func</a> (\*SuiteMap) [CipherListToSuites](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=5281:5351#L161)
``` go
func (s *SuiteMap) CipherListToSuites(list []string) ([]uint16, error)
```
CipherListToSuites takes an ordered list of cipher suite names, and returns their hex codes in the same order




### <a name="SuiteMap.List">func</a> (\*SuiteMap) [List](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=4853:4887#L144)
``` go
func (s *SuiteMap) List() []string
```
List returns the names of the cipher suites in an untrustable order




### <a name="SuiteMap.Suite">func</a> (\*SuiteMap) [Suite](https://github.com/cognusion/go-jar/tree/master/crypto.go?s=5640:5686#L175)
``` go
func (s *SuiteMap) Suite(number uint16) string
```
Suite reverse lookups a suitename given the number




## <a name="TemplateError">type</a> [TemplateError](https://github.com/cognusion/go-jar/tree/master/errors.go?s=4306:4715#L145)
``` go
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
TemplateError is a static structure to pass into error-wrapping templates










## <a name="Timeout">type</a> [Timeout](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=3559:3623#L119)
``` go
type Timeout struct {
    Duration time.Duration
    Message  string
}

```
Timeout is a middleware that causes a 503 Service Unavailable message to be handed back if the timeout trips










### <a name="Timeout.Handler">func</a> (\*Timeout) [Handler](https://github.com/cognusion/go-jar/tree/master/handlers.go?s=3663:3720#L125)
``` go
func (t *Timeout) Handler(next http.Handler) http.Handler
```
Handler is the handler for Timeout




## <a name="ZulipWork">type</a> [ZulipWork](https://github.com/cognusion/go-jar/tree/master/worker-zulip.go?s=997:1092#L38)
``` go
type ZulipWork struct {
    Client  *zulip.Zulip
    Stream  string
    Topic   string
    Message string
}

```
ZulipWork is a generic Work that can send Zulip notifications










### <a name="ZulipWork.Return">func</a> (\*ZulipWork) [Return](https://github.com/cognusion/go-jar/tree/master/worker-zulip.go?s=1318:1364#L54)
``` go
func (z *ZulipWork) Return(rthing interface{})
```
Return dumps the response. We don't care. :)




### <a name="ZulipWork.Work">func</a> (\*ZulipWork) [Work](https://github.com/cognusion/go-jar/tree/master/worker-zulip.go?s=1123:1161#L46)
``` go
func (z *ZulipWork) Work() interface{}
```
Work is called to do work








- - -
Generated by [godoc2md](http://godoc.org/github.com/cognusion/godoc2md)
