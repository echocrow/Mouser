// Package monitor allows listening to and dispatching hotkey events.
package monitor

// #cgo CFLAGS: -D CGO
// #cgo darwin LDFLAGS: -framework Carbon
// #include "monitor.h"
import "C"
import (
	"errors"
	"sync"
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/birdkid/mouser/pkg/log"
)

// Monitor errors raised by package monitor.
var (
	ErrInitFailed               = errors.New("monitor initialization failed")
	ErrDeinitFailed             = errors.New("monitor de-initialization failed")
	ErrAlreadyStarted           = errors.New("monitor already started")
	ErrNotYetStarted            = errors.New("monitor not yet started")
	ErrGlobalCMonitorMissing    = errors.New("global C monitor missing")
	ErrGlobalCMonitorAlreadySet = errors.New("global C monitor already set")
)

// HotkeyEvent holds a hotkey event.
type HotkeyEvent struct {
	HkID hotkey.ID
	IsOn bool
	T    time.Time
}

// Monitor holds a hotkey monitor.
type Monitor struct {
	Hotkeys hotkey.Registrar
	eventCh chan HotkeyEvent
	doneCh  chan struct{}
	engine  Engine
	logCb   log.Callback
	mu      sync.Mutex
}

// New constructs a new monitor.
func New(hkReg hotkey.Registrar, engine Engine) *Monitor {
	if engine == nil {
		engine = CEngine{}
	}
	return &Monitor{
		Hotkeys: hkReg,
		engine:  engine,
	}
}

// SetLogCb sets the monitor lgo callback.
func (m *Monitor) SetLogCb(logCb log.Callback) {
	m.logCb = logCb
}

// Start starts hotkey monitoring.
func (m *Monitor) Start() (
	hotkeyCh chan HotkeyEvent,
	doneCh chan struct{},
	err error,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.eventCh != nil {
		return nil, nil, ErrAlreadyStarted
	}

	if ok := m.engine.Init(); !ok {
		return nil, nil, ErrInitFailed
	}

	err = m.engine.Start(m)
	if err != nil {
		defer m.engine.Deinit()
		return nil, nil, err
	}

	m.doneCh = make(chan struct{})
	m.eventCh = make(chan HotkeyEvent)

	if m.logCb != nil {
		m.logCb("Hotkey monitor started.")
	}

	return m.eventCh, m.doneCh, nil
}

// Stop stops hotkey monitoring.
func (m *Monitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.eventCh == nil {
		return ErrNotYetStarted
	}

	m.engine.Stop()

	if ok := m.engine.Deinit(); !ok {
		return ErrDeinitFailed
	}

	go func(doneCh chan struct{}) {
		doneCh <- struct{}{}
		close(doneCh)
	}(m.doneCh)
	m.doneCh = nil
	close(m.eventCh)
	m.eventCh = nil

	if m.logCb != nil {
		m.logCb("Hotkey monitor stopped.")
	}

	return nil
}

// Dispatch dispatches a hotkey even through the monitor.
func (m *Monitor) Dispatch(event HotkeyEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.eventCh == nil {
		return ErrNotYetStarted
	}
	m.eventCh <- event
	return nil
}

// CEngine implements monitor engine via C.
type CEngine struct{}

// Init initializes the engine for monitoring.
func (CEngine) Init() (ok bool) {
	return bool(C.initMonitor())
}

// Start starts the engine for monitoring.
func (CEngine) Start(m *Monitor) error {
	err := setGlobalMonitor(m)
	if err != nil {
		return err
	}
	go C.startMonitor()
	return nil
}

// Stop stops the engine from monitoring.
func (CEngine) Stop() {
	C.stopMonitor()
	setGlobalMonitor(nil)
}

// Deinit deinitializes the engine for monitoring.
func (CEngine) Deinit() (ok bool) {
	return bool(C.deinitMonitor())
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

//export goHandleHotkeyEvent
func goHandleHotkeyEvent(cHotkeyID uint8, isOn bool) {
	globalCMonitorMx.Lock()
	defer globalCMonitorMx.Unlock()

	m := globalCMonitor

	if m == nil {
		panic(ErrGlobalCMonitorMissing)
	}

	hotkeyID := hotkey.ID(cHotkeyID)
	event := HotkeyEvent{hotkeyID, isOn, time.Now()}
	err := m.Dispatch(event)

	if err != nil {
		panic(ErrGlobalCMonitorMissing)
	}
}
