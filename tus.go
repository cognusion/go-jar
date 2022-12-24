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
	ConfigTUSTargetURI      = ConfigKey("tus.targeturi")
	ConfigTUSAppendFilename = ConfigKey("tus.appendfilename")
	ErrTUSTargetURIMissing  = Error("tus.targeturi missing from path options")
)

func init() {
	// Set up the static finishers
	Finishers["tus"] = nil
	FinisherSetups["tus"] = func(p *Path) (http.HandlerFunc, error) {
		var (
			targetURI string
			t         *tus.TUS
			conf      tus.Config
			err       error
		)
		if targetURI = p.Options.GetString(ConfigTUSTargetURI); targetURI == "" {
			// Missing target!
			return nil, ErrTUSTargetURIMissing
		}

		// Prep config
		conf = tus.Config{
			TargetURI:      p.Options.GetString(ConfigTUSTargetURI),
			AppendFilename: p.Options.GetBool(ConfigTUSAppendFilename),
		}

		if strings.HasPrefix(strings.ToLower(targetURI), "s3://") {
			// S3 target
			conf.S3Client = AWSSession.S3Client()
			t, err = tus.New(p.Path, conf)
		} else if strings.HasPrefix(strings.ToLower(targetURI), "file://") {
			// File target
			t, err = tus.New(p.Path, conf)
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
		tus.ErrorOut = ErrorOut
	})
}
