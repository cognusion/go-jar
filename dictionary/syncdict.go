package dictionary

import (
	"fmt"
	"strings"
	"sync"
)

// SyncDict is a string map Dictionary to do simple key-value replacements. It is goro-safe for updates, and optimized for read-mostly implementations.
type SyncDict struct {
	dict sync.Map
	lock sync.Mutex
}

// AddValues adds the key/value pairs listed to the the SyncDict
func (m *SyncDict) AddValues(values map[string]string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for k, v := range values {
		m.dict.Store(k, v)
	}
}

// Resolve walks the dictionary and updates any values that contain macros with static strings.
func (m *SyncDict) Resolve() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dict.Range(func(key, value interface{}) bool {
		if strings.Contains(value.(string), "%%") {
			m.dict.Store(key, m.Replacer(value.(string)))
		}
		return true
	})

}

// Replacer takes a string, and expands any %%-prefixed strings registered as macros, with their corresponding values
func (m *SyncDict) Replacer(in string) string {
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
func (m *SyncDict) replacer(in string) (string, bool) {
	if !strings.Contains(in, "%%") {
		// Don't waste energy
		return in, false
	}

	out := in
	rep := false

	m.dict.Range(func(key, value interface{}) bool {
		before := out
		out = strings.Replace(out, fmt.Sprintf("%s%s", "%%", key.(string)), value.(string), -1)
		if before != out {
			rep = true
		}
		if !strings.Contains(out, "%%") {
			// Let's be done here
			return false
		}
		return true
	})

	return out, rep
}
