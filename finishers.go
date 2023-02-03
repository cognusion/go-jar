package jar

import (
	"github.com/gorilla/mux"

	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	// ErrFinisher404 returned by HandleFinisher if the requested finisher doesn't exist. Other errors should be treated as failures
	ErrFinisher404 = Error("requested finisher does not exist")
)

var (
	// Finishers is a map of available HandlerFuncs
	Finishers = make(FinisherMap)

	// FinisherSetups is a map of Finishers that need exec-time setup checks
	FinisherSetups = make(map[string]FinisherSetupFunc)
)

func init() {
	// Set up the static finishers
	Finishers["forbidden"] = Forbidden
	Finishers["poolmemberadder"] = PoolMemberAdder
	Finishers["poolmemberloser"] = PoolMemberLoser
	Finishers["poolmemberlister"] = PoolMemberLister
	Finishers["poollister"] = PoolLister
}

// FinisherMap maps Finisher names to their HandlerFuncs
type FinisherMap map[string]http.HandlerFunc

// List returns the names of all of the Finishers
func (h *FinisherMap) List() []string {
	l := make([]string, len(*h))
	i := 0
	for k := range *h {
		l[i] = k
		i++
	}
	return l
}

// FinisherSetupFunc is declared for Finishers that need exec-time setup checks
type FinisherSetupFunc func(*Path) (http.HandlerFunc, error)

// HandleFinisher takes a Finisher HandlerFunc name, and returns the function for it and nil, or nil and and error
func HandleFinisher(handler string, path *Path) (http.HandlerFunc, error) {
	lcHandler := strings.ToLower(handler)
	var (
		h  http.HandlerFunc
		ok bool
	)

	// Handle dynamic "httpstatusNNN" finishers
	if strings.HasPrefix(lcHandler, "httpstatus") {
		strNum := strings.TrimPrefix(lcHandler, "httpstatus")
		n, err := strconv.Atoi(strNum)
		if err != nil {
			// Ahhhh
			return nil, ErrConfigurationError{fmt.Sprintf("Finisher '%s' selected, but '%s' is maybe non-numeric: %s", handler, strNum, err.Error())}
		} else if n < 100 || n >= 600 {
			// Not amused.
			return nil, ErrConfigurationError{fmt.Sprintf("Finisher '%s' selected, but '%d' too small/large to be an HTTP code", handler, n)}
		}
		return StatusFinisher(n).Finisher, nil
	}

	if h, ok = Finishers[lcHandler]; !ok {
		// Finisher cannot be found
		return nil, ErrFinisher404
	}

	if fs, ok := FinisherSetups[lcHandler]; ok {
		// Finisher has a setup component

		var (
			fsh   http.HandlerFunc
			fsErr error
		)
		if fsh, fsErr = fs(path); fsErr != nil {
			// Error!
			return nil, fsErr
		} else if fsh != nil {
			// A FinisherSetup *may* return a HandlerFunc
			h = fsh
		}
	}

	return h, nil
}

// StatusFinisher is an abstracted type to dynamically provide Finishers of standard HTTP status codes
type StatusFinisher int

// Finisher writes a response of the set HTTP status code and text
func (sf StatusFinisher) Finisher(w http.ResponseWriter, r *http.Request) {
	isf := int(sf)
	w.WriteHeader(isf)
	w.Write([]byte(http.StatusText(isf)))
}

// Forbidden is a Finisher that returns 403 for the requested Path
func Forbidden(w http.ResponseWriter, r *http.Request) {
	//http.Error(w, ErrRequestError{r, ErrForbiddenError.Error()}.Error(), http.StatusForbidden)
	RequestErrorResponse(r, w, ErrForbiddenError.Error(), http.StatusForbidden)
}

// Redirect is a Finisher that returns 301 for the requested Path
type Redirect struct {
	URL    string
	Code   int
	Regexp *regexp.Regexp
}

// Finisher is a ... Finisher for the instantiated Redirect
func (rd *Redirect) Finisher(w http.ResponseWriter, r *http.Request) {
	u := rd.URL
	if rd.Code < 300 || rd.Code >= 400 {
		rd.Code = http.StatusMovedPermanently
	}

	if rd.Regexp != nil && rd.Regexp.NumSubexp() > 0 {
		// there are submatch groups to care about
		m := rd.Regexp.FindStringSubmatch(r.Host)
		if m == nil || len(m) < rd.Regexp.NumSubexp()+1 {
			RequestErrorResponse(r, w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		for i := 0; i < rd.Regexp.NumSubexp(); i++ {
			if g := m[i+1]; g != "" {
				s := fmt.Sprintf("$%d", i+1) // $1 $2 $3 etc
				u = strings.Replace(u, s, g, 1)
			}
		}
	}
	u = strings.Replace(u, "%1", r.URL.RequestURI(), -1)
	http.Redirect(w, r, u, rd.Code)
}

// GenericResponse is a Finisher that returns a possibly-wrapped response
type GenericResponse struct {
	Message string
	Code    int
}

// Finisher is a ... Finisher for the instantiated GenericResponse
func (gr *GenericResponse) Finisher(w http.ResponseWriter, r *http.Request) {
	if gr.Code == 0 {
		gr.Code = 200
	}
	RequestErrorResponse(r, w, gr.Message, gr.Code)
}

// PoolLister is a finisher to list the pools
func PoolLister(w http.ResponseWriter, r *http.Request) {

	for _, pool := range LoadBalancers.List() {
		w.Write([]byte(pool + "\n"))
	}

	w.Write([]byte("\n"))
}

// PoolMemberLister is a finisher to list the members of an existing pool
func PoolMemberLister(w http.ResponseWriter, r *http.Request) {

	var (
		poolName string
	)

	mvars := mux.Vars(r)
	if v, ok := mvars["poolname"]; ok {
		poolName = v
	} else {
		// Not OK
		http.Error(w, "Value of 'poolname' not found", http.StatusBadRequest)
		return
	}

	if pool, ok := LoadBalancers.Get(poolName); ok {
		if !pool.IsMaterialized() {
			_, err := pool.GetPool() // Materializes if not already
			if err != nil {
				ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("%s failed to materialize pool %s: %s", "PoolMemberLister", poolName, err.Error())})
				http.Error(w, "Pool management error", http.StatusInternalServerError)
				return
			}
		}

		members := pool.ListMembers()
		for _, m := range members {
			w.Write([]byte(m.String() + "\n"))
		}
	} else {
		http.Error(w, "Pool not found", http.StatusNotFound)
		return
	}

	w.Write([]byte("\n"))
}

// PoolMemberAdder is a finisher to add a member to an existing pool
func PoolMemberAdder(w http.ResponseWriter, r *http.Request) {

	var (
		poolName  string
		memberURL string
	)

	mvars := mux.Vars(r)
	if v, ok := mvars["poolname"]; ok {
		poolName = v
	} else {
		// Not OK
		http.Error(w, "Value of 'poolname' not found", http.StatusBadRequest)
		return
	}

	if v, ok := mvars["b64memberurl"]; ok {
		memberURL = v
	} else {
		// Not OK
		http.Error(w, "Value of 'member' not found", http.StatusBadRequest)
		return
	}

	if pool, ok := LoadBalancers.Get(poolName); ok {
		if !pool.IsMaterialized() {
			_, err := pool.GetPool() // Materializes if not already
			if err != nil {
				ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("%s failed to materialize pool %s: %s", "PoolMemberAdder", poolName, err.Error())})
				http.Error(w, ErrRequestError{r, "Pool management error"}.String(), http.StatusInternalServerError)
				return
			}
		}

		mu, derr := base64.StdEncoding.DecodeString(memberURL)
		if derr != nil {
			http.Error(w, ErrRequestError{r, fmt.Sprintf("Error decoding memberURL: %s", derr.Error())}.String(), http.StatusBadRequest)
			return
		}
		mus := strings.TrimSpace(string(mu))

		err := pool.AddMember(mus)
		if err != nil {
			http.Error(w, ErrRequestError{r, fmt.Sprintf("Error adding member: %s", err.Error())}.String(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Pool not found", http.StatusNotFound)
		return
	}

	w.Write([]byte("Added"))
}

// PoolMemberLoser is a finisher to remove a member from an existing pool
func PoolMemberLoser(w http.ResponseWriter, r *http.Request) {

	var (
		poolName  string
		memberURL string
	)

	mvars := mux.Vars(r)
	if v, ok := mvars["poolname"]; ok {
		poolName = v
	} else {
		// Not OK
		http.Error(w, "Value of 'poolname' not found", http.StatusBadRequest)
		return
	}

	if v, ok := mvars["b64memberurl"]; ok {
		memberURL = v
	} else {
		// Not OK
		http.Error(w, "Value of 'member' not found", http.StatusBadRequest)
		return
	}

	if pool, ok := LoadBalancers.Get(poolName); ok {
		if !pool.IsMaterialized() {
			_, err := pool.GetPool() // Materializes if not already
			if err != nil {
				ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("%s failed to materialize pool %s: %s", "PoolMemberLoser", poolName, err.Error())})
				http.Error(w, "Pool management error", http.StatusInternalServerError)
				return
			}
		}

		mu, derr := base64.StdEncoding.DecodeString(memberURL)
		if derr != nil {
			http.Error(w, ErrRequestError{r, fmt.Sprintf("Error decoding memberURL: %s", derr.Error())}.String(), http.StatusBadRequest)
			return
		}
		mus := strings.TrimSpace(string(mu))

		err := pool.DeleteMember(mus)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error removing member: %s", err.Error()), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Pool not found", http.StatusNotFound)
		return
	}

	w.Write([]byte("Removed"))
}
