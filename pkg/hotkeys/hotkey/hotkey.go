// Package hotkey (hotkey sub-package) provides keycodes and hotkey registries.
package hotkey

import (
	"errors"
	"sync"
)

// Hotkey errors raised by package hotkey.
var (
	ErrRegistrationFailed = errors.New("hotkey registration failed")
	ErrIncompleteRegistry = errors.New("hotkey registry is incomplete")
)

// ID holds the ID of a hotkey.
type ID uint8

// NoID is the empty ID value.
const NoID ID = 0

// IDProvider describes a hotkey ID provider.
//go:generate mockery --name "IDProvider"
type IDProvider interface {
	NextID() ID
}

// Engine describes a hotkey registry engine.
//go:generate mockery --name "Engine"
type Engine interface {
	Register(id ID, key KeyName) error
	Unregister(id ID)
}

// Registrar describes a hotkey registrar.
//go:generate mockery --name "Registrar"
type Registrar interface {
	Add(key KeyName) (ID, error)
	Remove(id ID)
}

// Registry holds a hotkey registry.
type Registry struct {
	ipd    IDProvider
	engine Engine
}

// NewRegistry constructs a new hotkey registry.
func NewRegistry(engine Engine, ipd IDProvider) Registry {
	if engine == nil {
		engine = defaultEngine()
	}
	if ipd == nil {
		ipd = NewIDCounter()
	}
	return Registry{
		ipd:    ipd,
		engine: engine,
	}
}

// Add adds a new hotkey to the reg registry.
func (reg Registry) Add(key KeyName) (ID, error) {
	if reg.ipd == nil {
		return 0, ErrIncompleteRegistry
	}
	id := reg.ipd.NextID()
	if err := reg.engine.Register(id, key); err != nil {
		return 0, err
	}
	return id, nil
}

// Remove removes a hotkey to the reg registry.
func (reg Registry) Remove(id ID) {
	reg.engine.Unregister(id)
}

// IDCounter implements a simple incremental ID counter.
type IDCounter struct {
	nid ID
	mu  sync.Mutex
}

// NewIDCounter creates a new ID counter starting at 1.
func NewIDCounter() *IDCounter {
	return &IDCounter{nid: 1}
}

// NextID gets the next ID and increments the current ID counter.
func (idc *IDCounter) NextID() ID {
	idc.mu.Lock()
	defer idc.mu.Unlock()
	id := idc.nid
	idc.nid++
	return id
}
