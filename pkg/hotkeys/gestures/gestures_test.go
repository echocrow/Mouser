package gestures_test

import (
	"testing"
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/gestures"
	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/birdkid/mouser/pkg/hotkeys/monitor"
	"github.com/stretchr/testify/assert"
)

const (
	kDown  = gestures.KeyDown
	kUp    = gestures.KeyUp
	pShort = gestures.PressShort
	pLong  = gestures.PressLong
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
			gestEvC := gestures.FromHotkeysCustom(hkEvC, config)

			got := make([]gestures.Event, 0, gestEvsLen)
			got = sendEvs(evs, hkEvC, gestEvC, got)

			assert.Equal(t, want, got)
		})
	}
}

type hk struct {
	isOn bool
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
		}
	}
	close(hkEvs)

	<-received

	return got
}
