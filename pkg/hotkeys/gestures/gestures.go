// Package gestures maps hotkey events to key & gesture events.
package gestures

import (
	"time"

	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/birdkid/mouser/pkg/hotkeys/monitor"
)

// Gesture identifies the type of gesture.
type Gesture string

// Gesture types -- Nil event.
const (
	NoGesture Gesture = ""
)

// Gesture types -- Basic key events.
const (
	KeyDown Gesture = "key_down"
	KeyUp   Gesture = "key_up"
)

// Gesture types -- Key gestures.
const (
	PressShort Gesture = "tap"
	PressLong  Gesture = "hold"
)

// Config defines gestures settings.
type Config struct {
	ShortPressTTL time.Duration
	GestureTTL    time.Duration
}

// Event represents a key/mouse gestures event.
type Event struct {
	HkID  hotkey.ID
	Gests []Gesture
	T     time.Time
}

// Default settings.
const (
	defaultShortPressTTL = time.Millisecond * 500
	defaultGestureTTL    = time.Millisecond * 500
)

// FromHotkeys maps hotkey events to key gestures.
func FromHotkeys(
	hkEvs <-chan monitor.HotkeyEvent,
) <-chan Event {
	config := Config{
		ShortPressTTL: defaultShortPressTTL,
		GestureTTL:    defaultGestureTTL,
	}
	return FromHotkeysCustom(hkEvs, config)
}

// FromHotkeysCustom maps hotkey events to key gestures with custom options.
//
// When hkEvs is depleted, the returned channel closed.
//
func FromHotkeysCustom(
	hkEvs <-chan monitor.HotkeyEvent,
	config Config,
) <-chan Event {
	ch := make(chan Event)
	go mapHkEvs(hkEvs, config, ch)
	return ch
}

func mapHkEvs(
	hkEvs <-chan monitor.HotkeyEvent,
	config Config,
	ch chan<- Event,
) {
	defer close(ch)
	var (
		prvT  time.Time
		hk    hotkey.ID
		prvHk hotkey.ID
		gests []Gesture
	)
	for hkEv := range hkEvs {
		ch <- Event{hkEv.HkID, []Gesture{keyGesture(hkEv)}, hkEv.T}

		t := hkEv.T
		dt := t.Sub(prvT)
		prvT = t

		if hkEv.IsOn {
			if hkEv.HkID != prvHk || dt > config.GestureTTL {
				gests = nil
			}
			hk = hkEv.HkID
			prvHk = 0
		} else {
			if hkEv.HkID == hk {
				hk = 0
				prvHk = hkEv.HkID
				if dt <= config.ShortPressTTL {
					gests = append(gests, PressShort)
				} else {
					gests = append(gests, PressLong)
				}
				ch <- Event{hkEv.HkID, gests, t}
			}
		}
	}
}

func keyGesture(hkEv monitor.HotkeyEvent) Gesture {
	if hkEv.IsOn {
		return KeyDown
	}
	return KeyUp
}
