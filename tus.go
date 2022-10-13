package jar

import (
	"github.com/cognusion/go-jar/tus"

	"net/http"
)

/*
	TODO
	* Configurable triggers on events
	  * Zulip worker integration
	  * MQ integration
*/

// Constants for configuration key strings and Errors
const (
	ConfigTUSTargetURI     = ConfigKey("tus.targeturi")
	ErrTUSTargetURIMissing = Error("tus.targeturi missing from path options")
)

func init() {
	// Set up the static finishers
	Finishers["tus"] = tusFinisher
	FinisherSetups["tus"] = func(p *Path) error {
		if !p.Options.GetBool(ConfigTUSTargetURI) {
			return ErrTUSTargetURIMissing
		}
		return nil
	}
}

// tusFinisher implements the TUS endpoint dynamically
func tusFinisher(w http.ResponseWriter, r *http.Request) {
	var (
		pathOptions PathOptions
	)

	if popts := r.Context().Value(PathOptionsKey); popts != nil {
		pathOptions = popts.(PathOptions)
	}

	targetURI := pathOptions.GetString(ConfigTUSTargetURI)
	tus, err := tus.NewTUS(targetURI, r.RequestURI)
	if err != nil {
		// Ugh run-time error
		ErrorOut.Printf("%s%s\n", ErrRequestError{r, "Error creating a TUS"}, err.Error())
		RequestErrorResponse(r, w, "There was an error creating the data upload endpoint", http.StatusInternalServerError)
		return
	}

	tus.ServeHTTP(w, r)
}
