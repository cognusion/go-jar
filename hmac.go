package jar

import (
	"github.com/cognusion/go-timings"

	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PathOptions for HMAC signing
const (
	ConfigHMACKey            = ConfigKey("hmac.key")
	ConfigHMACSalt           = ConfigKey("hmac.salt")
	ConfigHMACExpiration     = ConfigKey("hmac.expiration")
	ConfigHMACExpirationName = ConfigKey("hmac.expirationfield")
)

func init() {
	// Set up the static finishers
	Finishers["hmacsigner"] = nil
	FinisherSetups["hmacsigner"] = func(p *Path) (http.HandlerFunc, error) {
		key := p.Options.GetString(ConfigHMACKey)
		salt := p.Options.GetString(ConfigHMACSalt)
		exp, durErr := p.Options.GetDuration(ConfigHMACExpiration)
		expname := p.Options.GetString(ConfigHMACExpirationName)

		if key == "" {
			return nil, ErrConfigurationError{"HMAC Signer requires at least a key"}
		} else if durErr != nil {
			return nil, fmt.Errorf("HMAC Signer Expiration must be a valid duration string: %w", durErr)
		}

		verif := NewHMAC(key, salt, exp)
		if expname != "" {
			verif.ExpirationField = expname
		}

		return http.StripPrefix(p.Path, verif).ServeHTTP, nil
	}
}

// HMAC is a Handler that verifies the signature and possibly the timestamp of a request URL,
// and a Finisher that can sign URLs if so desired.
type HMAC struct {
	key  []byte
	salt []byte
	// If non-zero, the Handler will enforce timestamp (UnixMilli) comparison to "now"
	Expiration time.Duration
	// If set, this will be the name of the query parameter holding the timestamp.
	// If unset, "expiration" is assumed.
	ExpirationField string
}

// NewHMAC returns an initialized Verifier. If expiration is unset, expirations are not enforced.
func NewHMAC(key, salt string, expiration time.Duration) *HMAC {
	return &HMAC{
		key:             []byte(key),
		salt:            []byte(salt),
		Expiration:      expiration,
		ExpirationField: "expiration",
	}
}

// ServeHTTP is a Finisher to handle the request. It assumes that any preceding URI cruft
// has been stripped
func (h *HMAC) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer timings.Track("HMAC.Signer", time.Now(), TimingOut)

	var exp time.Time
	if h.Expiration > 0 {
		exp = time.Now().Add(h.Expiration)
		req.URL.RawQuery += fmt.Sprintf("&%s=%d", h.ExpirationField, exp.UnixMilli())
	}
	uri := craftURI(req.URL.Path, req.URL.Query())
	hash := signHMAC([]byte(uri), h.key, h.salt)
	signedURI := craftURI(fmt.Sprintf("%s/%s", req.URL.Path, hash), req.URL.Query())

	DebugOut.Printf("HMAC Signed %s as %s\n", uri, signedURI)
	http.Redirect(w, req, signedURI, http.StatusTemporaryRedirect) // Maybe we should print as text instead of Redirect?
}

// Handler does the HMAC verification, and possibly expiration calculation, of the request
func (h *HMAC) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var (
			hashPart string
			realURI  string
			parts    []string
		)
		// Timings
		t := timings.Tracker{}
		t.Start()

		if parts = strings.Split(r.URL.Path, "/"); len(parts) < 2 {
			DebugOut.Printf(ErrRequestError{r, "Request path doesn't have at least 2 /-delimited parts. Unsigned or malformed"}.Error())
			http.Error(w, ErrRequestError{r, http.StatusText(http.StatusForbidden)}.Error(), http.StatusForbidden) // Machine-readable
			return
		}
		DebugOut.Printf("URI: %s Parts: %+v\n", r.URL.Path, parts)
		hashPart = parts[len(parts)-1] // hash
		r.URL.Path = strings.Join(parts[:len(parts)-1], "/")
		realURI = craftURI(r.URL.Path, r.URL.Query()) // set URI without the hmac

		DebugOut.Printf(ErrRequestError{r, fmt.Sprintf("Verifying '%s' against %s", realURI, hashPart)}.Error())
		ok, err := verifyHMAC([]byte(realURI), h.key, h.salt, hashPart)
		if err != nil {
			DebugOut.Printf(ErrRequestError{r, fmt.Sprintf("Error while verifying HMAC: %s", err.Error())}.Error())
			http.Error(w, ErrRequestError{r, http.StatusText(http.StatusForbidden)}.Error(), http.StatusForbidden) // Machine-readable
			return
		} else if !ok {
			DebugOut.Printf(ErrRequestError{r, "HMAC verification failed"}.Error())
			http.Error(w, ErrRequestError{r, http.StatusText(http.StatusForbidden)}.Error(), http.StatusForbidden) // Machine-readable
			return
		}

		if h.Expiration > 0 {
			var expirationField = "expiration"
			if h.ExpirationField != "" {
				expirationField = h.ExpirationField
			}

			expStr := r.URL.Query().Get(expirationField)
			exp, perr := strconv.ParseInt(expStr, 10, 0)
			if perr != nil {
				DebugOut.Printf(ErrRequestError{r, fmt.Sprintf("Error while converting expiration number to int: %s", perr.Error())}.Error())
				http.Error(w, ErrRequestError{r, http.StatusText(http.StatusForbidden)}.Error(), http.StatusForbidden) // Machine-readable
				return
			}

			expTime := time.UnixMilli(exp)
			if !time.Now().Before(expTime) {
				DebugOut.Printf(ErrRequestError{r, "HMAC expired"}.Error())
				http.Error(w, ErrRequestError{r, http.StatusText(http.StatusForbidden)}.Error(), http.StatusForbidden) // Machine-readable
				return
			}
		}

		TimingOut.Printf("HMACHandler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func craftURI(uri string, q url.Values) string {
	if len(q) > 0 {
		return fmt.Sprintf("%s?%s", uri, q.Encode())
	}
	return uri
}
