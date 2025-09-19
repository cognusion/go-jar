package jar

import (
	"github.com/cognusion/go-dictionary"

	"fmt"
	"runtime"
	"strings"
)

// Constants for configuration key strings
const (
	ConfigMacros = ConfigKey("macros")
)

var (
	// MacroDictionary is a Dictionary for doing mcro
	MacroDictionary dictionary.Resolver
)

func init() {

	InitFuncs.Add(func() {

		var s dictionary.SyncDict

		// Build dictionary
		d := map[string]string{
			"NAME":         "JAR",
			"VERSION":      VERSION,
			"SHORTVERSION": fmt.Sprintf("%s%s/%s%s", "%%", "NAME", "%%", "VERSION"),
			"FULLVERSION":  fmt.Sprintf("%s%s/%s%s %s %d:%d", "%%", "NAME", "%%", "VERSION", runtime.Version(), runtime.GOMAXPROCS(0), runtime.NumCPU()),
		}

		if dents := Conf.GetStringMapString(ConfigMacros); len(dents) > 0 {
			for k, v := range dents {
				d[strings.ToUpper(k)] = v
			}
		}

		s.AddValues(d) // add stringmap
		s.Resolve()    // resolves any embedded macros with static strings

		MacroDictionary = &s
	})

}
