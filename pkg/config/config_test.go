package config_test

import (
	"testing"

	"github.com/birdkid/mouser/pkg/config"
	"github.com/stretchr/testify/assert"
)

var ds = config.DefaultSettings

func TestParseYAML(t *testing.T) {
	tests := []struct {
		name   string
		yml    string
		config config.Config
		wantOk bool
	}{
		{
			"empty config",
			``,
			config.Config{Settings: ds},
			true,
		},
		{
			"simple mapping",
			`
mappings:
  K1: fookey
  K2: barkey
`,
			config.Config{
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
hotkeys:
  K1:
    foo_down: foo:action
    bar_tap.bar_hold.bar_tap: bar:action
  K2:
    fizz_left: fizz:buzz
`,
			config.Config{
				HotKeys: map[config.KeyAlias]config.GestureActions{
					"K1": {
						{
							Gesture: config.GestureSeries{"foo_down"},
							Action: config.ActionRef{config.BasicAction{
								Name: "foo:action",
							}},
						},
						{
							Gesture: config.GestureSeries{"bar_tap", "bar_hold", "bar_tap"},
							Action: config.ActionRef{config.BasicAction{
								Name: "bar:action",
							}},
						},
					},
					"K2": {
						{
							Gesture: config.GestureSeries{"fizz_left"},
							Action: config.ActionRef{config.BasicAction{
								Name: "fizz:buzz",
							}},
						},
					},
				},
				Settings: ds,
			},
			true,
		},
		{
			"simple actions",
			`
actions:
  foo:alias:0: bar:0
  foo:alias:1: bar:1
  foo:special:
    type: app-branch
    branches:
      /App/Foo: foo:action
      /App/Bar: null
      /App/Baz: baz:action
    fallback: fall:back:action
`,
			config.Config{
				Actions: map[string]config.ActionRef{
					"foo:alias:0": {config.BasicAction{
						Name: "bar:0",
					}},
					"foo:alias:1": {config.BasicAction{
						Name: "bar:1",
					}},
					"foo:special": {config.AppBranchAction{
						Branches: map[string]config.ActionRef{
							"/App/Foo": {config.BasicAction{
								Name: "foo:action",
							}},
							"/App/Bar": {},
							"/App/Baz": {config.BasicAction{
								Name: "baz:action",
							}},
						},
						Fallback: config.ActionRef{config.BasicAction{
							Name: "fall:back:action",
						}},
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
			config.Config{
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

hotkeys:
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
			config.Config{
				Mappings: map[config.KeyAlias]config.MappingKey{
					"K1": {Key: "fookey"},
					"K2": {Key: "barkey"},
					"K3": {Key: "bazkey"},
				},
				HotKeys: map[config.KeyAlias]config.GestureActions{
					"K1": {
						{
							Gesture: config.GestureSeries{"foo_down"},
							Action: config.ActionRef{config.BasicAction{
								Name: "foo:action",
							}},
						},
						{
							Gesture: config.GestureSeries{"bar_tap", "bar_hold", "bar_tap"},
							Action: config.ActionRef{config.BasicAction{
								Name: "bar:action",
							}},
						},
					},
					"fizzkey": {
						{
							Gesture: config.GestureSeries{"fizz_left"},
							Action: config.ActionRef{config.BasicAction{
								Name: "fizz:buzz",
							}},
						},
					},
					"K2": nil,
					"K3": {
						{
							Gesture: config.GestureSeries{"bar0", "bar1"},
							Action: config.ActionRef{config.BasicAction{
								Name: "bar:some:action",
							}},
						},
						{
							Gesture: config.GestureSeries{"bar2", "bar3"},
							Exact:   true,
							Action: config.ActionRef{config.BasicAction{
								Name: "bar:another:action",
							}},
						},
					},
					"K9": {
						{
							Gesture: config.GestureSeries{"foo"},
							Action: config.ActionRef{config.BasicAction{
								Name: "foo:action",
							}},
						},
						{
							Gesture: config.GestureSeries{"bar"},
							Action: config.ActionRef{config.BasicAction{
								Name: "bar:action",
							}},
						},
						{
							Gesture: config.GestureSeries{"fizz", "buzz"},
							Action: config.ActionRef{config.BasicAction{
								Name: "fizz:buzz",
								Args: []interface{}{1, 2, "fizz", 4, "buzz"},
							}},
						},
						{
							Gesture: config.GestureSeries{"baz"},
							Action: config.ActionRef{config.AppBranchAction{
								Branches: map[string]config.ActionRef{
									"/Far.app": {config.BasicAction{
										Name: "far:far",
									}},
								},
								Fallback: config.ActionRef{config.BasicAction{
									Name: "baz:baz",
								}},
							}},
						},
					},
				},
				Actions: map[string]config.ActionRef{
					"foo:0": {config.BasicAction{
						Name: "foo:0",
					}},
					"foo:1": {config.BasicAction{
						Name: "foo:1",
					}},
					"foo:2": {},
					"foo:3": {config.BasicAction{
						Name: "bar:3",
					}},
					"foo:4": {config.BasicAction{
						Name: "bar:4",
					}},
					"foo:5": {config.BasicAction{
						Name: "bar:5",
					}},
					"foo:6": {config.BasicAction{
						Name: "bar:6",
						Args: []interface{}{"foo", 1, 2, "bar", 3},
					}},
					"foo:7": {config.AppBranchAction{
						Branches: map[string]config.ActionRef{
							"/App/Foo": {config.BasicAction{
								Name: "foo:action",
							}},
							"/App/Bar": {},
							"/App/Baz": {config.BasicAction{
								Name: "baz:action",
							}},
						},
						Fallback: config.ActionRef{config.BasicAction{
							Name: "fall:back:action",
						}},
					}},
					"foo:9000": {config.AppBranchAction{
						Branches: map[string]config.ActionRef{
							"Foo.app": {config.BasicAction{
								Name: "foo:action",
							}},
							"": {config.BasicAction{
								Name: "bar:action",
							}},
							"Baz.app": {config.BasicAction{
								Name: "nested:action",
								Args: []interface{}{"inception"},
							}},
							"TooFar.app": {config.AppBranchAction{
								Branches: map[string]config.ActionRef{
									"Bar.Foo.app": {config.BasicAction{
										Name: "far:action",
									}},
									"Bar.Baz.app": {config.BasicAction{
										Name: "double:nested:action",
										Args: []interface{}{"in", "too", "deep"},
									}},
								},
								Fallback: config.ActionRef{config.BasicAction{
									Name: "nested:fallback",
								}},
							}},
						},
						Fallback: config.ActionRef{},
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
hotkeys:
  K1:
    "": foo:action
`,
			config.Config{},
			false,
		},
		{
			"empty gesture sequence 2",
			`
hotkeys:
  K1:
    - gesture: []
      action: foo:action
`,
			config.Config{},
			false,
		},
		{
			"empty gesture",
			`
hotkeys:
  K1:
    "foo.bar.": foo:action
`,
			config.Config{},
			false,
		},
		{
			"invalid action name",
			`
actions:
  foo:action: [123]
`,
			config.Config{},
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
			config.Config{},
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
			config.Config{},
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
