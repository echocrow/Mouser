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

// Engine describes the interface of a hotkey registry engine.
type Engine interface {
	register(id ID, keyIndex KeyIndex) (ok bool)
	unregister(id ID)
}

// Registry holds a hotkey registry.
type Registry struct {
	idc    *idCounter
	engine Engine
}

// NewRegistry constructs a new hotkey registry.
func NewRegistry(engine Engine, idc *idCounter) Registry {
	if engine == nil {
		engine = CEngine{}
	}
	if idc == nil {
		idc = newIDCounter()
	}
	return Registry{
		idc:    idc,
		engine: engine,
	}
}

// Register adds a new hotkey to the reg registry.
func (reg Registry) Register(key KeyName) (ID, error) {
	if reg.idc == nil {
		return 0, ErrIncompleteRegistry
	}

	keyIndex, err := NameToIndex(key)
	if err != nil {
		return 0, err
	}

	id := reg.idc.nextID()
	if ok := reg.engine.register(id, keyIndex); !ok {
		return 0, ErrRegistrationFailed
	}

	return id, nil
}

// Unregister removes a hotkey to the reg registry.
func (reg Registry) Unregister(id ID) {
	reg.engine.unregister(id)
}

// CEngine implements hotkey engine via C.
type CEngine struct{}

func (CEngine) register(id ID, keyIndex KeyIndex) (ok bool) {
	return bool(C.registerHotkey(C.MouserHotKeyID(id), C.MouserKeyIndex(keyIndex)))
}
func (CEngine) unregister(id ID) {
	C.unregisterHotkey(C.MouserHotKeyID(id))
}

// idCounter implements a simple incremental ID counter.
type idCounter struct {
	nid ID
	mu  sync.Mutex
}

// newIDCounter creates a new ID counter starting at 1.
func newIDCounter() *idCounter {
	return &idCounter{nid: 1}
}

// nextID gets the next ID and increments the current ID counter.
func (idc *idCounter) nextID() ID {
	idc.mu.Lock()
	defer idc.mu.Unlock()
	id := idc.nid
	idc.nid++
	return id
}
