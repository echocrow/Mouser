package monitor

// #cgo darwin CFLAGS: -D CGO
// #cgo darwin LDFLAGS: -framework Carbon
// #include "monitor_darwin.h"
import "C"
import (
	"errors"
	"sync"
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
)

// Monitor errors raised by package monitor.
var (
	ErrGlobalCMonitorMissing    = errors.New("global C monitor missing")
	ErrGlobalCMonitorAlreadySet = errors.New("global C monitor already set")
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
	globalCMonitorMx.Lock()
	defer globalCMonitorMx.Unlock()

	m := globalCMonitor
	if m == nil {
		panic(ErrGlobalCMonitorMissing)
	}

	eEvent := hotkey.EngineEvent(cEvent)
	hotkeyID, err := m.Hotkeys.IDFromEvent(eEvent)
	if err != nil {
		panic(err)
	}
	event := HotkeyEvent{hotkeyID, isOn, time.Now()}
	if err := m.Dispatch(event); err != nil {
		panic(err)
	}
}
