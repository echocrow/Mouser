package actions_test

import (
	"fmt"
	"testing"

	"github.com/echocrow/Mouser/pkg/actions"
)

func TestNewRequireAppCustom(t *testing.T) {
	const (
		app      = "app"
		fallback = "fallback"
	)

	calls := make(chan string, 2)

	var (
		appAct      act = func() { calls <- app }
		fallbackAct     = func() { calls <- fallback }
	)

	isAppRunning := false
	getRunning := func() bool { return isAppRunning }
	setRunning := func(isRunning bool) { isAppRunning = isRunning }

	t.Run("Branch", func(t *testing.T) {
		tests := []struct {
			name        string
			isRunning   bool
			appAct      act
			fallbackAct act
			wantCall    string
		}{
			{"calls action 1", true, appAct, fallbackAct, app},
			{"calls action 2", true, appAct, nil, app},
			{"skips action 1", true, nil, fallbackAct, ""},

			{"calls fallback 1", false, appAct, fallbackAct, fallback},
			{"calls fallback 2", false, nil, fallbackAct, fallback},
			{"skips fallback 1", false, appAct, nil, ""},

			{"skips nil 1", true, nil, nil, ""},
			{"skips nil 2", false, nil, nil, ""},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				setRunning(tc.isRunning)

				reqAppAction := actions.NewRequireAppCustom(
					getRunning,
					tc.appAct,
					tc.fallbackAct,
				)

				wantOk := tc.wantCall != ""
				assertActionCalls(t, reqAppAction, wantOk, tc.wantCall, calls)
			})
		}
	})

	t.Run("Repeat", func(t *testing.T) {
		reqAppAction := actions.NewRequireAppCustom(
			getRunning,
			appAct,
			fallbackAct,
		)

		tests := []struct {
			isRunning bool
			wantCall  string
		}{
			{true, app},
			{false, fallback},
			{true, app},
			{true, app},
			{true, app},
			{false, fallback},
			{false, fallback},
		}

		for i, tc := range tests {
			tc := tc
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				setRunning(tc.isRunning)
				wantOk := tc.wantCall != ""
				assertActionCalls(t, reqAppAction, wantOk, tc.wantCall, calls)
			})
		}
	})
}
