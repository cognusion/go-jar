package jar

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
	// Stickyness
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
}
