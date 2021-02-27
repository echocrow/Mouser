// Package bootstrap kickstarts mouser.
package bootstrap

import (
	"errors"

	"github.com/birdkid/mouser/pkg/config"
	"github.com/birdkid/mouser/pkg/hotkeys"
	"github.com/birdkid/mouser/pkg/hotkeys/gestures"
	"github.com/birdkid/mouser/pkg/hotkeys/gestures/swipes"
	"github.com/birdkid/mouser/pkg/hotkeys/hotkey"
	"github.com/birdkid/mouser/pkg/hotkeys/monitor"
)

// Bootstrap kickstarts mouser.
func Bootstrap(conf config.Config) (
	run func() error,
	stop func() error,
	err error,
) {
	m := hotkeys.DefaultMonitor(conf.Settings.Debug)

	hkGas, err := registerHotKeys(m, conf)
	if err != nil {
		return nil, nil, err
	}

	run = func() error {
		hkEvs, err := m.Start()
		if err != nil {
			return err
		}
		gestCh := gestures.FromHotkeysCustom(
			hkEvs,
			newGesturesConfig(conf.Settings.Gestures),
			swipes.NewCustomMonitor(newSwipesConfig(conf.Settings.Swipes)),
		)
		watchEvs(gestCh, hkGas)
		return nil
	}

	stop = m.Stop

	return run, stop, nil
}

func makeKey(alias config.KeyAlias, mapping config.Mapping) hotkey.KeyName {
	if mk, ok := mapping[alias]; ok {
		return hotkey.KeyName(mk.Key)
	}
	return hotkey.KeyName(alias)
}

func registerHotKeys(
	m *monitor.Monitor,
	conf config.Config,
) (map[hotkey.ID][]gestureAction, error) {
	if len(conf.HotKeys) == 0 {
		return nil, errors.New("no hotkeys specified")
	}

	actRepo := newActionsRepo(conf.Actions, conf.Settings)

	hkGas := make(map[hotkey.ID][]gestureAction, len(conf.HotKeys))
	for alias, gestActs := range conf.HotKeys {
		key := makeKey(alias, conf.Mappings)
		gas := make([]gestureAction, len(gestActs))

		for i, gac := range gestActs {
			ga, err := makeGestureAction(gac, actRepo)
			if err != nil {
				return nil, err
			}
			gas[i] = ga
		}

		hkID, err := m.Hotkeys.Add(key)
		if err != nil {
			return nil, err
		}
		hkGas[hkID] = gas
	}
	return hkGas, nil
}

func watchEvs(
	gestCh <-chan gestures.Event,
	hkGas map[hotkey.ID][]gestureAction,
) {
	for event := range gestCh {
		if gas, ok := hkGas[event.HkID]; ok {
			for _, ga := range gas {
				if ga.G.matches(event.Gests) {
					if ga.A != nil {
						ga.A()
					}
					break
				}
			}
		}
	}
}

func newGesturesConfig(gs config.GestureSettings) gestures.Config {
	return gestures.Config{
		ShortPressTTL: gs.ShortPressTTL.Duration(),
		GestureTTL:    gs.ShortPressTTL.Duration(),
		Cap:           int(gs.Cap),
	}
}

func newSwipesConfig(ss config.SwipeSettings) swipes.Config {
	return swipes.Config{
		MinDist:  float64(ss.MinDist),
		Throttle: ss.Throttle.Duration(),
		PollRate: ss.PollRate.Duration(),
	}
}
