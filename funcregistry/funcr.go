package funcregistry

import (
	"io"
	"log"
	"sync"
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "", 0)
)

// FuncRegistry is an aggregation of func()s that should be called during an orderly shutdown/restart.
// Examples include context.CancelFuncs, resource closing, etc.
type FuncRegistry struct {
	funcLock  sync.Mutex
	funcs     []func()
	onceFuncs []func()
	callOnce  bool
	invoked   bool
}

// NewFuncRegistry returns an initialized Stopper
func NewFuncRegistry(callOnce bool) *FuncRegistry {
	return &FuncRegistry{
		funcs:     make([]func(), 0),
		onceFuncs: make([]func(), 0),
		callOnce:  callOnce,
	}
}

// Close will zero out the func list, and trip the invocation breaker
// preventing further use
func (s *FuncRegistry) Close() {
	s.funcLock.Lock()
	defer s.funcLock.Unlock()

	s.invoked = true
	s.funcs = make([]func(), 0)
	s.onceFuncs = make([]func(), 0)
}

// Add will append f to the list of funcs to call, if Call() is called
func (s *FuncRegistry) Add(f func()) {
	s.funcLock.Lock()
	defer s.funcLock.Unlock()
	if !s.invoked {
		s.funcs = append(s.funcs, f)
	}
}

// AddOnce will append f to the list of funcs to call, if Call() is called,
// but will only be called once, regardless of the number of Call()s.
func (s *FuncRegistry) AddOnce(f func()) {
	s.funcLock.Lock()
	defer s.funcLock.Unlock()
	if !s.invoked {
		s.onceFuncs = append(s.onceFuncs, f)
	}
}

// Call will iterate over the registered functions, calling them.
// Call may only be called once with effect if callOnce is true,
// but is safe to call multiple times, regardless.
func (s *FuncRegistry) Call() {
	s.funcLock.Lock()
	defer s.funcLock.Unlock()

	if !s.invoked {
		if s.callOnce {
			s.invoked = true
		}

		for _, f := range s.funcs {
			f()
		}

		if len(s.onceFuncs) > 0 {
			for _, f := range s.onceFuncs {
				f()
			}
			s.onceFuncs = make([]func(), 0)
		}
	}
	DebugOut.Printf("FuncRegistry.Call called!\n")
}
