// Package swipes allows listening to mouse swipes.
package swipes

import (
	"math"
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

// Event represents a swipe direction at a given time.
type Event struct {
	Dir Dir
	T   time.Time
}

// PointerEvent represents a pointer position at a given time.
type PointerEvent struct {
	Pos vec.Vec2D
	T   time.Time
}

// Monitor describes a swipes monitor.
//go:generate mockery --name "Monitor"
type Monitor interface {
	Ch() <-chan Event
	Restart()
	Pause()
	Stop()
}

// NewDefaultMonitor creates a new default swipes monitor.
func NewDefaultMonitor() Monitor {
	return NewPointerMonitor(nil)
}

// Default pointer monitor settings.
const (
	defaultMinDist  float64       = 30
	defaultThrottle time.Duration = time.Millisecond * 250
)

// PointerEngine describes a swipes monitor engine.
//go:generate mockery --name "PointerEngine"
type PointerEngine interface {
	GetPointerPos() vec.Vec2D
	Start(evs chan<- PointerEvent)
	Pause()
	Stop()
}

// PointerMonitor detects and shares swipes.
type PointerMonitor struct {
	MinDist float64
	ThrotD  time.Duration
	ch      chan Event
	ptEvs   chan PointerEvent
	pause   chan struct{}
	isOn    bool
	engine  PointerEngine
}

// NewPointerMonitor creates a new swipes pointer monitor.
func NewPointerMonitor(engine PointerEngine) *PointerMonitor {
	if engine == nil {
		engine = newRobotGoEngine()
	}
	ch := make(chan Event)
	return &PointerMonitor{
		MinDist: defaultMinDist,
		ThrotD:  defaultThrottle,
		ch:      ch,
		ptEvs:   make(chan PointerEvent),
		pause:   make(chan struct{}),
		engine:  engine,
	}
}

// Ch returns the swipe events channel.
func (m *PointerMonitor) Ch() <-chan Event {
	return m.ch
}

// Restart starts swipe monitoring (after first pausing previous monitoring
// when previously started).
func (m *PointerMonitor) Restart() {
	m.softPause()
	m.isOn = true
	go m.run()
	m.engine.Start(m.ptEvs)
}

// Pause pauses swipe monitoring.
func (m *PointerMonitor) Pause() {
	m.softPause()
}

// Stop stops swipe monitoring.
func (m *PointerMonitor) Stop() {
	m.softPause()
	m.engine.Stop()
	close(m.ptEvs)
	close(m.ch)
	close(m.pause)
}

func (m *PointerMonitor) softPause() {
	if m.isOn {
		m.pause <- struct{}{}
		m.engine.Pause()
	}
	m.isOn = false
}

func (m *PointerMonitor) run() {
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
			dir := dirFromVec2(dp, m.MinDist)
			if dir != NoSwipe {
				if dir != prEv.Dir || ptEv.T.Sub(prEv.T) > m.ThrotD {
					ev := Event{dir, ptEv.T}
					prEv = ev
					m.ch <- ev
				}
				origin = p
			}
		case <-m.pause:
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

// Default robotGoEngine settings.
const (
	defaultPollRate time.Duration = time.Millisecond * 100
)

type robotGoEngine struct {
	pollRate time.Duration
	ticker   *time.Ticker
	pause    chan struct{}
}

func newRobotGoEngine() *robotGoEngine {
	return &robotGoEngine{
		pollRate: defaultPollRate,
		ticker:   newStoppedTicker(),
		pause:    make(chan struct{}),
	}
}

func (*robotGoEngine) GetPointerPos() vec.Vec2D {
	x, y := robotgo.GetMousePos()
	return vec.Vec2D{X: float64(x), Y: float64(-y)}
}

func (e *robotGoEngine) Start(ch chan<- PointerEvent) {
	go e.watch(ch)
	e.ticker.Reset(e.pollRate)
}

func (e *robotGoEngine) Pause() {
	e.ticker.Stop()
	e.pause <- struct{}{}
}

func (e *robotGoEngine) Stop() {
	close(e.pause)
}

func (e *robotGoEngine) watch(ch chan<- PointerEvent) {
	for {
		select {
		case t := <-e.ticker.C:
			pos := e.GetPointerPos()
			ch <- PointerEvent{pos, t}
		case <-e.pause:
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
