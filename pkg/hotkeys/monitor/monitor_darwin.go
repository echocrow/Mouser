package monitor

// #cgo darwin CFLAGS: -D CGO
// #cgo darwin LDFLAGS: -framework Carbon
// #include "monitor_darwin.h"
import "C"
import (
	"errors"
	"sync"
	"time"
	"unsafe"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
)

// Monitor errors raised by package monitor.
var (
	ErrGlobalCMonitorMissing    = errors.New("global C monitor missing")
	ErrGlobalCMonitorAlreadySet = errors.New("global C monitor already set")
	ErrEventDetailsLookupFailed = errors.New("failed to get hotkey event details")
	ErrInvalidEventReceived     = errors.New("received an invalid hotkey event")
)

func defaultEngine() Engine {
	return &CEngine{}
}

var (
	globalCMonitor   *Monitor
	globalCMonitorMx sync.Mutex
)

func setGlobalMonitor(m *Monitor) error {
	globalCMonitorMx.Lock()
	defer globalCMonitorMx.Unlock()
	if globalCMonitor != nil && m != nil {
		return ErrGlobalCMonitorAlreadySet
	}
	globalCMonitor = m
	return nil
}

// CEngine implements monitor engine via C.
type CEngine struct {
	handlerRefs [2]C.EventHandlerRef
}

// Init initializes the engine for monitoring.
func (e *CEngine) Init() (ok bool) {
	evSpecDown := C.EventTypeSpec{C.kEventClassKeyboard, C.kEventHotKeyPressed}
	if status := C.InstallEventHandler(
		C.GetEventDispatcherTarget(),
		(*[0]byte)(C.handleHotkeyEventDown),
		1,
		&evSpecDown,
		nil,
		&e.handlerRefs[0],
	); status != C.noErr {
		return false
	}

	evSpecUp := C.EventTypeSpec{C.kEventClassKeyboard, C.kEventHotKeyReleased}
	if status := C.InstallEventHandler(
		C.GetEventDispatcherTarget(),
		(*[0]byte)(C.handleHotkeyEventUp),
		1,
		&evSpecUp,
		nil,
		&e.handlerRefs[1],
	); status != C.noErr {
		return false
	}

	return true
}

// Start starts the engine for monitoring.
func (*CEngine) Start(m *Monitor) error {
	if err := setGlobalMonitor(m); err != nil {
		return err
	}

	go C.RunApplicationEventLoop()

	return nil
}

// Stop stops the engine from monitoring.
func (*CEngine) Stop() {
	C.QuitApplicationEventLoop()

	setGlobalMonitor(nil)
}

// Deinit deinitializes the engine for monitoring.
func (e *CEngine) Deinit() (ok bool) {
	ok = true

	if status := C.RemoveEventHandler(e.handlerRefs[0]); status != C.noErr {
		ok = false
	}
	e.handlerRefs[0] = nil

	if status := C.RemoveEventHandler(e.handlerRefs[1]); status != C.noErr {
		ok = false
	}
	e.handlerRefs[1] = nil

	return
}

//export goHandleHotkeyEvent
func goHandleHotkeyEvent(
	cEvent C.EventRef,
	isOn bool,
) {
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
		panic(ErrEventDetailsLookupFailed)
	}

	if uint(cEventID.signature) != hotkey.MouserHotKeySig {
		panic(ErrInvalidEventReceived)
	}

	globalCMonitorMx.Lock()
	defer globalCMonitorMx.Unlock()

	m := globalCMonitor
	if m == nil {
		panic(ErrGlobalCMonitorMissing)
	}

	hotkeyID := hotkey.ID(cEventID.id)
	event := HotkeyEvent{hotkeyID, isOn, time.Now()}
	if err := m.Dispatch(event); err != nil {
		panic(err)
	}
}
