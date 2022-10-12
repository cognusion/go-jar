package jar

import (
	"github.com/spf13/viper"

	"fmt"
	"os"
	"strings"
)

// Constants for configuration key strings
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

var (
	// ConfigValidations is used to wire in func()error to be run, validating distributed configs
	ConfigValidations = make(map[string]func() error)
	// ConfigAdditions is used to wire in additional default configurations
	ConfigAdditions = make(configDefaultSetter)
)

// ConfigKey is a string type for static config key name consistency
type ConfigKey = string

// InitConfig creates an config, initialized with defaults and environment-set values, and returns it
func InitConfig() *viper.Viper {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvPrefix("JAR")

	loadDefaults(v)

	return v
}

type configDefaultSetter map[string]interface{}

func (c *configDefaultSetter) Set(v *viper.Viper) {
	for k, val := range *c {
		v.SetDefault(k, val)
	}
}

// LoadConfig read the config file and returns a config object or an error
func LoadConfig(configFilename string, v *viper.Viper) error {

	if configFilename != "" {
		configFilenames := strings.Split(configFilename, ",")
		v.SetConfigFile(configFilenames[0])

		err := v.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigParseError); ok {
				return err
			}
			return fmt.Errorf("unable to locate config file '%s': %w", configFilenames[0], err)
		}
		for _, configFile := range configFilenames[1:] {
			file, err := os.Open(configFile) // For read access.
			if err != nil {
				return fmt.Errorf("unable to open config file '%s': %w", configFile, err)
			}
			defer file.Close()
			if err = v.MergeConfig(file); err != nil {
				return fmt.Errorf("unable to parse/merge Config file '%s': %w", configFile, err)
			}
		}
	}

	return nil
}

// loadDefaults sets all the default values for the config. An error may be returned, to support future operations.
func loadDefaults(v *viper.Viper) error {
	v.SetDefault(ConfigDebug, false)                            // Enable vociferous output
	v.SetDefault(ConfigDebugRequests, false)                    // Enable vociferous output of requests
	v.SetDefault(ConfigDebugResponses, false)                   // Enable vociferous output of responses
	v.SetDefault(ConfigListen, ":8080")                         // ip:port or :port to listen on
	v.SetDefault(ConfigAccessLog, "")                           // Path to file where accesslog, else stdout
	v.SetDefault(ConfigDebugLog, "")                            // Path to file where debug should log to, else stderr
	v.SetDefault(ConfigErrorLog, "")                            // Path to file where errorlog should log to, else stderr
	v.SetDefault(ConfigSlowLog, "")                             // Path to file where the slow requests should log to
	v.SetDefault(ConfigLogSize, 100)                            // Maximum size, in MB, that the currently log can be before rolling
	v.SetDefault(ConfigLogBackups, 3)                           // Maximum number of rolled logs to keep
	v.SetDefault(ConfigLogAge, 28)                              // Maximum age, in days, to keep rolled logs
	v.SetDefault(ConfigPaths, make([]interface{}, 0))           // paths are blocks to define URI prefixes, and the handlers they should use
	v.SetDefault(ConfigPools, make([]interface{}, 0))           // pools are blocks to define loadbalancer pools, their configuration, and their members
	v.SetDefault(ConfigKeys, make(map[string]string))           // keys[keyname] = base64-encodedkey
	v.SetDefault(ConfigAuthPool, make(map[string]interface{}))  // authpool[poolname] = map[opt]value
	v.SetDefault(ConfigHandlers, make([]string, 0))             // handlers is a list of global handlers
	v.SetDefault(ConfigHotUpdate, false)                        // hotupdate controls whether we allow JAR to replace itself
	v.SetDefault(ConfigHotConfig, false)                        // hotconfig controls whether we watch for config changes, and reload on change, or not
	v.SetDefault(ConfigTempFolder, "/tmp")                      // tempfolder is where temporary files go
	v.SetDefault(ConfigUpdatePath, "")                          // update path is a full S3 URL to where an update file is
	v.SetDefault(ConfigStripRequestHeaders, make([]string, 0))  // striprequestheaders lists request headers that should be removed
	v.SetDefault(ConfigTimeout, 0)                              // Sets the total request/response timeout. Can be overridden per-path
	v.SetDefault(ConfigKeepaliveTimeout, "5s")                  // Sets the amount of time keptalive sockets linger
	v.SetDefault(ConfigMaxConnections, 0)                       // Maximum number of simultaneous connections to the listener
	v.SetDefault(ConfigRequestIDHeaderName, "X-Request-ID")     // Name of the header to use for adding the requestid
	v.SetDefault(ConfigSlowRequestMax, "")                      // Sets the default slow request time for logging
	v.SetDefault(ConfigTrustRequestIDHeader, false)             // Enable trusting of incoming ConfigRequestIDHeaderName to set the requestid
	v.SetDefault(ConfigEC2, false)                              // Enable AWS EC2-specific features
	v.SetDefault(ConfigAuthoritativeDomains, make([]string, 0)) // If set, restricts request handling to the domains listed, otherwise 400s
	v.SetDefault(ConfigVersionRequired, "")                     // Sets the minimum ``go-jar`` VERSION that the config is valid for.

	v.SetDefault(ConfigURLRouteHeaders, false)                       // Enables setting the headers below, if SwitchHandler is used
	v.SetDefault(ConfigURLRouteIDHeaderName, "X-URLROUTEID")         // Name of the header to use for route ID
	v.SetDefault(ConfigURLRouteEndpointHeaderName, "X-URLROUTEDEST") // Name of the header to use for route endpoint
	v.SetDefault(ConfigURLRouteNameHeaderName, "X-URLROUTENAME")     // Name of the header to use for route name
	v.SetDefault(ConfigPoolHeaderName, "X-URLPOOL")                  // Name of the header to use when capturing the servicing pool
	v.SetDefault(ConfigPoolMemberHeaderName, "X-URLPOOLMEMBER")      // Name of the header to use when capturing the servicing request pool member
	ConfigAdditions.Set(v)

	return nil
}

// ValidateExtras runs through a list of referenced functions, and returns any errors they return.
// All functions will be run, so an array of errors may be returned
func ValidateExtras() []error {
	errs := make([]error, 0)

	for _, f := range ConfigValidations {
		err := f()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
