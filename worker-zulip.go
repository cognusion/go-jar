package jar

import (
	"time"

	zulip "github.com/cognusion/go-zulipsend"
)

const (
	// ErrNoZulipClient is returned by a worker when there is no Zulip client defined
	ErrNoZulipClient = Error("no Zulip client defined")
)

// Constants for configuration key strings
const (
	ConfigZulipBaseURL       = ConfigKey("zulip.url")
	ConfigZulipUsername      = ConfigKey("zulip.username")
	ConfigZulipToken         = ConfigKey("zulip.token")
	ConfigZulipRetryCount    = ConfigKey("zulip.retrycount")
	ConfigZulipRetryInterval = ConfigKey("zulip.retryinterval")
)

// ZulipClient is a global Zulip client to use for messaging, or nil if not
var ZulipClient *zulip.Zulip

// newZulipClient returns a zulip.Zulip
func newZulipClient(url, username, token string, retries int, interval time.Duration) *zulip.Zulip {
	return &zulip.Zulip{
		BaseURL:  url,
		Username: username,
		Token:    token,
		Retries:  retries,
		Interval: interval,
	}
}

//ZulipWork is a generic Work that can send Zulip notifications
type ZulipWork struct {
	Client  *zulip.Zulip
	Stream  string
	Topic   string
	Message string
}

// Work is called to do work
func (z *ZulipWork) Work() interface{} {
	if z.Client == nil {
		return ErrNoZulipClient
	}
	return z.Client.Send(z.Stream, z.Topic, z.Message)
}

// Return dumps the response. We don't care. :)
func (z *ZulipWork) Return(rthing interface{}) {
	// Dump it
	if rthing != nil {
		ErrorOut.Printf("ZulipWork to %s/%s \"%s\" returned error: %v\n", z.Stream, z.Topic, z.Message, rthing)
	}
}
