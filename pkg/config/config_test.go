package config_test

import (
	"testing"

	"github.com/echocrow/Mouser/pkg/config"
	"github.com/stretchr/testify/assert"
)

var ds = config.DefaultSettings

type Conf = config.Config

type ARef = config.ActionRef
type BasicA = config.BasicAction
type ToggleA = config.ToggleAction
type AppBrA = config.AppBranchAction
type ReqAppA = config.RequireAppAction

type Gests = config.GestureSeries

func TestParseYAML(t *testing.T) {
	tests := []struct {
		name   string
		yml    string
		config Conf
		wantOk bool
	}{
		{
			"empty config",
			``,
			Conf{Settings: ds},
			true,
		},
		{
			"simple mapping",
			`
      mappings:
        K1: fookey
        K2: barkey
      `,
			Conf{
				Mappings: map[config.KeyAlias]config.MappingKey{
					"K1": {Key: "fookey"},
					"K2": {Key: "barkey"},
				},
				Settings: ds,
			},
			true,
		},
		{
			"simple hotkeys",
			`
      gestures:
        K1:
          foo_down: foo:action
          bar_tap.bar_hold.bar_tap: bar:action
        K2:
          fizz_left: fizz:buzz
      `,
			Conf{
				Gestures: map[config.KeyAlias]config.GestureActions{
					"K1": {
						{
							Gesture: Gests{"foo_down"},
							Action:  ARef{BasicA{Name: "foo:action"}},
						},
						{
							Gesture: Gests{"bar_tap", "bar_hold", "bar_tap"},
							Action:  ARef{BasicA{Name: "bar:action"}},
						},
					},
					"K2": {
						{
							Gesture: Gests{"fizz_left"},
							Action:  ARef{BasicA{Name: "fizz:buzz"}},
						},
					},
				},
				Settings: ds,
			},
			true,
		},
		{
			"custom toggles",
			`
      actions:
        foo:0:toggle:
          type: toggle
          action: foo
        foo:1:toggle:
          type: toggle
          action: timed-foo
          init-delay: 123
          repeat-delay: 456
        fizz:buzz:toggle:
          type: toggle
          action:
            action: fizz:buzz
            args: [3, 5]
      `,
			Conf{
				Actions: map[string]ARef{
					"foo:0:toggle": {ToggleA{
						Action:      ARef{BasicA{Name: "foo"}},
						InitDelay:   ds.Toggles.InitDelay,
						RepeatDelay: ds.Toggles.RepeatDelay,
					}},
					"foo:1:toggle": {ToggleA{
						Action:      ARef{BasicA{Name: "timed-foo"}},
						InitDelay:   123,
						RepeatDelay: 456,
					}},
					"fizz:buzz:toggle": {ToggleA{
						Action: ARef{BasicA{
							Name: "fizz:buzz",
							Args: []interface{}{3, 5},
						}},
						InitDelay:   ds.Toggles.InitDelay,
						RepeatDelay: ds.Toggles.RepeatDelay,
					}},
				},
				Settings: ds,
			},
			true,
		},
		{
			"app branch",
			`
      actions:
        foo:special:
          type: app-branch
          branches:
            /App/Foo: foo:action
            /App/Bar: null
            /App/Baz:
              type: action
              action: baz:action
          fallback: fall:back:action
      `,
			Conf{
				Actions: map[string]ARef{
					"foo:special": {AppBrA{
						Branches: map[string]ARef{
							"/App/Foo": {BasicA{Name: "foo:action"}},
							"/App/Bar": {},
							"/App/Baz": {BasicA{Name: "baz:action"}},
						},
						Fallback: ARef{BasicA{Name: "fall:back:action"}},
					}},
				},
				Settings: ds,
			},
			true,
		},
		{
			"require app",
			`
      actions:
        foo:require-app:
          type: require-app
          app: /App/Foo
          do: foo:action
          fallback: fall:back:action
      `,
			Conf{
				Actions: map[string]ARef{
					"foo:require-app": {ReqAppA{
						App:      "/App/Foo",
						Do:       ARef{BasicA{Name: "foo:action"}},
						Fallback: ARef{BasicA{Name: "fall:back:action"}},
					}},
				},
				Settings: ds,
			},
			true,
		},
		{
			"simple settings",
			`
      settings:
        debug: true
        gestures:
          cap: 42
        swipes:
          min-dist: 123
      `,
			Conf{
				Settings: config.Settings{
					Debug: true,
					Gestures: config.GestureSettings{
						TTL:           ds.Gestures.TTL,
						ShortPressTTL: ds.Gestures.ShortPressTTL,
						Cap:           42,
					},
					Swipes: config.SwipeSettings{
						MinDist:  123,
						Throttle: ds.Swipes.Throttle,
						PollRate: ds.Swipes.PollRate,
					},
					Toggles: ds.Toggles,
				},
			},
			true,
		},
		{
			"complex config",
			`
      mappings:
        K1: fookey
        K2: barkey
        K3: {key: bazkey}

      gestures:
        K1:
          foo_down: foo:action
          bar_tap.bar_hold.bar_tap: bar:action
        fizzkey:
          fizz_left: fizz:buzz
        K2:
        K3:
          -
            gesture: bar0.bar1
            action: bar:some:action
          -
            gesture: [bar2, bar3]
            exact: true
            action: bar:another:action
        K9:
          foo:
            action: foo:action
            somekey: ignore_me
          bar:
            type: action
            action: bar:action
          fizz.buzz:
            action: fizz:buzz
            args: [1, 2, fizz, 4, buzz]
          baz:
            type: app-branch
            branches:
              /Far.app: far:far
            fallback: baz:baz

      actions:
        foo:0: foo:0
        foo:1: foo:1
        foo:2:
        foo:3:
          action: bar:3
        foo:4:
          type: action
          action: bar:4
        foo:5:
          action: bar:5
          random_key: ignore
        foo:6:
          action: bar:6
          args: [foo, 1, 2, bar, 3]
        foo:7:
          type: app-branch
          branches:
            /App/Foo: foo:action
            /App/Bar: null
            /App/Baz: baz:action
          fallback: fall:back:action
        foo:8:
          type: require-app
          app: Foo.app
          do:
            action: bar:8
          fallback:
            action: bar:8:fallback
            args: [foo, 1, 2, bar, 3]
        foo:9000:
          type: app-branch
          branches:
            Foo.app: foo:action
            "": bar:action
            Baz.app:
              type: action
              action: nested:action
              args: [inception]
            TooFar.app:
              type: app-branch
              branches:
                Bar.Foo.app: far:action
                Bar.Baz.app:
                  action: double:nested:action
                  args: [in, too, deep]
              fallback: nested:fallback
          fallback: null

      settings:
        gestures:
          ttl: 12
          short-press-ttl: 34
          cap: 56
        swipes:
          min-dist: 987
          throttle: 654
          poll-rate: 321
        toggles:
          init-delay: 111
          repeat-delay: 222
      `,
			Conf{
				Mappings: map[config.KeyAlias]config.MappingKey{
					"K1": {Key: "fookey"},
					"K2": {Key: "barkey"},
					"K3": {Key: "bazkey"},
				},
				Gestures: map[config.KeyAlias]config.GestureActions{
					"K1": {
						{
							Gesture: Gests{"foo_down"},
							Action:  ARef{BasicA{Name: "foo:action"}},
						},
						{
							Gesture: Gests{"bar_tap", "bar_hold", "bar_tap"},
							Action:  ARef{BasicA{Name: "bar:action"}},
						},
					},
					"fizzkey": {
						{
							Gesture: Gests{"fizz_left"},
							Action:  ARef{BasicA{Name: "fizz:buzz"}},
						},
					},
					"K2": nil,
					"K3": {
						{
							Gesture: Gests{"bar0", "bar1"},
							Action:  ARef{BasicA{Name: "bar:some:action"}},
						},
						{
							Gesture: Gests{"bar2", "bar3"},
							Exact:   true,
							Action:  ARef{BasicA{Name: "bar:another:action"}},
						},
					},
					"K9": {
						{
							Gesture: Gests{"foo"},
							Action:  ARef{BasicA{Name: "foo:action"}},
						},
						{
							Gesture: Gests{"bar"},
							Action:  ARef{BasicA{Name: "bar:action"}},
						},
						{
							Gesture: Gests{"fizz", "buzz"},
							Action: ARef{BasicA{
								Name: "fizz:buzz",
								Args: []interface{}{1, 2, "fizz", 4, "buzz"},
							}},
						},
						{
							Gesture: Gests{"baz"},
							Action: ARef{AppBrA{
								Branches: map[string]ARef{
									"/Far.app": {BasicA{Name: "far:far"}},
								},
								Fallback: ARef{BasicA{Name: "baz:baz"}},
							}},
						},
					},
				},
				Actions: map[string]ARef{
					"foo:0": {BasicA{Name: "foo:0"}},
					"foo:1": {BasicA{Name: "foo:1"}},
					"foo:2": {},
					"foo:3": {BasicA{Name: "bar:3"}},
					"foo:4": {BasicA{Name: "bar:4"}},
					"foo:5": {BasicA{Name: "bar:5"}},
					"foo:6": {BasicA{
						Name: "bar:6",
						Args: []interface{}{"foo", 1, 2, "bar", 3},
					}},
					"foo:7": {AppBrA{
						Branches: map[string]ARef{
							"/App/Foo": {BasicA{Name: "foo:action"}},
							"/App/Bar": {},
							"/App/Baz": {BasicA{Name: "baz:action"}},
						},
						Fallback: ARef{BasicA{Name: "fall:back:action"}},
					}},
					"foo:8": {ReqAppA{
						App: "Foo.app",
						Do:  ARef{BasicA{Name: "bar:8"}},
						Fallback: ARef{BasicA{
							Name: "bar:8:fallback",
							Args: []interface{}{"foo", 1, 2, "bar", 3},
						}},
					}},
					"foo:9000": {AppBrA{
						Branches: map[string]ARef{
							"Foo.app": {BasicA{Name: "foo:action"}},
							"":        {BasicA{Name: "bar:action"}},
							"Baz.app": {BasicA{
								Name: "nested:action",
								Args: []interface{}{"inception"},
							}},
							"TooFar.app": {AppBrA{
								Branches: map[string]ARef{
									"Bar.Foo.app": {BasicA{Name: "far:action"}},
									"Bar.Baz.app": {BasicA{
										Name: "double:nested:action",
										Args: []interface{}{"in", "too", "deep"},
									}},
								},
								Fallback: ARef{BasicA{Name: "nested:fallback"}},
							}},
						},
						Fallback: ARef{},
					}},
				},
				Settings: config.Settings{
					Debug: false,
					Gestures: config.GestureSettings{
						TTL:           12,
						ShortPressTTL: 34,
						Cap:           56,
					},
					Swipes: config.SwipeSettings{
						MinDist:  987,
						Throttle: 654,
						PollRate: 321,
					},
					Toggles: config.ToggleSettings{
						InitDelay:   111,
						RepeatDelay: 222,
					},
				},
			},
			true,
		},
		{
			"empty gesture sequence 1",
			`
      gestures:
        K1:
          "": foo:action
      `,
			Conf{},
			false,
		},
		{
			"empty gesture sequence 2",
			`
      gestures:
        K1:
          - gesture: []
            action: foo:action
      `,
			Conf{},
			false,
		},
		{
			"empty gesture",
			`
      gestures:
        K1:
          "foo.bar.": foo:action
      `,
			Conf{},
			false,
		},
		{
			"invalid action name",
			`
      actions:
        foo:action: [123]
      `,
			Conf{},
			false,
		},
		{
			"invalid action type",
			`
      actions:
        foo:action:
          type: invalid_type
          action: bar:action
      `,
			Conf{},
			false,
		},
		{
			"invalid action args",
			`
      actions:
        foo:action:
          action: bar:action
          args: not_a_list
      `,
			Conf{},
			false,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := config.ParseYAML([]byte(tc.yml))
			if tc.wantOk {
				assert.Equal(t, tc.config, c)
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
