package jar

import (
	"github.com/cognusion/go-timings"

	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Access is a type to provide binary validation of
// addresses, based on the contents of "Allow/Deny" rules.
type Access struct {
	allowAddresses []*net.IP
	allowNetworks  []*net.IPNet
	denyAddresses  []*net.IP
	denyNetworks   []*net.IPNet
	allowAll       bool
	denyAll        bool
}

// NewAccessFromStrings is the safest way to create a safe, valid Access
// type. The supplied "allow" and "deny" strings may be comma-delimited
// lists of IP addresses and/or CIDR networks.
func NewAccessFromStrings(allow, deny string) (*Access, error) {
	var a Access

	if allow != "" {
		for _, addr := range strings.Split(allow, ",") {
			err := a.AddAddress(strings.TrimSpace(addr), true)
			if err != nil {
				return nil, ErrConfigurationError{fmt.Sprintf("Allow rule parsing problem: '%s': %s", allow, err)}
			}
		}
	}

	if deny != "" {
		for _, addr := range strings.Split(deny, ",") {
			err := a.AddAddress(strings.TrimSpace(addr), false)
			if err != nil {
				return nil, ErrConfigurationError{fmt.Sprintf("Deny rule parsing problem: '%s': %s", deny, err)}
			}
		}
	}

	return &a, nil
}

// AddAddress adds the supplied address to either the allow or deny
// lists, depending on the value of the suppled boolean. An error is
// returned if the supplied address cannot be parsed.
func (a *Access) AddAddress(address string, allow bool) error {
	if address == "all" {
		if allow {
			a.allowAll = true
		} else {
			// Deny
			a.denyAll = true
		}
		return nil
	}
	if address == "none" {
		if allow {
			a.allowAll = false
		} else {
			// Deny
			a.denyAll = false
		}
		return nil
	}

	if strings.Contains(address, "/") {
		// Network
		_, addrnet, err := net.ParseCIDR(address)
		if err != nil {
			return err
		}

		if allow {
			a.allowNetworks = append(a.allowNetworks, addrnet)
		} else {
			// Deny
			a.denyNetworks = append(a.denyNetworks, addrnet)
		}
	} else {
		// Address
		addr := net.ParseIP(address)
		if addr == nil {
			return fmt.Errorf("address '%s' doesn't appear to be a valid IP address", address)
		}

		if allow {
			a.allowAddresses = append(a.allowAddresses, &addr)
		} else {
			// Deny
			a.denyAddresses = append(a.denyAddresses, &addr)
		}
	}

	return nil
}

// AccessHandler is a handler that consults r.RemoteAddr and validates
// it against the Access type.
func (a *Access) AccessHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		// Assumes r.RemoteAddr has been vetted by e.g. RealAddr()
		if !a.Validate(r.RemoteAddr) {
			DebugOut.Printf("%s\n", ErrRequestError{r, fmt.Sprintf("Access not allowed for '%s'", r.RemoteAddr)})
			//http.Error(w, ErrRequestError{r, "Access not allowed"}.Error(), http.StatusForbidden)
			RequestErrorResponse(r, w, "Access not allowed", http.StatusForbidden)
			return
		}
		TimingOut.Printf("Access handler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// Validate tests the supplied address against the Access type,
// returning boolean
func (a *Access) Validate(address string) bool {
	defer timings.Track("Access.Validate", time.Now(), TimingOut)
	if address == "" {
		return false
	}

	var addr string
	if strings.Contains(address, ":") {
		if strings.HasSuffix(address, "]") {
			// IPv6 without port. Contrived :(
			addr = strings.TrimPrefix(address, "[")
			addr = strings.TrimSuffix(addr, "]")
		} else {
			var err error
			addr, _, err = net.SplitHostPort(address)
			if err != nil {
				DebugOut.Printf("Error splitting host and port from '%s': %v\n", address, err)
				return false
			}
		}
	} else {
		addr = address
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		DebugOut.Printf("Bad IP address '%s'\n", addr)
		return false
	}

	for _, aa := range a.allowAddresses {
		if aa.Equal(ip) {
			DebugOut.Printf("Explicitly allowed address '%s'\n", addr)
			return true
		}
	}
	for _, an := range a.allowNetworks {
		if an.Contains(ip) {
			DebugOut.Printf("Explicitly allowed network '%s'\n", addr)
			return true
		}
	}

	for _, da := range a.denyAddresses {
		if da.Equal(ip) {
			DebugOut.Printf("Explicitly denied address '%s'\n", addr)
			return false
		}
	}
	for _, dn := range a.denyNetworks {
		if dn.Contains(ip) {
			DebugOut.Printf("Explicitly denied network '%s'\n", addr)
			return false
		}
	}
	// POST: Not explicitly Denied or Allowed
	if a.denyAll && !a.allowAll {
		DebugOut.Printf("Explicit denyAll, and no allow for '%s'\n", addr)
		return false
	}

	// If we have allows, but no denies, assume we want to deny if we get here
	if (len(a.allowAddresses) > 0 || len(a.allowNetworks) > 0) && (len(a.denyAddresses) == 0 && len(a.denyNetworks) == 0) {
		DebugOut.Printf("Implicit denyAll, and no allow for '%s'\n", addr)
		return false
	}

	DebugOut.Printf("Implicit allow for '%s'\n", addr)
	return true
}
