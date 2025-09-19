// Package dictionary provides the Dictionary interface for abstraction, and a simple stringmap implementation called SimpleDict, used for macro
// definition and replacement.
// Dictionary is an abstraction from JAR.
//
// Deprecated: Use https://github.com/cognusion/go-dictionary/
package dictionary

// Dictionary is an interface for macro-expanding structures
//
// Deprecated: Use https://github.com/cognusion/go-dictionary/
type Dictionary interface {
	// Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values
	Replacer(string) string
}

// Resolver is an interface for macro-expanded structures that also can replace their own embedded macros with static strings
//
// Deprecated: Use https://github.com/cognusion/go-dictionary/
type Resolver interface {
	// Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values
	Replacer(string) string
	// Resolve is intended to walk the dictionary and replace any dictionary items with static strings
	Resolve()
}
