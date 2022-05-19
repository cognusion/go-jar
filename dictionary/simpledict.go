package dictionary

import (
	"fmt"
	"strings"
)

// SimpleDict is a string map Dictionary to do simple key-value replacements
type SimpleDict map[string]string

// Resolve walks the dictionary and updates any values that contain macros with static strings.
func (m *SimpleDict) Resolve() {

	newm := make(map[string]string)
	for k, v := range *m {
		if strings.Contains(v, "%%") {
			newm[k] = m.Replacer(v)
		} else {
			newm[k] = v
		}
	}

	*m = newm
}

// Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values.
// Note that shortest-prefixes *may* match first, so for dictionaries of %%VERSION1="one" and %%VERSION12="twelve" you may find
// cases where %%VERSION12 is expanded to "twelve" or "one2". WONTFIX
func (m *SimpleDict) Replacer(in string) string {

	var (
		out      = in
		replaced bool
		c        int
	)

	for {
		out, replaced = m.replacer(out)
		if !replaced || c > 20 {
			// We're done, or circuit breaker popped
			break
		}
		c++
	}

	return out
}

// replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values, returning false when there was nothing to replace
func (m *SimpleDict) replacer(in string) (string, bool) {
	if !strings.Contains(in, "%%") {
		// Don't waste energy
		return in, false
	}

	out := in
	rep := false

	for k, v := range *m {
		before := out
		out = strings.Replace(out, fmt.Sprintf("%s%s", "%%", k), v, -1)
		if before != out {
			rep = true
		}
		if !strings.Contains(out, "%%") {
			// Let's be done here
			break
		}
	}
	return out, rep
}
