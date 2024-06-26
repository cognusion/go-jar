

# plugins
`import "github.com/cognusion/go-jar/plugins"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package plugins provides support for Yaegi-style http.Handler "plugins", for use
in JAR. Code should be self-contained. The standard library is available for
imports, anything else needs GoPaths and whatnot defined, and isnt't greatly
supported.

If the handler needs some configuration data, there is a `map[string]string`
that may be used. If so, ensure you have a function called
`SetConfig(map[string]string)` defined that will take the provided map and set it
in your code. See `testHandlerConfigSrc` in the included _test file.




## <a name="pkg-index">Index</a>
* [type HandlerPlugin](#HandlerPlugin)
  * [func NewHandlerPlugin(src, funcName string) (*HandlerPlugin, error)](#NewHandlerPlugin)
  * [func NewHandlerPluginWithConfig(src, funcName string, config map[string]string) (*HandlerPlugin, error)](#NewHandlerPluginWithConfig)
  * [func (h *HandlerPlugin) Bootstrap() (berr error)](#HandlerPlugin.Bootstrap)
  * [func (h *HandlerPlugin) CopyHandler() func(http.Handler) http.Handler](#HandlerPlugin.CopyHandler)
  * [func (h *HandlerPlugin) Handler(next http.Handler) http.Handler](#HandlerPlugin.Handler)
* [type InterpreterOptions](#InterpreterOptions)


#### <a name="pkg-files">Package files</a>
[plugins.go](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go)






## <a name="HandlerPlugin">type</a> [HandlerPlugin](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go?s=959:1121#L26)
``` go
type HandlerPlugin struct {
    Source   string
    FuncName string
    Config   map[string]string
    Options  InterpreterOptions
    // contains filtered or unexported fields
}

```
HandlerPlugin is a stuct to handle the creation, validation, and inclusion of
yaegi-loaded "plugins"







### <a name="NewHandlerPlugin">func</a> [NewHandlerPlugin](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go?s=2400:2467#L84)
``` go
func NewHandlerPlugin(src, funcName string) (*HandlerPlugin, error)
```
NewHandlerPlugin is a quick creator that also runs Bootrap, returning
an error if Bootstrap did, otherwise returning a reference to the created
HandlerPlugin.


### <a name="NewHandlerPluginWithConfig">func</a> [NewHandlerPluginWithConfig](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go?s=2791:2894#L100)
``` go
func NewHandlerPluginWithConfig(src, funcName string, config map[string]string) (*HandlerPlugin, error)
```
NewHandlerPluginWithConfig is a quick creator that also runs Bootrap, returning
an error if Bootstrap did, otherwise returning a reference to the created
HandlerPlugin.





### <a name="HandlerPlugin.Bootstrap">func</a> (\*HandlerPlugin) [Bootstrap](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go?s=1270:1318#L36)
``` go
func (h *HandlerPlugin) Bootstrap() (berr error)
```
Bootstrap is an initializer that *must* be called before Handler, and anytime
the public attributes (Source, FuncName, Options) are changed.




### <a name="HandlerPlugin.CopyHandler">func</a> (\*HandlerPlugin) [CopyHandler](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go?s=2141:2210#L77)
``` go
func (h *HandlerPlugin) CopyHandler() func(http.Handler) http.Handler
```
CopyHandler returns the Handler function itself.




### <a name="HandlerPlugin.Handler">func</a> (\*HandlerPlugin) [Handler](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go?s=1998:2061#L72)
``` go
func (h *HandlerPlugin) Handler(next http.Handler) http.Handler
```
Handler is a HandlerFunc representation of the HandlerPlugin




## <a name="InterpreterOptions">type</a> [InterpreterOptions](https://github.com/cognusion/go-jar/tree/master/plugins/plugins.go?s=810:850#L22)
``` go
type InterpreterOptions = interp.Options
```
InterpreterOptions is used to change how the interpreter runs. Aliased so including
packages needn't import interp directly.














- - -
Generated by [godoc2md](http://github.com/cognusion/godoc2md)
