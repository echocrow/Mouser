package hotkey

// #cgo darwin LDFLAGS: -framework Carbon
// #include <Carbon/Carbon.h>
import "C"
import (
	"errors"
	"unsafe"
)

// EngineEvent is a platform-specific hotkey engine event.
type EngineEvent interface{}

// EngineKeyboardEvent is a platform-specific hotkey engine keyboard event.
type EngineKeyboardEvent unsafe.Pointer

// EngineMouseEvent is a platform-specific hotkey engine mouse event.
type EngineMouseEvent uintptr

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
	initKeyboardKeysLen uint = 8
	initMouseBtnsLen    uint = 2
)

func defaultEngine() Engine {
	return &CEngine{
		make(map[ID]C.EventHotKeyRef, initKeyboardKeysLen),
		make(map[MouseBtnCode]ID, initMouseBtnsLen),
	}
}

// mouserHotKeySig is the four-char code signature for mouser hotkey events.
const mouserHotKeySig uint = 'M'<<24 + 'S'<<16 + 'E'<<8 + 'R'<<0

// CEngine implements hotkey engine via C.
type CEngine struct {
	keyboardHkRefs map[ID]C.EventHotKeyRef

	mouseHkIDs map[MouseBtnCode]ID
}

// Register registers a hotkey via CEngine.
func (e *CEngine) Register(id ID, key KeyName) error {
	code, err := NameToCode(key)
	if err != nil {
		return err
	}
	switch c := code.(type) {
	case KeyCode:
		return e.registerKeyboard(id, c)
	case MouseBtnCode:
		return e.registerMouse(id, c)
	}
	return ErrRegistrationFailed
}

func (e *CEngine) registerKeyboard(id ID, keyCode KeyCode) error {
	modifiers := uint32(0)

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

	e.keyboardHkRefs[id] = hotkeyRef

	return nil
}

func (e *CEngine) registerMouse(id ID, btnCode MouseBtnCode) error {
	if _, ok := e.mouseHkIDs[btnCode]; ok {
		return ErrRegistrationFailed
	}
	e.mouseHkIDs[btnCode] = id
	return nil
}

// Unregister unregisters a hotkey via CEngine.
func (e *CEngine) Unregister(id ID) {
	if ref, ok := e.keyboardHkRefs[id]; ok {
		C.UnregisterEventHotKey(ref)
		delete(e.keyboardHkRefs, id)
	}
	for btnCode, hkID := range e.mouseHkIDs {
		if hkID == id {
			delete(e.mouseHkIDs, btnCode)
		}
	}
}

// IDFromEvent recovers the hotkey ID from an engine event.
func (e *CEngine) IDFromEvent(eEvent EngineEvent) (ID, error) {
	switch ee := eEvent.(type) {
	case EngineKeyboardEvent:
		return e.idFromHotkeyEvent(C.EventRef(ee))
	case EngineMouseEvent:
		return e.idFromMouseEvent(C.CGEventRef(ee))
	}
	return NoID, ErrInvalidEventReceived
}

func (e *CEngine) idFromHotkeyEvent(cEvent C.EventRef) (ID, error) {
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

func (e *CEngine) idFromMouseEvent(cEvent C.CGEventRef) (ID, error) {
	cBtnCode := C.CGEventGetIntegerValueField(
		cEvent,
		C.kCGMouseEventButtonNumber,
	)
	btnCode := MouseBtnCode(cBtnCode)
	if id, ok := e.mouseHkIDs[btnCode]; ok {
		return id, nil
	}
	return NoID, nil
}
