package hotkey

// #cgo darwin LDFLAGS: -framework Carbon
// #include <Carbon/Carbon.h>
import "C"

// MouserHotKeySig is the four-char code signature for mouser hotkey events.
const MouserHotKeySig uint = 'M'<<24 + 'S'<<16 + 'E'<<8 + 'R'<<0

const (
	initKeysLen uint = 8
)

func defaultEngine() Engine {
	return &CEngine{
		make(map[ID]C.EventHotKeyRef, initKeysLen),
	}
}

// CEngine implements hotkey engine via C.
type CEngine struct {
	refs map[ID]C.EventHotKeyRef
}

// Register registers a hotkey via CEngine.
func (e *CEngine) Register(id ID, keyCode KeyCode) (ok bool) {
	modifiers := uint32(0)

	eventID := C.EventHotKeyID{
		C.uint(MouserHotKeySig),
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
		return false
	}

	e.refs[id] = hotkeyRef

	return true
}

// Unregister unregisters a hotkey via CEngine.
func (e *CEngine) Unregister(id ID) {
	if ref, ok := e.refs[id]; ok {
		C.UnregisterEventHotKey(ref)
		delete(e.refs, id)
	}
}
