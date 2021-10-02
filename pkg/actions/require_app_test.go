package actions_test

import (
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
		appAct act = func() { calls <- app }
		// nilAct act = nil
		// emptyAppAct act = func() { calls <- emptyApp }
		fallbackAct = func() { calls <- fallback }
	)

	isAppRunning := false
	getRunning := func() bool { return isAppRunning }
	setRunning := func(isRunning bool) { isAppRunning = isRunning }

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

			assertNoCall(t, calls, "premature")

			reqAppAction()

			if tc.wantCall != "" {
				assertCall(t, calls, tc.wantCall)
			} else {
				assertNoCall(t, calls, "nil")
			}

			assertNoCall(t, calls, "subsequent")
		})
	}
}
