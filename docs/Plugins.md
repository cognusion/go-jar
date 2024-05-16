# Plugins

## Background

There is disabled-by-default support for dynamic handlers. [Yaegi](https://github.com/traefik/yaegi) is awesome for short-lived and carefully-tailored applications, but unfortunately doesn't participate in GC, is too slow for request/response actions, leaks heap, and has a bunch of interface-related gotchas that preclude its use. Below are the results from a trivial http.HandlerFunc that outputs a static `[]byte` implemented natively and `Eval`d by Yaegi:

```
Benchmark_NativeHandler-12    	51765292	        23.83 ns/op	      41 B/op	       0 allocs/op
Benchmark_YaegiHandler-12     	  630565	        1720 ns/op	     818 B/op	      20 allocs/op
```

That said, all the kids are using it these days since *two magnitudes* of performance slide and a ~20% larger binary is ok. 

## Activation

Building or testing with `--tags plugins` will activate the plugins system.

```bash
go-jar$ go test -run Plugin
testing: warning: no tests to run
PASS
ok  	github.com/cognusion/go-jar	0.010s
go-jar$ go test -run Plugin --tags plugins
14 total assertions
PASS
ok  	github.com/cognusion/go-jar	0.025s

```

## Rules & Configuration

### Code 

Handler code should be self-contained, including any imports needed. See `tests/plugins/testhandler.src`. Additionally, if you are importing anything outside of the standard library, you may need to set a `GoPath` in your configuration or it may refuse to import properly.

A minimal config snippet to define a plugins:

```yaml
plugins:
  helloworld:
    path: tests/plugins/testhandler.src
    name: TestHandler
```

The full struct is reproduced below for convenience:

```go
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
```

### Hooks

Assuming you've defined the plugin properly per the above, you can now reference `helloworld` anywhere you would normally reference a handler.

## Caveats

* Compiled binaries grow 10-12MB in size, depending on architecture
* GC is *quite different* JAR-wide
* Requests traversing plugins will take demonstrably longer
* Plugins are not re-read after the initial boostrap, nor are their config
* If you need the same handler with different configs, you need to define it multiple times in `plugins`