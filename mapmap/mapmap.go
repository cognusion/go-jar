package mapmap

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// MapMap is essentially a goro-safe map[string]map[string]string specialized for handling multi-URL "URL Routes".
// MapMap is optimized for read-mostly patterns
type MapMap struct {
	sync.Mutex                   // used ONLY BY WRITERS
	m               atomic.Value // map[string]map[string]string
	idMapName       string
	endpointMapName string
}

// NewMapMap returns an initiaized MapMap
func NewMapMap() *MapMap {
	return NewMapMapWithMapNames("ids", "endpoints")
}

// NewMapMapWithMapNames returns an initiaized MapMap, with the map names set to non-default values.
func NewMapMapWithMapNames(idMap, endpointMap string) *MapMap {
	m := MapMap{}
	m.m.Store(make(mapmap))
	m.idMapName = idMap
	m.endpointMapName = endpointMap
	return &m
}

// mapmap is a map[string]map[string]string which is the
// underlying data type of a MapMap, but without the safety
// of atomicity and writer locks
type mapmap map[string]map[string]string

// URLRoute is an informative structure
type URLRoute struct {
	Name     string
	ID       string
	Endpoint string
}

// Get takes a map name and key, and returns the stored value or an empty string
func (o *MapMap) Get(name, key string) string {
	m := o.m.Load().(mapmap)
	if i, ok := m[name][key]; ok {
		return i
	}
	return ""
}

// GetURLRoute takes a routeName (first block of a hostname) and may return a reconstituted URLRoute.
// The Name is populated by the ``routeName`` requested.
// The ID field is pulled from the value of the ``ids`` map with the ``routeName`` as key.
// The Endpoint field is pulled from the value of the ``endpoints`` map with the ``routeName`` as key.
// Partial results are returns as applicable.
func (o *MapMap) GetURLRoute(routeName string) *URLRoute {
	oo := URLRoute{Name: routeName}
	m := o.m.Load().(mapmap)

	if i, ok := m[o.idMapName][routeName]; ok {
		oo.ID = i
	}
	if i, ok := m[o.endpointMapName][routeName]; ok {
		oo.Endpoint = i
	}
	return &oo
}

// Set takes an "Org Map" name, and replaces its string map with the provided one
func (o *MapMap) Set(name string, newmap map[string]string) {
	o.Lock()
	m := o.m.Load().(mapmap)
	m[name] = newmap
	o.m.Store(m)
	o.Unlock()
}

// Size returns the number of toplevel keys in the MapMap
func (o *MapMap) Size() int {
	o.Lock()
	m := o.m.Load()
	o.Unlock()
	if m == nil {
		return 0
	}
	return len(m.(mapmap))
}

// Load takes an "Org Map" name and a filename, and builds a new "Org Map" from the file, .Set()ing it
func (o *MapMap) Load(name, filename string) error {
	m := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	line := 1
	for scanner.Scan() {
		lineparts := strings.Split(strings.TrimSpace(scanner.Text()), " ")
		if len(lineparts) == 2 {
			m[lineparts[0]] = lineparts[1]
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	o.Set(name, m)

	return nil
}
