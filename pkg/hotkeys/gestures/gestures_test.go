package gestures_test

import (
	"testing"
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/gestures"
	"github.com/birdkid/mouser/pkg/hotkeys/gestures/swipes"
	swpMocks "github.com/birdkid/mouser/pkg/hotkeys/gestures/swipes/mocks"
	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/birdkid/mouser/pkg/hotkeys/monitor"
	"github.com/stretchr/testify/assert"
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
)

func newConfig() gestures.Config {
	return gestures.Config{
		ShortPressTTL: time.Second * shortPressTTL,
		GestureTTL:    time.Second * gestureTTL,
	}
}

func newMockSwipesMonitor(
	ch <-chan swipes.Event,
) *swpMocks.Monitor {
	m := new(swpMocks.Monitor)
	m.On("Ch").Return(ch)
	m.On("Restart").Return()
	m.On("Pause").Return()
	m.On("Stop").Return()
	return m
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
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evs, want, gestEvsLen := newHkGestEvs(tc.hkGestEvs)

			hkEvC := make(chan monitor.HotkeyEvent)
			gestEvC := gestures.FromHotkeysCustom(hkEvC, config, nil)

			got := make([]gestures.Event, 0, gestEvsLen)
			got = sendEvs(evs, hkEvC, nil, gestEvC, got)

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
			swpMon := newMockSwipesMonitor(swpEvs)
			defer close(swpEvs)

			hkEvC := make(chan monitor.HotkeyEvent)
			gestEvC := gestures.FromHotkeysCustom(hkEvC, config, swpMon)

			sendEvs(evs, hkEvC, nil, gestEvC, nil)

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
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			evs, want, gestEvsLen := newHkGestEvs(tc.hkGestEvs)

			swpEvs := make(chan swipes.Event)
			swpMon := newMockSwipesMonitor(swpEvs)
			defer close(swpEvs)

			hkEvC := make(chan monitor.HotkeyEvent)
			gestEvC := gestures.FromHotkeysCustom(hkEvC, config, swpMon)

			got := make([]gestures.Event, 0, gestEvsLen)
			got = sendEvs(evs, hkEvC, swpEvs, gestEvC, got)

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
		if _, ok := action.(hk); ok {
			gestEvsLen = gestEvsLen + 1
		}
		if len(rawEv.gestsList) > 0 {
			gestEvsLen = gestEvsLen + 1
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
		default:
			panic("Unexpected action type")
		}
		evI++

		// Add simple down or up event.
		if hk, ok := action.(hk); ok {
			gestEvs[gstI] = gestures.Event{
				HkID:  hkID,
				Gests: []gst{keyPressGest(hk.isOn)},
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

func sendEvs(
	evs []interface{},
	hkEvs chan<- monitor.HotkeyEvent,
	swpEvs chan<- swipes.Event,
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
			swpEvs <- e
		}
	}
	close(hkEvs)

	<-received

	return got
}
