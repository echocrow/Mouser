package swipes_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/gestures/swipes"
	"github.com/birdkid/mouser/pkg/hotkeys/gestures/swipes/mocks"
	"github.com/birdkid/mouser/pkg/vec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	sNil   = swipes.NoSwipe
	sRight = swipes.SwipeRight
	sUp    = swipes.SwipeUp
	sLeft  = swipes.SwipeLeft
	sDown  = swipes.SwipeDown
)

type sDir = swipes.Dir

func newMockPointerEngine(
	staticPos vec.Vec2D,
	evs <-chan swipes.PointerEvent,
	sent chan<- struct{},
) *mocks.PointerEngine {
	e := new(mocks.PointerEngine)
	e.On("GetPointerPos").Return(staticPos)
	pause := make(chan struct{})
	e.On("Start", mock.AnythingOfType("chan<- swipes.PointerEvent")).Run(
		func(args mock.Arguments) {
			go func() {
				ch := args.Get(0).(chan<- swipes.PointerEvent)
				for {
					select {
					case ev := <-evs:
						ch <- ev
						if sent != nil {
							sent <- struct{}{}
						}
					case <-pause:
						return
					}
				}
			}()
		},
	)
	e.On("Pause").Run(func(args mock.Arguments) {
		pause <- struct{}{}
	})
	e.On("Stop").Run(func(args mock.Arguments) {
		close(pause)
	})
	return e
}

func newMockPointerMonitor(e swipes.PointerEngine) *swipes.PointerMonitor {
	return swipes.NewPointerMonitor(e)
}

func TestRotateDir(t *testing.T) {
	tests := []struct {
		name  string
		dir   sDir
		rad   float64
		wants []sDir
	}{
		{
			"empty swipe",
			sNil,
			math.Pi / 2, []sDir{sNil},
		},
		{
			"quarter turn",
			sRight, math.Pi / 2, []sDir{sUp, sLeft, sDown, sRight},
		},
		{
			"counter-quarter turn",
			sRight, -math.Pi / 2, []sDir{sDown, sLeft, sUp, sRight},
		},
		{
			"horizontal reflection",
			sRight, math.Pi, []sDir{sLeft, sRight},
		},
		{
			"vertical reflection",
			sDown, math.Pi, []sDir{sUp, sDown},
		},
		{
			"empty turn 1",
			sUp, 0, []sDir{sUp},
		},
		{
			"empty turn 2",
			sRight, 0, []sDir{sRight},
		},
		{
			"empty turn 3",
			sDown, 0, []sDir{sDown},
		},
		{
			"empty turn 4",
			sLeft, 0, []sDir{sLeft},
		},
	}
	for _, tc := range tests {
		tc := tc
		for f := float64(-2); f <= 2; f++ {
			f := f
			tn := fmt.Sprintf("handles %s rotation (%dx overturn)", tc.name, int(f))
			t.Run(tn, func(t *testing.T) {
				rad := tc.rad + (math.Pi * 2 * f)
				got := tc.dir
				for _, want := range tc.wants {
					got = rotateDir(got, rad)
					assert.Equal(t, want, got)
				}
			})
		}
	}
}

func TestPointerMonitorStartStop(t *testing.T) {
	t.Parallel()
	for repeats := 0; repeats <= 10; repeats++ {
		repeats := repeats
		tn := fmt.Sprintf("gracefully restarts & pauses (x%d)", repeats)
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			e := newMockPointerEngine(vec.Vec2D{}, nil, nil)
			m := newMockPointerMonitor(e)
			for i := 0; i < repeats; i++ {
				m.Restart()
				m.Pause()
			}
			m.Stop()

			e.AssertNumberOfCalls(t, "Start", repeats)
			e.AssertNumberOfCalls(t, "Pause", repeats)
			e.AssertNumberOfCalls(t, "Stop", 1)
		})
	}
}

func TestPointerMonitorCh(t *testing.T) {
	t.Parallel()

	const (
		minDist = float64(12)
		throtD  = uint(12)
	)
	minDistPyth2b := math.Sqrt(minDist * minDist / 5)

	tests := []struct {
		name     string
		origin   vec.Vec2D
		ptSwpEvs []ptSwpEv
	}{
		{
			"detects nothing",
			vec.Vec2D{},
			[]ptSwpEv{},
		},
		{
			"detects no movement",
			vec.Vec2D{},
			[]ptSwpEv{
				{1, 0, 0, sNil},
				{1, 0, 0, sNil},
				{1, 0, 0, sNil},
				{1, 0, 0, sNil},
			},
		},
		{
			"detect simple swipes",
			vec.Vec2D{},
			[]ptSwpEv{
				{1, minDist, 0, sRight},
				{1, 0, minDist, sUp},
				{1, -minDist, 0, sLeft},
				{1, 0, -minDist, sDown},

				{1, minDist, (minDist / 2), sRight},
				{1, (minDist / 2), minDist, sUp},
				{1, -minDist, (minDist / 2), sLeft},
				{1, +(minDist / 2), -minDist, sDown},

				{1, +(minDist * 4), +(minDist * 2), sRight},
				{1, +(minDist * 2), +(minDist * 4), sUp},
				{1, -(minDist * 4), +(minDist * 2), sLeft},
				{1, +(minDist * 2), -(minDist * 4), sDown},
			},
		},
		{
			"ignores small movement",
			vec.Vec2D{},
			[]ptSwpEv{
				{1, minDist / 4, 0, sNil},
				{1, 0, minDist / 4, sNil},
				{1, -minDist / 4, 0, sNil},
				{1, 0, -minDist / 4, sNil},
			},
		},
		{
			"throttles duplicates",
			vec.Vec2D{},
			[]ptSwpEv{
				{1, minDist, 0, sRight},
				{throtD / 3, minDist, 0, sNil},
				{throtD / 3, minDist, 0, sNil},
				{throtD / 3, minDist, 0, sNil},
				{throtD / 3, minDist, 0, sRight},
			},
		},
		{
			"does not throttle after direction change",
			vec.Vec2D{},
			[]ptSwpEv{
				{throtD / 4, +minDist, 0, sRight},
				{throtD / 4, -minDist, 0, sLeft},
				{throtD / 4, +minDist, 0, sRight},
				{throtD / 4, -minDist, 0, sLeft},
				{throtD / 4, +minDist, 0, sRight},
				{throtD / 4, -minDist, 0, sLeft},
			},
		},
		{
			"detects slow swipes",
			vec.Vec2D{},
			[]ptSwpEv{
				{throtD * 10, minDist, 0, sRight},
				{throtD * 10, 0, minDist, sUp},
				{throtD * 10, -minDist, 0, sLeft},
				{throtD * 10, 0, -minDist, sDown},

				{throtD * 10, minDist, 0, sRight},
				{throtD * 10, minDist, 0, sRight},
				{throtD * 10, minDist, 0, sRight},
				{throtD * 10, minDist, 0, sRight},
			},
		},
		{
			"builds swipes and resets",
			vec.Vec2D{},
			[]ptSwpEv{
				{throtD, 0, minDist / 3, sNil},
				{throtD, 0, minDist / 3, sNil},
				{throtD, 0, minDist / 3, sUp},
				{throtD, 0, minDist / 3, sNil},
				{throtD, 0, minDist / 3, sNil},
				{throtD, 0, minDist / 3, sUp},
			},
		},
		{
			"discards throttled swipe direction",
			vec.Vec2D{},
			[]ptSwpEv{
				{0, minDist, 0, sRight},
				{throtD / 4, minDist * 10, 0, sNil},
				{throtD / 4, 0, minDist, sUp},
			},
		},
		{
			"detects progressive diagonal swipes",
			vec.Vec2D{},
			[]ptSwpEv{
				{throtD, 0, minDist * 1 / 2, sNil},
				{1, minDist * 2 / 3, 0, sNil},
				{1, 0, minDist * 1 / 2, sUp},

				{throtD, minDistPyth2b, 0, sNil},
				{1, minDistPyth2b, 0, sNil},
				{1, 0, minDistPyth2b, sRight},

				{throtD, 0, minDistPyth2b, sNil},
				{1, minDistPyth2b, 0, sNil},
				{1, minDistPyth2b, 0, sRight},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		for rot := uint(0); rot < 4; rot++ {
			rot := rot
			t.Run(fmt.Sprintf("%s (rot %d)", tc.name, rot), func(t *testing.T) {
				t.Parallel()

				evsCh := make(chan swipes.PointerEvent)
				defer close(evsCh)

				sent := make(chan struct{})
				defer close(sent)

				e := newMockPointerEngine(tc.origin, evsCh, sent)
				m := newMockPointerMonitor(e)
				m.MinDist = minDist
				m.ThrotD = time.Duration(throtD) * time.Second

				ptrEvs, wantSwpEvs, swpsL := newPtSwpEvs(tc.ptSwpEvs, rot)
				got := make([]swipes.Event, 0, swpsL)

				m.Restart()

				done := make(chan struct{})
				defer close(done)
				go func() {
					for ev := range m.Ch() {
						got = append(got, ev)
					}
					done <- struct{}{}
				}()

				for _, ev := range ptrEvs {
					evsCh <- ev
					<-sent
				}

				m.Stop()
				<-done

				assert.Equal(t, wantSwpEvs, got)
				extra := []swipes.Event{}
				select {
				case ev, ok := <-m.Ch():
					if ok {
						extra = append(extra, ev)
					}
				default:
				}
				assert.Equal(t, []swipes.Event{}, extra, "keeps no extra events")
			})
		}
	}
}

type ptSwpEv struct {
	dt     uint
	xd, yd float64
	swpD   sDir
}

func newPtSwpEvs(rawEvs []ptSwpEv, rot uint) (
	ptrEvs []swipes.PointerEvent,
	swpEvs []swipes.Event,
	swpsL uint,
) {
	rad := float64(rot) * math.Pi / 2

	for _, rawEv := range rawEvs {
		if rawEv.swpD != sNil {
			swpsL++
		}
	}
	ptrEvs = make([]swipes.PointerEvent, len(rawEvs))
	swpEvs = make([]swipes.Event, swpsL)

	t := time.Time{}
	x := float64(0)
	y := float64(0)
	ptrI := 0
	swpI := 0
	for _, rawEv := range rawEvs {
		t = t.Add(time.Second * time.Duration(rawEv.dt))

		x = x + rawEv.xd
		y = y + rawEv.yd
		v := vec.Vec2D{X: x, Y: y}
		ptrEvs[ptrI] = swipes.PointerEvent{
			Pos: v.Rot(rad),
			T:   t,
		}
		ptrI++

		if rawEv.swpD != sNil {
			swpEvs[swpI] = swipes.Event{
				Dir: rotateDir(rawEv.swpD, rad),
				T:   t,
			}
			swpI++
		}
	}
	return
}

// rotateDir changes swipe direction dir by radian rad.
func rotateDir(dir sDir, rad float64) sDir {
	if dir == sNil {
		return sNil
	}
	b := sDir(1)
	c := float64(4)
	d := float64(dir - b)
	d = d + rad/(math.Pi*2/c)
	d = math.Mod((math.Mod(d, c) + c), c)
	return sDir(d) + b
}
