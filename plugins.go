//go:build plugins

package jar

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cognusion/go-jar/plugins"
	"github.com/spf13/viper"
)

func init() {
	// Creates any plugins defined in the config, and adds them to the Handlers map.
	InitFuncs.Add(pluginHandlerMap)
}

// pluginHandlerMap is a panic-on-error setter called by InitFuncs to add config-defined
// plugins to the Handlers map.
func pluginHandlerMap() {
	for k := range Conf.GetStringMap("plugins") {
		keyname := fmt.Sprintf("plugins.%s", k)
		h, err := PluginHandler(keyname, Conf)
		if err != nil {
			panic(err) // TODO: !!!
		}
		Handlers[strings.ToLower((k))] = h
	}
}

// PluginHandler is a glue function that takes a config key and returns either a
// handler function or an error, using the global Conf config.
func PluginHandler(name string, conf *viper.Viper) (func(http.Handler) http.Handler, error) {
	// create a PluginConfig
	if pc, err := NewPluginConfig(name, conf); err != nil {
		return nil, err
	} else if h, err := pc.CreateHandler(); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}

// PluginConfig is a marshallable configuration structure with useful member functions.
type PluginConfig struct {
	// Path is the full path to file to load
	Path string

	// Name is the function or package.function that is the handler
	// we call.
	Name string

	// Config is any configuration information the handler itself needs.
	// If used, ensure there is a `SetConfig(map[string]string)` function
	// in the plugin so receive this properly. See tests for details.
	Config map[string]string

	// GoPath sets GOPATH for the interpreter.
	GoPath string

	// BuildTags sets build constraints for the interpreter.
	BuildTags []string

	// Args are the cmdline args fed to the interpreter, defaults to os.Args.
	Args []string

	// Env is the environment of interpreter. Entries are in the form "key=values".
	Env []string

	// Unrestricted allows to run non sandboxed stdlib symbols such as os/exec and environment
	Unrestricted bool
}

// NewPluginConfig attempts to unmarshal a subconfiguration into a PluginConfig. Returning either
// a reference to it, or an error.
func NewPluginConfig(name string, conf *viper.Viper) (*PluginConfig, error) {
	c := PluginConfig{}
	if err := conf.UnmarshalKey(name, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateHandler will attempt to create a HandlerPlugin from the parent config,
// returning either the handler or an error.
func (pc *PluginConfig) CreateHandler() (func(http.Handler) http.Handler, error) {
	options := pc.buildOpts()

	buff := RecyclableBufferPool.Get()
	defer buff.Close()

	f, err := os.Open(pc.Path)
	if err != nil {
		return nil, err
	}
	buff.ResetFromReader(f)
	f.Close()

	plugin := plugins.HandlerPlugin{
		FuncName: pc.Name,
		Options:  *options,
		Source:   buff.String(),
		Config:   pc.Config,
	}

	if err = plugin.Bootstrap(); err != nil {
		return nil, err
	}

	return plugin.CopyHandler(), nil
}

// buildOpts returns a reference to an InterpreterOptions based on the parent
// config.
func (pc *PluginConfig) buildOpts() *plugins.InterpreterOptions {
	// Check for options overrides
	opts := plugins.InterpreterOptions{} // start with the defaults

	if len(pc.Args) > 0 {
		opts.Args = pc.Args
	}

	if pc.GoPath != "" {
		opts.GoPath = pc.GoPath
	}

	if len(pc.BuildTags) > 0 {
		opts.BuildTags = pc.BuildTags
	}

	if len(pc.Env) > 0 {
		opts.Env = pc.Env
	}

	return &opts
}
