package hotkey

// #cgo darwin LDFLAGS: -framework Carbon
// #include <Carbon/Carbon.h>
import "C"
import (
	"errors"
	"unsafe"
)

// EngineEvent is a platform-specific hotkey engine event.
type EngineEvent unsafe.Pointer

// MockEngineEvent returns an empty mock engine event.
func MockEngineEvent() EngineEvent {
	return nil
}

// Hotkey engine errors raised by package hotkey.
var (
	ErrEventDetailsLookupFailed = errors.New("failed to get hotkey event details")
	ErrInvalidEventReceived     = errors.New("received an invalid hotkey event")
)

const (
	initKeysLen uint = 8
)

func defaultEngine() Engine {
	return &CEngine{
		make(map[ID]C.EventHotKeyRef, initKeysLen),
	}
}

// mouserHotKeySig is the four-char code signature for mouser hotkey events.
const mouserHotKeySig uint = 'M'<<24 + 'S'<<16 + 'E'<<8 + 'R'<<0

// CEngine implements hotkey engine via C.
type CEngine struct {
	refs map[ID]C.EventHotKeyRef
}

// Register registers a hotkey via CEngine.
func (e *CEngine) Register(id ID, key KeyName) error {
	modifiers := uint32(0)

	keyCode, err := NameToCode(key)
	if err != nil {
		return err
	}

	eventID := C.EventHotKeyID{
		C.uint(mouserHotKeySig),
		C.uint(id),
	}

	var hotkeyRef C.EventHotKeyRef
	if status := C.RegisterEventHotKey(
		C.uint(keyCode),
		C.uint(modifiers),
		eventID,
		C.GetEventDispatcherTarget(),
		0,
		&hotkeyRef,
	); status != C.noErr {
		return ErrRegistrationFailed
	}

	e.refs[id] = hotkeyRef

	return nil
}

// Unregister unregisters a hotkey via CEngine.
func (e *CEngine) Unregister(id ID) {
	if ref, ok := e.refs[id]; ok {
		C.UnregisterEventHotKey(ref)
		delete(e.refs, id)
	}
}

// IDFromEvent recovers the hotkey ID from an engine event.
func (e *CEngine) IDFromEvent(eEvent EngineEvent) (ID, error) {
	cEvent := C.EventRef(eEvent)
	var cEventID C.EventHotKeyID
	if status := C.GetEventParameter(
		cEvent,
		C.kEventParamDirectObject,
		C.typeEventHotKeyID,
		nil,
		C.ulong(unsafe.Sizeof(cEventID)),
		nil,
		unsafe.Pointer(&cEventID),
	); status != C.noErr {
		return NoID, ErrEventDetailsLookupFailed
	}

	if uint(cEventID.signature) != mouserHotKeySig {
		return NoID, ErrInvalidEventReceived
	}

	return ID(cEventID.id), nil
}
