package gestures_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/echocrow/Mouser/pkg/hotkeys/gestures"
	"github.com/echocrow/Mouser/pkg/hotkeys/gestures/swipes"
	swpMocks "github.com/echocrow/Mouser/pkg/hotkeys/gestures/swipes/mocks"
	"github.com/echocrow/Mouser/pkg/hotkeys/hotkey"
	"github.com/echocrow/Mouser/pkg/hotkeys/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	kDown  = gestures.KeyDown
	kUp    = gestures.KeyUp
	pShort = gestures.PressShort
	pLong  = gestures.PressLong
	sUp    = gestures.SwipeUp
	sDown  = gestures.SwipeDown
	sLeft  = gestures.SwipeLeft
	sRight = gestures.SwipeRight
)

type gst = gestures.Gesture

const (
	shortPressTTL = 100
	longPressT    = shortPressTTL + 1
	gestureTTL    = 100
	gesturesCap   = 4
)

func newConfig() gestures.Config {
	return gestures.Config{
		ShortPressTTL: time.Second * shortPressTTL,
		GestureTTL:    time.Second * gestureTTL,
		Cap:           gesturesCap,
	}
}

type swpEvSetter func(swipes.Event)

func newMockSwipesMonitor(
	ch <-chan swipes.Event,
) (m *swpMocks.Monitor, setPauseEv swpEvSetter) {
	m = new(swpMocks.Monitor)
	m.On("Init").Return(ch)
	m.On("Restart").Return()
	var pauseEv swipes.Event
	setPauseEv = func(ev swipes.Event) { pauseEv = ev }
	m.On("Pause", mock.AnythingOfType("time.Time")).Return(
		func(_ time.Time) swipes.Event { return pauseEv },
	)
	m.On("Stop").Return()
	return m, setPauseEv
}

const (
	sdNil   = swipes.NoSwipe
	sdRight = swipes.SwipeRight
	sdUp    = swipes.SwipeUp
	sdLeft  = swipes.SwipeLeft
	sdDown  = swipes.SwipeDown
)

func TestFromHotkeys(t *testing.T) {
	t.Parallel()

	config := newConfig()

	tests := []struct {
		name      string
		hkGestEvs []hkGestEv
	}{
		{
			"detects nothing",
			[]hkGestEv{},
		},
		{
			"detects short press",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort}},
				{1, 2, hk{true}, nil},
				{1, 2, hk{}, []gst{pShort}},
				{1, 3, hk{true}, nil},
				{1, 3, hk{}, []gst{pShort}},
			},
		},
		{
			"detects long press",
			[]hkGestEv{
				{1, 123, hk{true}, nil},
				{longPressT, 123, hk{}, []gst{pLong}},
			},
		},
		{
			"differentiates hotkeys",
			[]hkGestEv{
				{1, 11, hk{true}, []gst{}},
				{1, 11, hk{}, []gst{pShort}},
				{1, 13, hk{true}, nil},
				{1, 13, hk{}, []gst{pShort}},
				{1, 17, hk{true}, nil},
				{1, 17, hk{}, []gst{pShort}},
			},
		},
		{
			"chains gestures",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort}},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort, pShort}},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort, pShort, pShort}},

				{1, 2, hk{true}, nil},
				{1, 2, hk{}, []gst{pShort}},
				{1, 2, hk{true}, nil},
				{longPressT, 2, hk{}, []gst{pShort, pLong}},
				{1, 2, hk{true}, nil},
				{1, 2, hk{}, []gst{pShort, pLong, pShort}},
			},
		},
		{
			"handles repeated signals",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, hk{true}, nil},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort}},
				{1, 1, hk{}, nil},
				{longPressT, 1, hk{}, nil},
			},
		},
		{
			"discards prior hotkeys on multi-press",
			[]hkGestEv{
				{1, 11, hk{true}, nil},
				{1, 13, hk{true}, nil},
				{1, 17, hk{true}, nil},
				{1, 11, hk{}, nil},
				{1, 17, hk{}, []gst{pShort}},
				{1, 13, hk{}, nil},
			},
		},
		{
			"honors gesture cap",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort}},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort, pShort}},
				{1, 1, hk{true}, nil},
				{longPressT, 1, hk{}, []gst{pShort, pShort, pLong}},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort, pShort, pLong, pShort}},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort, pLong, pShort, pShort}},
				{1, 1, hk{true}, nil},
				{longPressT, 1, hk{}, []gst{pLong, pShort, pShort, pLong}},
				{1, 1, hk{true}, nil},
				{longPressT, 1, hk{}, []gst{pShort, pShort, pLong, pLong}},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort, pLong, pLong, pShort}},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evs, want, gestEvsLen := newHkGestEvs(tc.hkGestEvs)

			hkEvC := make(chan monitor.HotkeyEvent)
			gestEvC := gestures.FromHotkeysCustom(hkEvC, config, nil)

			got := make([]gestures.Event, 0, gestEvsLen)
			got = sendEvs(evs, hkEvC, swpSender{}, gestEvC, got)

			assert.Equal(t, want, got)
		})
	}
}

func TestFromHotkeysSwipesMonitorStartStop(t *testing.T) {
	t.Parallel()

	config := newConfig()

	tests := []struct {
		name      string
		hkGestEvs []hkGestEv
		restarts  int
		pauses    int
		stops     int
	}{
		{
			"does not start",
			[]hkGestEv{},
			0, 0, 1,
		},
		{
			"starts with down-presses",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, nil},
				{1, 2, hk{true}, nil},
				{1, 2, hk{}, nil},
				{1, 3, hk{true}, nil},
				{1, 3, hk{}, nil},
			},
			3, 3, 1,
		},
		{
			"restarts with conflicting down-presses",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, nil},
				{1, 2, hk{true}, nil},
				{1, 3, hk{true}, nil},
				{1, 2, hk{}, nil},
				{1, 3, hk{}, nil},
			},
			3, 2, 1,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evs, _, _ := newHkGestEvs(tc.hkGestEvs)

			swpEvs := make(chan swipes.Event)
			swpMon, _ := newMockSwipesMonitor(swpEvs)
			defer close(swpEvs)

			hkEvC := make(chan monitor.HotkeyEvent)
			gestEvC := gestures.FromHotkeysCustom(hkEvC, config, swpMon)

			sendEvs(evs, hkEvC, swpSender{}, gestEvC, nil)

			swpMon.AssertNumberOfCalls(t, "Restart", tc.restarts)
			swpMon.AssertNumberOfCalls(t, "Pause", tc.pauses)
			swpMon.AssertNumberOfCalls(t, "Stop", tc.stops)
		})
	}
}

func TestFromHotkeysSwipes(t *testing.T) {
	t.Parallel()

	config := newConfig()

	tests := []struct {
		name      string
		hkGestEvs []hkGestEv
	}{
		{
			"detects nothing",
			[]hkGestEv{},
		},
		{
			"discards void getures",
			[]hkGestEv{
				{1, 1, swp{sdLeft}, nil},
				{1, 1, swp{sdRight}, nil},
				{1, 1, hk{true}, nil},
				{1, 1, hk{}, []gst{pShort}},
				{1, 1, swp{sdUp}, nil},
				{1, 1, swp{sdDown}, nil},
			},
		},
		{
			"detects simple swipes",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, swp{sdLeft}, []gst{sLeft}},
				{1, 1, hk{}, nil},

				{1, 2, hk{true}, nil},
				{1, 2, swp{sdUp}, []gst{sUp}},
				{1, 2, hk{}, nil},
			},
		},
		{
			"chains gestures",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, swp{sdLeft}, []gst{sLeft}},
				{1, 1, swp{sdRight}, []gst{sLeft, sRight}},
				{1, 1, swp{sdUp}, []gst{sLeft, sRight, sUp}},
				{1, 1, hk{}, nil},

				{1, 2, hk{true}, nil},
				{1, 2, hk{}, []gst{pShort}},
				{1, 2, hk{true}, nil},
				{1, 2, swp{sdDown}, []gst{pShort, sDown}},
				{1, 2, swp{sdRight}, []gst{pShort, sDown, sRight}},
				{1, 2, hk{}, nil},
				{1, 2, hk{true}, nil},
				{longPressT, 2, hk{}, []gst{pShort, sDown, sRight, pLong}},

				{1, 3, hk{true}, nil},
				{1, 3, swp{sdUp}, []gst{sUp}},
				{1, 3, swp{sdRight}, []gst{sUp, sRight}},
				{1, 3, hk{}, nil},
				{1, 3, hk{true}, nil},
				{1, 3, hk{}, []gst{sUp, sRight, pShort}},
				{1, 3, hk{true}, nil},
				{1, 3, swp{sdDown}, []gst{sUp, sRight, pShort, sDown}},
				{1, 3, hk{}, nil},

				{1, 4, hk{true}, nil},
				{1, 4, swp{sdLeft}, []gst{sLeft}},
				{1, 4, swp{sdLeft}, []gst{sLeft, sLeft}},
				{1, 4, swp{sdLeft}, []gst{sLeft, sLeft, sLeft}},
				{1, 4, swp{sdLeft}, []gst{sLeft, sLeft, sLeft, sLeft}},
				{1, 4, hk{}, nil},
			},
		},
		{
			"honors release swipes",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, releaseSwp{sdLeft}, []gst{sLeft}},

				{1, 2, hk{true}, nil},
				{longPressT, 2, releaseSwp{sdDown}, []gst{sDown}},
			},
		},
		{
			"discards prior hotkeys on multi-press",
			[]hkGestEv{
				{1, 5, hk{true}, nil},
				{1, 7, hk{true}, nil},
				{1, 9, hk{true}, nil},
				{1, 9, swp{sdDown}, []gst{sDown}},
				{1, 5, hk{}, nil},
				{1, 9, swp{sdUp}, []gst{sDown, sUp}},
				{1, 9, hk{}, nil},
				{1, 9, swp{sdLeft}, nil},
				{1, 7, hk{}, nil},
				{1, 9, swp{sdRight}, nil},
			},
		},
		{
			"honors gesture cap",
			[]hkGestEv{
				{1, 1, hk{true}, nil},
				{1, 1, swp{sdLeft}, []gst{sLeft}},
				{1, 1, swp{sdDown}, []gst{sLeft, sDown}},
				{1, 1, swp{sdRight}, []gst{sLeft, sDown, sRight}},
				{1, 1, swp{sdUp}, []gst{sLeft, sDown, sRight, sUp}},
				{1, 1, swp{sdLeft}, []gst{sDown, sRight, sUp, sLeft}},
				{1, 1, swp{sdDown}, []gst{sRight, sUp, sLeft, sDown}},
				{1, 1, swp{sdRight}, []gst{sUp, sLeft, sDown, sRight}},
				{1, 1, swp{sdUp}, []gst{sLeft, sDown, sRight, sUp}},
				{1, 1, hk{}, nil},

				{1, 2, hk{true}, nil},
				{1, 2, swp{sdLeft}, []gst{sLeft}},
				{1, 2, swp{sdLeft}, []gst{sLeft, sLeft}},
				{1, 2, swp{sdRight}, []gst{sLeft, sLeft, sRight}},
				{1, 2, swp{sdUp}, []gst{sLeft, sLeft, sRight, sUp}},
				{1, 2, swp{sdRight}, []gst{sLeft, sRight, sUp, sRight}},
				{1, 2, swp{sdDown}, []gst{sRight, sUp, sRight, sDown}},
				{1, 2, releaseSwp{sdDown}, []gst{sUp, sRight, sDown, sDown}},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evs, want, gestEvsLen := newHkGestEvs(tc.hkGestEvs)

			swpEvs := make(chan swipes.Event)
			swpMon, setSwpPauseEv := newMockSwipesMonitor(swpEvs)
			defer close(swpEvs)

			hkEvC := make(chan monitor.HotkeyEvent)
			gestEvC := gestures.FromHotkeysCustom(hkEvC, config, swpMon)

			got := make([]gestures.Event, 0, gestEvsLen)
			got = sendEvs(evs, hkEvC, swpSender{swpEvs, setSwpPauseEv}, gestEvC, got)

			assert.Equal(t, want, got)
		})
	}
}

type hk struct {
	isOn bool
}

type swp struct {
	dir swipes.Dir
}

type releaseSwp struct {
	dir swipes.Dir
}

type releaseSwpEv struct {
	swpEv swipes.Event
	hkEv  monitor.HotkeyEvent
}

type hkGestEv struct {
	dt        uint
	hkID      uint
	action    interface{}
	gestsList []gst
}

func newHkGestEvs(rawEvs []hkGestEv) (
	evs []interface{},
	gestEvs []gestures.Event,
	gestEvsLen uint,
) {
	for _, rawEv := range rawEvs {
		action := rawEv.action
		switch action.(type) {
		case hk, releaseSwp:
			gestEvsLen++
		}
		if len(rawEv.gestsList) > 0 {
			gestEvsLen++
		}
	}

	evs = make([]interface{}, len(rawEvs))
	gestEvs = make([]gestures.Event, gestEvsLen)

	t := time.Time{}
	evI := 0
	gstI := 0
	hkID := hotkey.ID(0)
	for _, rawEv := range rawEvs {
		t = t.Add(time.Second * time.Duration(rawEv.dt))
		action := rawEv.action
		hkID = hotkey.ID(rawEv.hkID)

		switch a := action.(type) {
		case hk:
			evs[evI] = monitor.HotkeyEvent{
				HkID: hkID,
				IsOn: a.isOn,
				T:    t,
			}
		case swp:
			evs[evI] = swipes.Event{
				Dir: a.dir,
				T:   t,
			}
		case releaseSwp:
			evs[evI] = releaseSwpEv{
				swpEv: swipes.Event{Dir: a.dir, T: t},
				hkEv:  monitor.HotkeyEvent{HkID: hkID, IsOn: false, T: t},
			}
		default:
			panic("Unexpected action type")
		}
		evI++

		// Add simple down or up event.
		switch a := action.(type) {
		case hk:
			gestEvs[gstI] = gestures.Event{
				HkID:  hkID,
				Gests: []gst{keyPressGest(a.isOn)},
				T:     t,
			}
			gstI++
		case releaseSwp:
			gestEvs[gstI] = gestures.Event{
				HkID:  hkID,
				Gests: []gst{keyPressGest(false)},
				T:     t,
			}
			gstI++
		}

		// Add additional desired gesture events.
		if len(rawEv.gestsList) > 0 {
			gestEvs[gstI] = gestures.Event{
				HkID:  hkID,
				Gests: rawEv.gestsList,
				T:     t,
			}
			gstI++
		}
	}
	return
}

func keyPressGest(isOn bool) gst {
	if isOn {
		return kDown
	}
	return kUp
}

type swpSender struct {
	c          chan<- swipes.Event
	setPauseEv swpEvSetter
}

func sendEvs(
	evs []interface{},
	hkEvs chan<- monitor.HotkeyEvent,
	swpSender swpSender,
	gestEvC <-chan gestures.Event,
	got []gestures.Event,
) []gestures.Event {
	received := make(chan struct{})
	defer close(received)
	go func() {
		for gestEv := range gestEvC {
			if got != nil {
				got = append(got, gestEv)
			}
		}
		received <- struct{}{}
	}()

	for _, ev := range evs {
		switch e := ev.(type) {
		case monitor.HotkeyEvent:
			hkEvs <- e
		case swipes.Event:
			swpSender.c <- e
		case releaseSwpEv:
			swpSender.setPauseEv(e.swpEv)
			hkEvs <- e.hkEv
		default:
			panic("Unexpected event type")
		}
	}
	close(hkEvs)

	<-received

	return got
}

func TestMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a    []gst
		b    []gst
		want bool
	}{
		{[]gst{}, []gst{}, true},
		{[]gst{kDown}, []gst{}, false},
		{[]gst{kDown}, []gst{kDown}, true},
		{[]gst{pLong}, []gst{pShort}, false},
		{[]gst{pShort, pLong}, []gst{pShort, pLong}, true},
		{[]gst{pShort, pLong}, []gst{pLong, pShort}, false},
		{[]gst{pShort, pLong, pLong}, []gst{pLong, pShort}, false},
		{[]gst{pShort, pLong, pLong}, []gst{pShort, pLong, pShort}, false},
		{[]gst{pShort, pLong, pLong}, []gst{pShort, pLong, pLong}, true},
	}
	for i, tc := range tests {
		tc := tc
		revs := [2]bool{false, true}
		for _, rev := range revs {
			rev := rev
			tn := fmt.Sprint(i)
			if rev {
				tn = tn + " (rev)"
			}
			t.Run(tn, func(t *testing.T) {
				t.Parallel()
				a, b := tc.a, tc.b
				if rev {
					a, b = b, a
				}
				got := gestures.Match(a, b)
				assert.Equal(t, tc.want, got)
			})
		}
	}
}

func TestMatchSingle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a    []gst
		g    gst
		want bool
	}{
		{[]gst{}, "", false},
		{[]gst{}, kDown, false},
		{[]gst{kDown}, kDown, true},
		{[]gst{kUp}, kUp, true},
		{[]gst{pShort}, pShort, true},
		{[]gst{kUp}, kDown, false},
		{[]gst{kDown}, kUp, false},
		{[]gst{kUp, kDown}, kDown, false},
		{[]gst{kDown, kUp}, kDown, false},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			a, g := tc.a, tc.g
			got := gestures.MatchSingle(a, g)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEndsIn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a    []gst
		g    gst
		want bool
	}{
		{[]gst{}, kDown, false},
		{[]gst{kDown}, kDown, true},
		{[]gst{kUp}, kDown, false},
		{[]gst{kUp, kDown}, kDown, true},
		{[]gst{kDown, kUp}, kDown, false},
		{[]gst{pShort, pShort, pLong}, pShort, false},
		{[]gst{pShort, pShort, pShort}, pShort, true},
		{[]gst{pLong, pLong, pLong, pShort}, pShort, true},
		{[]gst{pShort, pShort, pShort, pLong}, pShort, false},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			a, g := tc.a, tc.g
			got := gestures.EndsIn(a, g)
			assert.Equal(t, tc.want, got)
		})
	}
}
