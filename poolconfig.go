package jar

import (
	"strings"

	"github.com/spf13/cast"
	"github.com/vulcand/oxy/roundrobin"

	"net/http"
	"net/url"
)

// PoolConfig is type exposing expected configuration for a pool, abstracted for passing around
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
	// StickyCookieType allows for the setting of cookie values to "plain", "hash", or "aes"-encrypted.
	// The value of Conf.GetString(ConfigKeysStickyCookie) will be the salt for "hash" as-is, or the
	// base64-encoded key for "aes".
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

// PoolOptions is an MSI with a case-agnostic getter
type PoolOptions map[string]interface{}

// Get returns an interface{} if *key* matches, otherwise nil
func (p *PoolOptions) Get(key string) interface{} {
	if p == nil {
		return nil
	}
	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return v
		}
	}
	return nil
}

// GetString returns a string if *key* matches, otherwise empty string
func (p *PoolOptions) GetString(key string) string {
	if p == nil {
		return ""
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToString(v)
		}
	}
	return ""
}

// GetInt returns an int if *key* matches, otherwise -1
func (p *PoolOptions) GetInt(key string) int {
	if p == nil {
		return -1
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToInt(v)
		}
	}
	return -1
}

// GetFloat64 returns a float64 if *key* matches, otherwise -1
func (p *PoolOptions) GetFloat64(key string) float64 {
	if p == nil {
		return -1
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToFloat64(v)
		}
	}
	return -1
}

// GetBool returns a bool value if *key* matches, otherwise false
func (p *PoolOptions) GetBool(key string) bool {
	if p == nil {
		return false
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToBool(v)
		}
	}
	return false
}

// GetStringSlice returns a []string if *key* matches, otherwise an empty []string
func (p *PoolOptions) GetStringSlice(key string) []string {
	if p == nil {
		return make([]string, 0)
	}

	lckey := strings.ToLower(key)
	for k, v := range *p {
		if lckey == strings.ToLower(k) {
			return cast.ToStringSlice(v)
		}
	}
	return make([]string, 0)
}

// PoolManager is an interface to encompass oxy/roundrobin and our chpool
type PoolManager interface {
	Servers() []*url.URL
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	ServerWeight(u *url.URL) (int, bool)
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
	NextServer() (*url.URL, error)
	Next() http.Handler
}
