// Package swipes allows listening to mouse swipes.
package swipes

import (
	"math"
	"sync"
	"time"

	"github.com/birdkid/mouser/pkg/vec"
	"github.com/go-vgo/robotgo"
)

// Dir denotes a swipe direction.
type Dir uint

// Swipe directions.
const (
	NoSwipe Dir = iota
	SwipeRight
	SwipeUp
	SwipeLeft
	SwipeDown
)

// Config defines swipes settings.
type Config struct {
	MinDist  float64
	Throttle time.Duration
	PollRate time.Duration
}

// Event represents a swipe direction at a given time.
type Event struct {
	Dir Dir
	T   time.Time
}

// IsSwipe checks whether Event ev denotes a swipe movement.
func (ev Event) IsSwipe() bool {
	return ev.Dir != NoSwipe
}

// PointerEvent represents a pointer position at a given time.
type PointerEvent struct {
	Pos vec.Vec2D
	T   time.Time
}

// Monitor describes a swipes monitor.
//go:generate mockery --name "Monitor"
type Monitor interface {
	Init() <-chan Event
	Restart()
	Pause()
	Stop()
}

// Default pointer monitor settings.
const (
	defaultMinDist  float64       = 30
	defaultThrottle time.Duration = time.Millisecond * 250
)

// Default robotGoEngine settings.
const (
	defaultPollRate time.Duration = time.Millisecond * 100
)

// NewDefaultMonitor creates a new default swipes monitor.
func NewDefaultMonitor() Monitor {
	config := Config{
		MinDist:  defaultMinDist,
		Throttle: defaultThrottle,
		PollRate: defaultPollRate,
	}
	return NewCustomMonitor(config)
}

// NewCustomMonitor creates a new default swipes monitor with custom config.
func NewCustomMonitor(config Config) Monitor {
	return NewPointerMonitor(config, nil)
}

// Monitor states.
type monitorState uint8

const (
	monitorOff monitorState = iota
	monitorReady
	monitorOn
)

// PointerEngine describes a swipes monitor engine.
//go:generate mockery --name "PointerEngine"
type PointerEngine interface {
	GetPointerPos() vec.Vec2D
	Init(ptEvs chan<- PointerEvent)
	Resume()
	Pause()
	Stop()
}

// PointerMonitor detects and shares swipes.
type PointerMonitor struct {
	cfg    Config
	ch     chan Event
	engine PointerEngine
	ptEvs  <-chan PointerEvent
	reset  chan struct{}
	stop   chan struct{}
	state  monitorState
	mx     sync.RWMutex
}

// NewPointerMonitor creates a new swipes pointer monitor.
func NewPointerMonitor(config Config, engine PointerEngine) *PointerMonitor {
	if engine == nil {
		engine = newRobotGoEngine(config)
	}
	return &PointerMonitor{
		cfg:    config,
		engine: engine,
		reset:  make(chan struct{}),
		stop:   make(chan struct{}),
	}
}

// Init initializes the swipe monitor.
func (m *PointerMonitor) Init() <-chan Event {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.state != monitorOff {
		return nil
	}
	m.ch = make(chan Event, 1)
	ptEvs := make(chan PointerEvent)
	m.ptEvs = ptEvs
	m.engine.Init(ptEvs)
	go m.watch()
	m.state = monitorReady
	return m.ch
}

// Restart starts swipe monitoring (after first pausing previous monitoring
// when previously started).
func (m *PointerMonitor) Restart() {
	if m.state < monitorReady {
		return
	}

	m.Pause()

	m.reset <- struct{}{}
	m.engine.Resume()

	m.mx.Lock()
	defer m.mx.Unlock()
	m.state = monitorOn
}

// Pause pauses swipe monitoring.
func (m *PointerMonitor) Pause() {
	if m.state < monitorReady {
		return
	}

	if m.state == monitorOn {
		m.engine.Pause()
	}

	m.mx.Lock()
	defer m.mx.Unlock()
	m.state = monitorReady
}

// Stop stops swipe monitoring.
func (m *PointerMonitor) Stop() {
	if m.state < monitorReady {
		return
	}

	m.mx.Lock()
	m.state = monitorOff
	m.mx.Unlock()

	m.engine.Stop()
	close(m.stop)
	close(m.reset)
	close(m.ch)
}

func (m *PointerMonitor) watch() {
	origin := m.engine.GetPointerPos()
	prEv := Event{}

	for {
		select {
		case ptEv, ok := <-m.ptEvs:
			if !ok {
				return
			}
			p := ptEv.Pos
			dp := p.Sub(origin)
			dir := dirFromVec2(dp, m.cfg.MinDist)
			if dir != NoSwipe {
				if dir != prEv.Dir || ptEv.T.Sub(prEv.T) > m.cfg.Throttle {
					ev := Event{dir, ptEv.T}
					prEv = ev
					m.mx.RLock()
					if m.state == monitorOn {
						m.ch <- ev
					}
					m.mx.RUnlock()
				}
				origin = p
			}
		case _, ok := <-m.reset:
			if !ok {
				return
			}
			origin = m.engine.GetPointerPos()
		case <-m.stop:
			return
		}
	}
}

func dirFromVec2(v vec.Vec2D, minDist float64) Dir {
	if v.Len() < minDist {
		return NoSwipe
	}
	x, y := v.X, v.Y
	ux, uy := math.Abs(x), math.Abs(y)
	if uy >= ux {
		if y >= 0 {
			return SwipeUp
		}
		return SwipeDown
	}
	if x >= 0 {
		return SwipeRight
	}
	return SwipeLeft
}

type robotGoEngine struct {
	pollRate time.Duration
	ticker   *time.Ticker
	stop     chan struct{}
	ptEvs    chan<- PointerEvent
}

func newRobotGoEngine(config Config) *robotGoEngine {
	return &robotGoEngine{
		pollRate: config.PollRate,
		ticker:   newStoppedTicker(),
		stop:     make(chan struct{}),
	}
}

func (*robotGoEngine) GetPointerPos() vec.Vec2D {
	x, y := robotgo.GetMousePos()
	return vec.Vec2D{X: float64(x), Y: float64(-y)}
}

func (e *robotGoEngine) Init(ptEvs chan<- PointerEvent) {
	e.ptEvs = ptEvs
	go e.watch()
}

func (e *robotGoEngine) Resume() {
	e.ticker.Reset(e.pollRate)
}

func (e *robotGoEngine) Pause() {
	e.ticker.Stop()
}

func (e *robotGoEngine) Stop() {
	defer close(e.stop)
	defer close(e.ptEvs)
	e.stop <- struct{}{}
	e.ptEvs = nil
}

func (e *robotGoEngine) watch() {
	for {
		select {
		case t := <-e.ticker.C:
			pos := e.GetPointerPos()
			e.ptEvs <- PointerEvent{pos, t}
		case <-e.stop:
			return
		}
	}
}

func newStoppedTicker() *time.Ticker {
	ticker := time.NewTicker(time.Second)
	ticker.Stop()
	select {
	case <-ticker.C:
	default:
	}
	return ticker
}
