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
	IDFromEvent(EngineEvent) (ID, error)
}

// Registrar describes a hotkey registrar.
//go:generate mockery --name "Registrar"
type Registrar interface {
	Add(key KeyName) (ID, error)
	Remove(id ID)
	IDFromEvent(EngineEvent) (ID, error)
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

// IDFromEvent recovers the hotkey from an engine event.
func (reg Registry) IDFromEvent(eEvent EngineEvent) (ID, error) {
	return reg.engine.IDFromEvent(eEvent)
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

// IDSet holds a unique, non-sorted set of hotkey IDs
type IDSet struct {
	m map[ID]struct{}
}

// NewIDSet instantiates a new IDSet.
func NewIDSet() IDSet {
	return IDSet{make(map[ID]struct{})}
}

// Add adds ID id to IDSet s.
func (s IDSet) Add(id ID) {
	s.m[id] = struct{}{}
}

// Has checks whether IDSet s contains ID id.
func (s IDSet) Has(id ID) bool {
	_, ok := s.m[id]
	return ok
}

// Del removes ID id to IDSet s.
func (s IDSet) Del(id ID) {
	delete(s.m, id)
}
