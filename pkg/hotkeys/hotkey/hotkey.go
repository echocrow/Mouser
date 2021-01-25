// Package hotkey (hotkey sub-package) provides keycodes and hotkey registries.
package hotkey

// #cgo CFLAGS: -D CGO
// #cgo darwin LDFLAGS: -framework Carbon
// #include "hotkey.h"
import "C"
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

// Registry holds a hotkey registry.
type Registry struct {
	ipd    IDProvider
	engine Engine
}

// NewRegistry constructs a new hotkey registry.
func NewRegistry(engine Engine, ipd IDProvider) Registry {
	if engine == nil {
		engine = CEngine{}
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

	keyIndex, err := NameToIndex(key)
	if err != nil {
		return 0, err
	}

	id := reg.ipd.NextID()
	if ok := reg.engine.Register(id, keyIndex); !ok {
		return 0, ErrRegistrationFailed
	}

	return id, nil
}

// Remove removes a hotkey to the reg registry.
func (reg Registry) Remove(id ID) {
	reg.engine.Unregister(id)
}

// CEngine implements hotkey engine via C.
type CEngine struct{}

// Register registers a hotkey via CEngine.
func (CEngine) Register(id ID, keyIndex KeyIndex) (ok bool) {
	return bool(C.registerHotkey(C.MouserHotKeyID(id), C.MouserKeyIndex(keyIndex)))
}

// Unregister unregisters a hotkey via CEngine.
func (CEngine) Unregister(id ID) {
	C.unregisterHotkey(C.MouserHotKeyID(id))
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
