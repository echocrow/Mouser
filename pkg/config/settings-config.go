package config

import (
	"time"
)

// Settings contains custom config settings.
type Settings struct {
	Debug    bool
	Gestures GestureSettings
	Swipes   SwipeSettings
	Toggles  ToggleSettings
}

// GestureSettings contains custom gesture settings.
type GestureSettings struct {
	TTL           Ms
	ShortPressTTL Ms   `yaml:"short-press-ttl"`
	Cap           uint `yaml:"cap"`
}

// SwipeSettings contains custom swipe settings.
type SwipeSettings struct {
	MinDist  uint `yaml:"min-dist"`
	Throttle Ms   `yaml:"throttle"`
	PollRate Ms   `yaml:"poll-rate"`
}

// ToggleSettings contains custom toggle settings.
type ToggleSettings struct {
	InitDelay   Ms `yaml:"init-delay"`
	RepeatDelay Ms `yaml:"repeat-delay"`
}

// Ms represents a time duration in miliseconds.
type Ms uint

// Duration returns ms as time.Duration.
func (ms Ms) Duration() time.Duration {
	return time.Millisecond * time.Duration(ms)
}

// DefaultSettings contains all default settings.
var DefaultSettings = Settings{
	Debug: false,
	Gestures: GestureSettings{
		TTL:           500,
		ShortPressTTL: 500,
		Cap:           8,
	},
	Swipes: SwipeSettings{
		MinDist:  30,
		Throttle: 250,
		PollRate: 100,
	},
	Toggles: ToggleSettings{
		InitDelay:   200,
		RepeatDelay: 100,
	},
}
