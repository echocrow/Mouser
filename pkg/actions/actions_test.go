package actions_test

import (
	"fmt"
	"testing"

	"github.com/birdkid/mouser/pkg/actions"
	"github.com/stretchr/testify/assert"
)

type i = interface{}

func TestFromHotkeys(t *testing.T) {
	t.Parallel()

	tests := map[string][]struct {
		args   []i
		wantOk bool
	}{
		"vol:down":     {{nil, true}},
		"vol:up":       {{nil, true}},
		"vol:mute":     {{nil, true}},
		"media:toggle": {{nil, true}},
		"media:prev":   {{nil, true}},
		"media:next":   {{nil, true}},

		"misc:none": {
			{nil, true},
			{[]i{}, true},
			{[]i{0}, false},
			{[]i{false}, false},
			{[]i{nil}, false},
			{[]i{""}, false},
		},

		"io:tap": {
			{nil, false},
			{[]i{}, false},
			{[]i{0}, false},
			{[]i{"f1"}, true},
			{[]i{"f1", 0}, false},
			{[]i{"f1", "shift"}, true},
		},

		"io:type": {
			{nil, false},
			{[]i{}, false},
			{[]i{0}, false},
			{[]i{"foo"}, true},
			{[]i{"foo", "bar"}, false},
		},

		"io:scroll": {
			{nil, false},
			{[]i{}, false},
			{[]i{0}, false},
			{[]i{1}, false},
			{[]i{1, 1}, true},
			{[]i{uint(1), uint(1)}, false},
			{[]i{-1, -1}, true},
			{[]i{-5, 7}, true},
			{[]i{0, 0}, true},
			{[]i{"1", "1"}, false},
			{[]i{1, 1, 1}, false},
		},

		"os:open": {
			{nil, false},
			{[]i{}, false},
			{[]i{1}, false},
			{[]i{"foo"}, true},
			{[]i{"foo", "-v"}, true},
			{[]i{"foo", 1}, false},
		},

		"os:cmd": {
			{nil, false},
			{[]i{}, false},
			{[]i{1}, false},
			{[]i{"foo"}, true},
			{[]i{"foo", "-v"}, true},
			{[]i{"foo", 1}, false},
		},

		"misc:sleep": {
			{nil, false},
			{[]i{}, false},
			{[]i{0}, false},
			{[]i{uint(0)}, false},
			{[]i{1}, true},
			{[]i{uint(1)}, true},
			{[]i{"1"}, false},
			{[]i{-1}, false},
		},
	}
	for actionName, tcs := range tests {
		actionName := actionName
		for i, tc := range tcs {
			tc := tc
			t.Run(fmt.Sprintf("%s #%d", actionName, i), func(t *testing.T) {
				t.Parallel()
				got, err := actions.New(actionName, tc.args...)
				if tc.wantOk {
					assert.NotNil(t, got)
					assert.NoError(t, err)
				} else {
					assert.Nil(t, got)
					assert.Error(t, err)
				}
			})
		}
	}
}
