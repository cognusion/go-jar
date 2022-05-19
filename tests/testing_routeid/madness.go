package madness

import (
	"net/http"
	"strings"
)

// PassThroughHandler is the best case handler, for benchmark comparisons
func PassThroughHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ParamInspection_FormValue_Handler uses the http.Request.FormValue facility to grab the ROUTEID off the paramlist.
// This is by far the fastest way to do this, but alters the body, rendering application-level checksums invalid, and
// breaking brittle destinations that make expectations about its condition.
func ParamInspection_FormValue_Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if rkey := r.FormValue("ROUTEID"); rkey != "" {
			r.AddCookie(&http.Cookie{
				Name:  "ROUTEID",
				Value: rkey,
				Path:  "/",
			})
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ParamInspection_URLQuery_Handler uses the http.Request.URL.Query().Get facility to grab the ROUTEID off the paramlist.
// This is a very expensive operation regardless of whether the parameter exists or not.
func ParamInspection_URLQuery_Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if rkey := r.URL.Query().Get("ROUTEID"); rkey != "" {
			r.AddCookie(&http.Cookie{
				Name:  "ROUTEID",
				Value: rkey,
				Path:  "/",
			})
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ParamInspection_URLQueryContains_Handler is a variation of ParamInspection_URLQuery_Handler, that first inspects the
// http.Request.URL.RawQuery string, to see if it Contains (or may contain) the ROUTEID, before decoding the Query, which
// is a very expensive operation.
func ParamInspection_URLQueryContains_Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.URL.RawQuery, "ROUTEID") {
			if rkey := r.URL.Query().Get("ROUTEID"); rkey != "" {
				r.AddCookie(&http.Cookie{
					Name:  "ROUTEID",
					Value: rkey,
					Path:  "/",
				})
			}
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

var (
	// CookieInspectionHandler is an http.HandlerFunc that returns 400 if the ROUTEID cookie isn't set
	CookieInspectionHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("ROUTEID")
		if err != nil || c.Value == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(gross)
			return
		}
		w.Write(ok)
	})

	// TestHandler is an http.HandlerFunc that quickly returns 200
	TestHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(ok)
	})

	noroute = []byte("No Routeid")
	ok      = []byte("ok")
	gross   = []byte("gross")
)
