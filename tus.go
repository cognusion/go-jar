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
	Finishers["tus"] = nil
	FinisherSetups["tus"] = func(p *Path) (http.HandlerFunc, error) {
		var targetURI string
		if targetURI = p.Options.GetString(ConfigTUSTargetURI); targetURI == "" {
			return nil, ErrTUSTargetURIMissing
		}
		t, terr := tus.NewTUS(targetURI, p.Path)
		if terr != nil {
			return nil, terr
		}
		return http.StripPrefix(p.Path, t).ServeHTTP, nil
	}

	InitFuncs.Add(func() {
		tus.DebugOut = DebugOut
	})
}
