package jar

import (
	"github.com/cognusion/go-jar/tus"

	"net/http"
	"strings"
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
		var (
			targetURI string
			t         *tus.TUS
			err       error
		)
		if targetURI = p.Options.GetString(ConfigTUSTargetURI); targetURI == "" {
			// Missing target!
			return nil, ErrTUSTargetURIMissing
		} else if strings.HasPrefix(strings.ToLower(targetURI), "s3://") {
			// S3 target
			t, err = tus.NewTUSwithS3(targetURI, p.Path, AWSSession.S3Client())
		} else if strings.HasPrefix(strings.ToLower(targetURI), "file://") {
			// File target
			t, err = tus.NewTUS(targetURI, p.Path)
		} else {
			// Bad prefix
			err = tus.ErrBadTargetPrefix
		}

		if err != nil {
			return nil, err
		}
		return http.StripPrefix(p.Path, t).ServeHTTP, nil
	}

	InitFuncs.Add(func() {
		tus.DebugOut = DebugOut
	})
}
