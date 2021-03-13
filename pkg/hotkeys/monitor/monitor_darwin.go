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
	ErrInvalidEventReceived     = errors.New("received an invalid monitor event")
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

const appLoopQuitTimeout = time.Second

// CEngine implements monitor engine via C.
type CEngine struct {
	loopC chan struct{}

	handlerRefs [2]C.EventHandlerRef

	mouseEventTap C.CFMachPortRef
	mouseLoopSrc  C.CFRunLoopSourceRef
}

// Init initializes the engine for monitoring.
func (e *CEngine) Init() (ok bool) {
	if ok := e.initKeyboard(); !ok {
		return false
	}
	if ok := e.initMouse(); !ok {
		return false
	}
	e.loopC = make(chan struct{})
	return true
}

func (e *CEngine) initKeyboard() (ok bool) {
	// Note:
	// We use the older Carbon Event Manager instead of the newer, more
	// robust Quartz Event Services for keyboard events. This is because
	// currently, when iTerm2 is the foreground app, keyboard events are not being
	// received via Quartz Event Services; it appears as though iTerm2 always
	// fully consumes them. However, monitoring keyboard events as hotkeys via the
	// Carbon Event Manager, we are able to capture and consume the key events
	// before iTerm2 can.

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

func (e *CEngine) initMouse() (ok bool) {
	eventMask := uint(1<<C.kCGEventOtherMouseDown | 1<<C.kCGEventOtherMouseUp)
	eventTap := C.CGEventTapCreate(
		C.kCGHIDEventTap,
		C.kCGHeadInsertEventTap,
		0,
		C.CGEventMask(eventMask),
		(*[0]byte)(C.handleMouseButtonEvent),
		nil,
	)
	if eventTap == 0 {
		return false
	}
	e.mouseEventTap = eventTap

	C.CGEventTapEnable(eventTap, true)

	loopSrc := C.CFMachPortCreateRunLoopSource(
		C.kCFAllocatorDefault,
		e.mouseEventTap,
		0,
	)
	C.CFRunLoopAddSource(C.CFRunLoopGetMain(), loopSrc, C.kCFRunLoopCommonModes)
	e.mouseLoopSrc = loopSrc

	return true
}

// Start starts the engine for monitoring.
func (e *CEngine) Start(m *Monitor) error {
	if err := setGlobalMonitor(m); err != nil {
		return err
	}

	go e.runLoop()

	return nil
}

// Stop stops the engine from monitoring.
func (e *CEngine) Stop() {
	// QuitApplicationEventLoop() seems unstable, not always or constistantly
	// terminating RunApplicationEventLoop(). However, calling QuitEventLoop()
	// for the main event loop in addition seems to more reliably terminate the
	// main app loop. Just in case this still fails, we ignore the failed
	// termination after a timeout.
	C.QuitApplicationEventLoop()
	C.QuitEventLoop(C.GetMainEventLoop())
	select {
	case <-e.loopC:
	case <-time.After(appLoopQuitTimeout):
	}

	setGlobalMonitor(nil)
}

func (e *CEngine) runLoop() {
	C.RunApplicationEventLoop()
	e.loopC <- struct{}{}
}

// Deinit deinitializes the engine for monitoring.
func (e *CEngine) Deinit() (ok bool) {
	ok = true
	if ok := e.deinitKeyboard(); !ok {
		ok = false
	}
	if ok := e.deinitMouse(); !ok {
		ok = false
	}
	close(e.loopC)
	return
}

func (e *CEngine) deinitKeyboard() (ok bool) {
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

func (e *CEngine) deinitMouse() (ok bool) {
	if eventTap := e.mouseEventTap; eventTap != 0 {
		C.CGEventTapEnable(eventTap, false)
		e.mouseEventTap = 0
	}

	if loopSrc := e.mouseLoopSrc; loopSrc != 0 {
		C.CFRunLoopRemoveSource(
			C.CFRunLoopGetMain(),
			loopSrc,
			C.kCFRunLoopCommonModes,
		)
		e.mouseLoopSrc = 0
	}

	return true
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

	eEvent := hotkey.EngineKeyboardEvent(cEvent)
	hotkeyID, err := m.Hotkeys.IDFromEvent(eEvent)
	if err != nil {
		panic(err)
	}
	event := HotkeyEvent{hotkeyID, isOn, time.Now()}
	if err := m.Dispatch(event); err != nil {
		panic(err)
	}
}

//export goHandleMouseEvent
func goHandleMouseEvent(
	cEvent C.CGEventRef,
	cEventType C.CGEventType,
) C.CGEventRef {
	var isOn bool
	switch cEventType {
	case C.kCGEventOtherMouseDown:
		isOn = true
	case C.kCGEventOtherMouseUp:
		isOn = false
	case C.kCGEventTapDisabledByUserInput, C.kCGEventTapDisabledByTimeout:
		return cEvent
	default:
		panic(ErrInvalidEventReceived)
	}

	globalCMonitorMx.Lock()
	defer globalCMonitorMx.Unlock()

	m := globalCMonitor
	if m == nil {
		panic(ErrGlobalCMonitorMissing)
	}

	eEvent := hotkey.EngineMouseEvent(cEvent)
	hotkeyID, err := m.Hotkeys.IDFromEvent(eEvent)
	if err != nil {
		panic(err)
	}
	if hotkeyID == hotkey.NoID {
		return cEvent
	}

	event := HotkeyEvent{hotkeyID, isOn, time.Now()}
	if err := m.Dispatch(event); err != nil {
		panic(err)
	}
	return 0
}
