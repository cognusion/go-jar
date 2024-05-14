package jar

import (
	"crypto/tls"

	"golang.org/x/crypto/acme/autocert"
)

// Constants for configuration key strings
const (
	ConfigACMEEnabled        = ConfigKey("acme.enabled")
	ConfigACMEEmailAddress   = ConfigKey("acme.email")
	ConfigACMEHostWhiteList  = ConfigKey("acme.hosts")
	ConfigACMECacheDirectory = ConfigKey("acme.cachedir")
	ConfigACMERenewBefore    = ConfigKey("acme.renewbefore")
)

func boostrapAcme() *tls.Config {

	// TODO: Implement HostPolicy, RenewBefore!

	manager := &autocert.Manager{
		Cache:       autocert.DirCache(Conf.GetString(ConfigACMECacheDirectory)),
		Prompt:      autocert.AcceptTOS,
		Email:       Conf.GetString(ConfigACMEEmailAddress),
		RenewBefore: Conf.GetDuration(ConfigACMERenewBefore),
		HostPolicy:  autocert.HostWhitelist(Conf.GetStringSlice(ConfigACMEHostWhiteList)...),
	}
	return manager.TLSConfig()

}
