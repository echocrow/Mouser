// Package monitor allows listening to and dispatching hotkey events.
package monitor

import (
	"errors"
	"sync"
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/birdkid/mouser/pkg/log"
)

// Monitor errors raised by package monitor.
var (
	ErrInitFailed     = errors.New("monitor initialization failed")
	ErrDeinitFailed   = errors.New("monitor de-initialization failed")
	ErrAlreadyStarted = errors.New("monitor already started")
	ErrNotYetStarted  = errors.New("monitor not yet started")
)

// HotkeyEvent holds a hotkey event.
type HotkeyEvent struct {
	HkID hotkey.ID
	IsOn bool
	T    time.Time
}

// Engine describes a hotkeys monitor engine.
//go:generate mockery --name "Engine"
type Engine interface {
	Init() (ok bool)
	Start(m *Monitor) error
	Stop()
	Deinit() (ok bool)
}

// Monitor holds a hotkey monitor.
type Monitor struct {
	Hotkeys hotkey.Registrar
	eventCh chan HotkeyEvent
	engine  Engine
	logCb   log.Callback
	mu      sync.Mutex
}

// New constructs a new monitor.
func New(hkReg hotkey.Registrar, engine Engine) *Monitor {
	if engine == nil {
		engine = defaultEngine()
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
	err error,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.eventCh != nil {
		return nil, ErrAlreadyStarted
	}

	if ok := m.engine.Init(); !ok {
		return nil, ErrInitFailed
	}

	err = m.engine.Start(m)
	if err != nil {
		defer m.engine.Deinit()
		return nil, err
	}

	m.eventCh = make(chan HotkeyEvent)

	if m.logCb != nil {
		m.logCb("Started")
	}

	return m.eventCh, nil
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

	close(m.eventCh)
	m.eventCh = nil

	if m.logCb != nil {
		m.logCb("Stopped")
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
