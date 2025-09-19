

# dictionary
`import "github.com/cognusion/go-jar/dictionary"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
Package dictionary provides the Dictionary interface for abstraction, and a simple stringmap implementation called SimpleDict, used for macro
definition and replacement.
Dictionary is an abstraction from JAR.

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>




## <a name="pkg-index">Index</a>
* [type Dictionary](#Dictionary)
* [type Resolver](#Resolver)
* [type SimpleDict](#SimpleDict)
  * [func (m *SimpleDict) Replacer(in string) string](#SimpleDict.Replacer)
  * [func (m *SimpleDict) Resolve()](#SimpleDict.Resolve)
* [type SyncDict](#SyncDict)
  * [func (m *SyncDict) AddValues(values map[string]string)](#SyncDict.AddValues)
  * [func (m *SyncDict) Replacer(in string) string](#SyncDict.Replacer)
  * [func (m *SyncDict) Resolve()](#SyncDict.Resolve)


#### <a name="pkg-files">Package files</a>
[dict.go](https://github.com/cognusion/go-jar/tree/master/dictionary/dict.go) [simpledict.go](https://github.com/cognusion/go-jar/tree/master/dictionary/simpledict.go) [syncdict.go](https://github.com/cognusion/go-jar/tree/master/dictionary/syncdict.go)






## <a name="Dictionary">type</a> [Dictionary](https://github.com/cognusion/go-jar/tree/master/dictionary/dict.go?s=431:604#L11)
``` go
type Dictionary interface {
    // Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values
    Replacer(string) string
}
```
Dictionary is an interface for macro-expanding structures

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>










## <a name="Resolver">type</a> [Resolver](https://github.com/cognusion/go-jar/tree/master/dictionary/dict.go?s=798:1080#L19)
``` go
type Resolver interface {
    // Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values
    Replacer(string) string
    // Resolve is intended to walk the dictionary and replace any dictionary items with static strings
    Resolve()
}
```
Resolver is an interface for macro-expanded structures that also can replace their own embedded macros with static strings

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>










## <a name="SimpleDict">type</a> [SimpleDict](https://github.com/cognusion/go-jar/tree/master/dictionary/simpledict.go?s=193:226#L11)
``` go
type SimpleDict map[string]string
```
SimpleDict is a string map Dictionary to do simple key-value replacements

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>










### <a name="SimpleDict.Replacer">func</a> (\*SimpleDict) [Replacer](https://github.com/cognusion/go-jar/tree/master/dictionary/simpledict.go?s=972:1019#L35)
``` go
func (m *SimpleDict) Replacer(in string) string
```
Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values.
Note that shortest-prefixes *may* match first, so for dictionaries of %%VERSION1="one" and %%VERSION12="twelve" you may find
cases where %%VERSION12 is expanded to "twelve" or "one2". WONTFIX

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>




### <a name="SimpleDict.Resolve">func</a> (\*SimpleDict) [Resolve](https://github.com/cognusion/go-jar/tree/master/dictionary/simpledict.go?s=390:420#L16)
``` go
func (m *SimpleDict) Resolve()
```
Resolve walks the dictionary and updates any values that contain macros with static strings.

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>




## <a name="SyncDict">type</a> [SyncDict](https://github.com/cognusion/go-jar/tree/master/dictionary/syncdict.go?s=276:332#L12)
``` go
type SyncDict struct {
    // contains filtered or unexported fields
}

```
SyncDict is a string map Dictionary to do simple key-value replacements. It is goro-safe for updates, and optimized for read-mostly implementations.

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>










### <a name="SyncDict.AddValues">func</a> (\*SyncDict) [AddValues](https://github.com/cognusion/go-jar/tree/master/dictionary/syncdict.go?s=465:519#L20)
``` go
func (m *SyncDict) AddValues(values map[string]string)
```
AddValues adds the key/value pairs listed to the the SyncDict

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>




### <a name="SyncDict.Replacer">func</a> (\*SyncDict) [Replacer](https://github.com/cognusion/go-jar/tree/master/dictionary/syncdict.go?s=1203:1248#L48)
``` go
func (m *SyncDict) Replacer(in string) string
```
Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>




### <a name="SyncDict.Resolve">func</a> (\*SyncDict) [Resolve](https://github.com/cognusion/go-jar/tree/master/dictionary/syncdict.go?s=778:806#L32)
``` go
func (m *SyncDict) Resolve()
```
Resolve walks the dictionary and updates any values that contain macros with static strings.

Deprecated: Use <a href="https://github.com/cognusion/go-dictionary/">https://github.com/cognusion/go-dictionary/</a>








- - -
Generated by [godoc2md](http://github.com/cognusion/godoc2md)
