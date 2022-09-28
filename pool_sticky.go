package jar

import (
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/oxy/roundrobin/stickycookie"

	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// NewStickyPool returns a primed RoundRobin that honors pinning based on a cookie value
func NewStickyPool(poolName, cookieName, cookieType string, next http.Handler, opts ...roundrobin.LBOption) (*roundrobin.RoundRobin, error) {
	var (
		cookie         = fmt.Sprintf("%s%s", "jar", poolName)
		sticky         roundrobin.LBOption
		err            error
		lb             *roundrobin.RoundRobin
		cookieHTTPOnly = Conf.GetBool(ConfigStickyCookieHTTPOnly)
		cookieSecure   = Conf.GetBool(ConfigStickyCookieSecure)
	)

	if cookieName != "" {
		cookie = cookieName
	}
	DebugOut.Printf("\t\tSticky with cookie '%s'\n", cookie)

	if cookieType == "" {
		cookieType = "plain"
	}

	switch sct := strings.ToLower(cookieType); sct {
	case "aes":
		// AES-encrypted values
		DebugOut.Printf("\t\tSticky with AES-encryption!\n")
		if sskey := Conf.GetString(ConfigKeysStickyCookie); sskey != "" {
			var (
				ao       stickycookie.CookieValue
				clearKey []byte
			)

			clearKey, err = base64.StdEncoding.DecodeString(sskey)
			if err != nil {
				return nil, err
			}

			if cookielife := Conf.GetDuration(ConfigStickyCookieAESTTL); cookielife > 0 {
				DebugOut.Printf("\t\t... with expiration of %s\n", cookielife.String())
				ao, err = stickycookie.NewAESValue(clearKey, cookielife)
				if err != nil {
					return nil, err
				}
			} else {
				ao, err = stickycookie.NewAESValue(clearKey, time.Duration(0))
				if err != nil {
					return nil, err
				}
			}
			sticky = roundrobin.EnableStickySession(roundrobin.NewStickySessionWithOptions(cookie, roundrobin.CookieOptions{HTTPOnly: cookieHTTPOnly, Secure: cookieSecure}).SetCookieValue(ao))

		} else {
			// No key set!
			return nil, ErrPoolStickyAESNoKey
		}
	case "hash":
		// Hashed values
		DebugOut.Printf("\t\tSticky with hashed values!\n")
		sticky = roundrobin.EnableStickySession(roundrobin.NewStickySessionWithOptions(cookie, roundrobin.CookieOptions{HTTPOnly: cookieHTTPOnly, Secure: cookieSecure}).SetCookieValue(&stickycookie.HashValue{Salt: Conf.GetString(ConfigKeysStickyCookie)}))
	case "plain":
		DebugOut.Printf("\t\tSticky with plaintext values!\n")
		sticky = roundrobin.EnableStickySession(roundrobin.NewStickySessionWithOptions(cookie, roundrobin.CookieOptions{HTTPOnly: cookieHTTPOnly, Secure: cookieSecure}))
	default:
		return nil, fmt.Errorf("invalid Pool.StickyCookieType '%s'", sct)
	}

	if len(opts) > 0 {
		newopts := []roundrobin.LBOption{sticky}
		newopts = append(newopts, opts...)
		lb, err = roundrobin.New(next, newopts...)
	} else {
		lb, err = roundrobin.New(next, sticky)
	}

	if err != nil {
		return nil, err
	}

	return lb, nil
}
