// Package plugins provides support for Yaegi-style http.Handler "plugins", for use
// in JAR. Code should be self-contained. The standard library is available for
// imports, anything else needs GoPaths and whatnot defined, and isnt't greatly
// supported.
//
// If the handler needs some configuration data, there is a `map[string]string`
// that may be used. If so, ensure you have a function called
// `SetConfig(map[string]string)` defined that will take the provided map and set it
// in your code. See `testHandlerConfigSrc` in the included _test file.
package plugins

import (
	"fmt"
	"net/http"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// InterpreterOptions is used to change how the interpreter runs. Aliased so including
// packages needn't import interp directly.
type InterpreterOptions = interp.Options

// HandlerPlugin is a stuct to handle the creation, validation, and inclusion of
// yaegi-loaded "plugins"
type HandlerPlugin struct {
	Source   string
	FuncName string
	Config   map[string]string
	Options  InterpreterOptions
	hfunc    func(http.Handler) http.Handler
}

// Bootstrap is an initializer that *must* be called before Handler, and anytime
// the public attributes (Source, FuncName, Options) are changed.
func (h *HandlerPlugin) Bootstrap() (berr error) {
	i := interp.New(h.Options)
	i.Use(stdlib.Symbols)

	// recover on panic, and return the error
	defer func() {
		if recoveredPanic := recover(); recoveredPanic != nil {
			berr = fmt.Errorf("error during plugin boostrap: %v", recoveredPanic)
		}
	}()

	_, err := i.Eval(h.Source)
	if err != nil {
		return err
	}

	if h.Config != nil {
		v, cerr := i.Eval("SetConfig")
		if cerr != nil {
			return cerr
		}
		sc := v.Interface().(func(map[string]string))
		sc(h.Config)
	}

	v, err := i.Eval(h.FuncName)
	if err != nil {
		return err
	}

	h.hfunc = v.Interface().(func(http.Handler) http.Handler)

	return nil
}

// Handler is a HandlerFunc representation of the HandlerPlugin
func (h *HandlerPlugin) Handler(next http.Handler) http.Handler {
	return h.hfunc(next)
}

// CopyHandler returns the Handler function itself.
func (h *HandlerPlugin) CopyHandler() func(http.Handler) http.Handler {
	return h.hfunc
}

// NewHandlerPlugin is a quick creator that also runs Bootrap, returning
// an error if Bootstrap did, otherwise returning a reference to the created
// HandlerPlugin.
func NewHandlerPlugin(src, funcName string) (*HandlerPlugin, error) {
	h := HandlerPlugin{
		Source:   src,
		FuncName: funcName,
	}
	err := h.Bootstrap()
	if err != nil {
		return nil, err
	}
	return &h, nil

}

// NewHandlerPluginWithConfig is a quick creator that also runs Bootrap, returning
// an error if Bootstrap did, otherwise returning a reference to the created
// HandlerPlugin.
func NewHandlerPluginWithConfig(src, funcName string, config map[string]string) (*HandlerPlugin, error) {
	h := HandlerPlugin{
		Source:   src,
		FuncName: funcName,
		Config:   config,
	}
	err := h.Bootstrap()
	if err != nil {
		return nil, err
	}
	return &h, nil

}
