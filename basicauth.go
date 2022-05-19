package jar

import (
	"github.com/cognusion/go-timings"
	"github.com/fsnotify/fsnotify"

	"crypto/rand"
	"crypto/subtle"
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// ErrSourceVerificationFailed is an error returned when an authentication source cannot be verified
	ErrSourceVerificationFailed = Error("cannot verify the provided source")

	// ErrSourceNotSupported is an error returned when the authentication source scheme is not supported [yet]
	ErrSourceNotSupported = Error("specified source scheme is not supported")

	/* Commented out for now, as we're not supporting LDAP yet
	ErrLDAPURLPrefixInvalid   = Error("URL prefix must be ldap:// or ldaps://")
	ErrLDAPURLInvalid         = Error("URL must contain at least host/query")
	ErrLDAPURLHostPortInvalid = Error("hostname contains a colon (:), must contain at least host:port")
	ErrLDAPURLPortInvalid     = Error("port must be between 1 and 65535")
	*/
)

var (
	randMax = big.NewInt(1000)
)

// BasicAuth wraps a handler requiring HTTP basic auth
type BasicAuth struct {
	// List of allowed users
	Users  []string
	users  map[string]string
	realm  string
	source string
	mu     sync.RWMutex
}

// NewBasicAuth takes a source, realm, and list of users, returning an initialized *BasicAuth
func NewBasicAuth(source, realm string, users []string) *BasicAuth {

	b := BasicAuth{
		realm:  realm,
		source: source,
		Users:  users,
	}

	if strings.HasPrefix(source, "file://") {
		// Reload the fie on change
		FileWatcher.Add(source, func(e fsnotify.Event) { b.Load() })
		go b.Load()
	} else {
		// No file, no BasicAuth
		return nil
	}

	return &b
}

// NewVerifiedBasicAuth takes a source, realm, and list of users, verifies the auth source, and returns an initialized *BasicAuth or an error
func NewVerifiedBasicAuth(source, realm string, users []string) (*BasicAuth, error) {

	b := BasicAuth{
		realm:  realm,
		source: source,
		Users:  users,
	}

	if err := b.VerifySource(); err != nil {
		return nil, err
	}

	if strings.HasPrefix(source, "file://") {
		// Reload the fie on change
		FileWatcher.Add(source, func(e fsnotify.Event) { b.Load() })
		if err := b.Load(); err != nil {
			FileWatcher.Delete(source)
			return nil, err
		}
	} else {
		return nil, ErrSourceNotSupported
	}

	return &b, nil
}

// VerifySource checks that the requested authentication source is valid, and accessible
func (b *BasicAuth) VerifySource() error {
	if strings.HasPrefix(b.source, "file://") {
		// file
		file := strings.TrimPrefix(b.source, "file://")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return ErrSourceVerificationFailed
		}
	} else {
		return ErrSourceNotSupported
	}

	return nil
}

// Authenticate takes a username, password, realm, and return bool if the authentication is positive
func (b *BasicAuth) Authenticate(username, password, realm string) bool {
	return b.fileAuthenticate(username, password, realm)
}

func (b *BasicAuth) fileAuthenticate(username, password, realm string) bool {
	for _, u := range b.Users {
		if subtle.ConstantTimeCompare([]byte(u), []byte(username)) == 1 {
			// check auth
			b.mu.RLock()
			hpassword, exists := b.users[username]
			b.mu.RUnlock()
			if !exists {
				ErrorOut.Printf("User %s doesn't exist in hash\n", username)
				// Random sleep here, to foil timing attacks
				rnd, err := rand.Int(rand.Reader, randMax)
				if err != nil {
					ErrorOut.Printf("Error generating random number for sleep: %s\n", err)
					time.Sleep(2 * time.Millisecond)
				} else {
					stime := time.Duration(rnd.Int64()) * time.Microsecond
					DebugOut.Printf("Bad auth, sleeping for %s\n", stime.String())
					time.Sleep(stime)
				}

				return false
			}

			// Check MD5 first
			if err := compareMD5HashAndPassword([]byte(hpassword), []byte(password)); err != nil {
				// Failed, try SHA1
				if err := compareShaHashAndPassword([]byte(hpassword), []byte(password)); err != nil {
					// Failed again, fail
					return false
				}
				// Win @ SHA1
				return true
			}
			// Win @ MD5
			return true
		}
	}
	return false
}

func (b *BasicAuth) handler(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()

		ip := ipOnly(r.RemoteAddr)
		if ip == "" {
			// Orly?
			ErrorOut.Print(ErrRequestError{r, fmt.Sprintf("Request from '%s' apparently has no RemoteAddr!\n", ip)})
		}

		user, pass, ok := r.BasicAuth()

		if !ok || !b.Authenticate(user, pass, b.realm) {
			DebugOut.Printf("Auth for %s on %s failed.\n", user, b.realm)
			ErrorOut.Printf("%s\n", ErrRequestError{r, fmt.Sprintf("Basic Authentication for '%s' failed on realm '%s'", user, b.realm)})

			// Random sleep here, to foil timing attacks
			rnd, err := rand.Int(rand.Reader, randMax)
			if err != nil {
				ErrorOut.Printf("Error generating random number for sleep: %s\n", err)
				time.Sleep(2 * time.Millisecond)
			} else {
				stime := time.Duration(rnd.Int64()) * time.Microsecond
				DebugOut.Printf("Bad auth, sleeping for %s\n", stime.String())
				time.Sleep(stime)
			}

			// Always set this on bad auth
			w.Header().Set("WWW-Authenticate", `Basic realm="`+b.realm+`"`)

			//http.Error(w, ErrRequestError{r, http.StatusText(http.StatusUnauthorized)}.Error(), http.StatusUnauthorized)
			RequestErrorResponse(r, w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			TimingOut.Printf("BasicAuth.handler took (padded up to 2ms) %s\n", t.Since().String())
			return
		}

		TimingOut.Printf("BasicAuth.handler took %s\n", t.Since().String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// Load prepares any pre-auth dancing, caching, etc. necessary
func (b *BasicAuth) Load() error {
	file := strings.TrimPrefix(b.source, "file://")

	f, err := os.Open(file)
	if err != nil {
		ErrorOut.Printf("Error opening '%s': %s\n", file, err)
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ':'
	r.Comment = '#'
	r.TrimLeadingSpace = true

	lmap := make(map[string]string)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			ErrorOut.Printf("Error reading '%s': %s\n", file, err)
			return err
		}
		lmap[record[0]] = record[1]
	}

	b.mu.Lock()
	b.users = lmap
	b.mu.Unlock()

	return nil
}

/* Commenting out this LDAP-specific code for now. May revisit.
type LdapSource struct {
	Server string
	Port   int
	Tls    bool
	Query  string
}

func NewLdapSourceFromUrl(url string) (LdapSource, error) {
	l := LdapSource{}
	l.Port = 389 // default

	// Prefix evaluation
	if strings.HasPrefix(url, "ldap://") {
		url = strings.TrimPrefix(url, "ldap://")
		l.Tls = false
	} else if strings.HasPrefix(url, "ldaps://") {
		url = strings.TrimPrefix(url, "ldaps://")
		l.Tls = true
	} else {
		return LdapSource{}, ErrLDAPURLPrefixInvalid
	}

	// Hostname[:port] evaluation
	urlparts := strings.SplitN(url, "/", 2)
	if len(urlparts) < 2 {
		return LdapSource{}, ErrLDAPURLInvalid
	} else if strings.Contains(urlparts[0], ":") {
		// Port specified
		nameparts := strings.SplitN(urlparts[0], ":", 2)
		if len(nameparts) < 2 || nameparts[0] == "" || nameparts[1] == "" {
			// name: or :port or other nonsense
			return LdapSource{}, ErrLDAPURLHostPortInvalid
		}
		port, err := strconv.ParseInt(nameparts[1], 10, 0) // parse the port into an int
		if err != nil {
			return LdapSource{}, err
		}
		if port < 1 || port > 65535 {
			return LdapSource{}, ErrLDAPURLPortInvalid
		}
		l.Port = int(port)
		l.Server = nameparts[0]
	} else {
		l.Server = urlparts[0]
	}
	l.Query = urlparts[1]

	return l, nil
}
*/
