package jar

import (
	"github.com/cognusion/go-timings"
	"github.com/fsnotify/fsnotify"

	"github.com/cognusion/go-jar/mapmap"

	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	switchIDKey urlSwitchID = iota
	switchEndpointKey
	switchNameKey
)

// Constants for configuration key strings
const (
	ConfigMapFiles                 = ConfigKey("mapfiles")
	ConfigMapIDMap                 = ConfigKey("urlroute.idmap")
	ConfigMapEndpointMap           = ConfigKey("urlroute.endpointmap")
	ConfigSwitchHandlerEnforce     = ConfigKey("SwitchHandler.enforce")
	ConfigSwitchHandlerStripPrefix = ConfigKey("SwitchHandler.stripprefix") // e.g. xzy strips ^xyz.*-
)

var (
	// SwitchMaps are maps of URLs parts and their IDs and/or endpoints
	SwitchMaps = mapmap.NewMapMap()
)

type urlSwitchID int

func init() {

	ConfigAdditions[ConfigMapFiles] = make(map[string]string) // mapfiles[name]=filepath
	ConfigAdditions[ConfigMapIDMap] = "ids"
	ConfigAdditions[ConfigMapEndpointMap] = "endpoints"

	InitFuncs.Add(func() {
		if mapfiles := Conf.GetStringMapString(ConfigMapFiles); len(mapfiles) > 0 {

			if SwitchMaps.Size() == 0 {
				// Since InitFuncs may be called multiple times, we don't want to orphan these
				SwitchMaps = mapmap.NewMapMapWithMapNames(Conf.GetString(ConfigMapIDMap), Conf.GetString(ConfigMapEndpointMap))

				// Wrapper around HandleReload for Watcher
				hr := func(e fsnotify.Event) {
					HandleReload(e.Name, mapfiles)
				}

				// Spawn the watcher & load the maps
				for n, f := range mapfiles {
					err := SwitchMaps.Load(n, f)
					if err != nil {
						ErrorOut.Printf("Error loading map file '%s': %s\n", f, err)
						continue
					}

					err = FileWatcher.Add(f, hr)
					if err != nil {
						ErrorOut.Println(err)
						continue
					}
				}
			}
		}
	})

	// Set up the static handlers
	Handlers["switchhandler"] = SwitchHandler

	// Set up the static finishers
	Finishers["urlswitch"] = EndpointDecider
}

// SwitchHandler adds URL switching information to the request context
func SwitchHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Timings
		t := timings.Tracker{}
		t.Start()
		defer TimingOut.Printf("SwitchHandler took %s\n", t.Since().String())

		// Clean up unique-prefix (browser connection limit -buster) domains
		if sp := Conf.GetString(ConfigSwitchHandlerStripPrefix); sp != "" && strings.HasPrefix(r.Host, sp) && strings.Contains(r.Host, "-") {
			r.Host = strings.TrimPrefix(strings.TrimLeftFunc(r.Host, func(r rune) bool {
				return !unicode.Is(unicode.Hyphen, r)
			}), "-")
		}

		urlname := ""
		hostparts := strings.Split(r.Host, ".")
		if len(hostparts) > 1 {
			urlname = hostparts[0]
			if _, err := strconv.Atoi(urlname); err == nil {
				// Probably an IP address
				ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Name %s looks like a number. Host: '%s' Request: '%s'.", urlname, r.Host, r.URL.String())})
				RequestErrorResponse(r, w, "Invalid host header: IP address?", http.StatusBadRequest)
				return
			}
		} else {
			// no dots?
			urlname = r.Host
		}

		if c := Conf.GetStringMapString(ConfigMapFiles); len(c) > 0 {
			// We have some mapfiles
			o := SwitchMaps.GetURLRoute(urlname)
			if o.ID == "" && Conf.GetBool(ConfigSwitchHandlerEnforce) {
				ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Key %s not found, and enforcement requested.", urlname)})
				RequestErrorResponse(r, w, "Invalid host request", http.StatusBadRequest)
				return
			}

			//DebugOut.Printf("SwitchHandler endpoint: %s urlname: %s id: %s\n", o.Endpoint, o.Name, o.ID)
			r = r.WithContext(context.WithValue(context.WithValue(context.WithValue(r.Context(), switchEndpointKey, o.Endpoint), switchNameKey, o.Name), switchIDKey, o.ID))

			// Set the switch headers:
			if Conf.GetBool(ConfigURLRouteHeaders) {
				r.Header.Set(Conf.GetString(ConfigURLRouteIDHeaderName), o.ID)
				r.Header.Set(Conf.GetString(ConfigURLRouteEndpointHeaderName), o.Endpoint)
				r.Header.Set(Conf.GetString(ConfigURLRouteNameHeaderName), o.Name)
			}
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// EndpointDecider is a Finisher that inspects the ``switchEndpointKey`` context to determine which materialized
// Pool should get the request.
// Requests for clusters that are not materialized, or not having the ``clustername`` context value set
// will result in unrecoverable errors
func EndpointDecider(w http.ResponseWriter, r *http.Request) {
	// Timings
	t := timings.Tracker{}
	t.Start()

	var (
		requestID   string
		endpoint    string
		pathOptions PathOptions
	)

	if popts := r.Context().Value(PathOptionsKey); popts != nil {
		pathOptions = popts.(PathOptions)
	}
	if rid := r.Context().Value(requestIDKey); rid != nil {
		requestID = rid.(string)
	}
	if cid := r.Context().Value(switchEndpointKey); cid != nil {
		endpoint = fmt.Sprintf("%s%s%s", pathOptions.GetString("EndpointPrefix"), cid.(string), pathOptions.GetString("EndpointSuffix"))
	}

	// If cluster isn't set, we're DOA
	if endpoint == "" {
		var urlname string
		hostparts := strings.Split(r.Host, ".")
		if len(hostparts) > 1 {
			urlname = fmt.Sprintf("%s%s%s", pathOptions.GetString("EndpointPrefix"), hostparts[0], pathOptions.GetString("EndpointSuffix"))
		}

		if urlname != "" && LoadBalancers.Exists(urlname) {
			// The urlname is the name of a Pool, let it pass
			endpoint = urlname
		} else {
			// No cluster, and the urlname is not a valid Pool either
			ErrorOut.Printf("{%s} Switch request, but no endpoint set\n", requestID)
			RequestErrorResponse(r, w, "Invalid switch endpoint configuration", http.StatusBadRequest)
			return
		}
	}

	if pool, ok := LoadBalancers.Get(endpoint); ok {
		// we're skipping the error, because all Pool objects have been
		// materialzed by now, and would have errored out already
		handle, _ := pool.GetPool()
		DebugOut.Printf("{%s} Going to %s\n", requestID, endpoint)
		handle.ServeHTTP(w, r)
	} else {
		// not ok
		ErrorOut.Printf("{%s} Cannot find pool for requested endpoint '%s'\n", requestID, endpoint)
		RequestErrorResponse(r, w, "Invalid endpoint configuration", http.StatusBadRequest)
		return
	}

}

// GetSwitchName is a function to return the switch name in a request's context, if present
func GetSwitchName(request *http.Request) string {
	if oid := request.Context().Value(switchNameKey); oid != nil {
		return oid.(string)
	}
	return ""
}

// HandleReload waits 5 seconds after being called, and then rebuilds the SwitchMaps
func HandleReload(name string, mfiles map[string]string) {
	// In 5 seconds, Go-Go-Gadget Map Reload
	<-time.After(5 * time.Second)
	DebugOut.Printf("modified file: %s\n", name)
	for n, f := range mfiles {
		err := SwitchMaps.Load(n, f)
		if err != nil {
			ErrorOut.Printf("Error loading map file '%s' from watcher: %s\n", f, err)
		}
	}
}
