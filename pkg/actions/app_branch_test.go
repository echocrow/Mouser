package actions_test

import (
	"testing"

	"github.com/echocrow/Mouser/pkg/actions"
	"github.com/stretchr/testify/assert"
)

type act = actions.Action

func TestNewAppBranchCustom(t *testing.T) {
	const (
		app1     = "app_1"
		app2     = "app_2"
		emptyApp = ""
		fallback = "fallback"
	)

	calls := make(chan string, 2)

	branches := map[string]act{
		"known_app_1": func() { calls <- app1 },
		"known_app_2": func() { calls <- app2 },
		"nil_app":     nil,
		"":            func() { calls <- emptyApp },
	}
	fallbackAction := func() { calls <- fallback }

	currApp := ""
	getApp := func() string { return currApp }
	setApp := func(app string) { currApp = app }

	branchAction := actions.NewAppBranchCustom(
		branches,
		fallbackAction,
		getApp,
		'/',
	)

	tests := []struct {
		name     string
		app      string
		wantCall string
		wantOk   bool
	}{
		{"calls app #1 action", "known_app_1", app1, true},
		{"calls app #2 action", "known_app_2", app2, true},
		{"does not call nil action", "nil_app", "", false},
		{"re-calls app #1 action", "known_app_1", app1, true},
		{"calls empty app action", "", emptyApp, true},
		{"calls fallback on misc app #2", "unknown_app", fallback, true},
		{"calls app action from path suffix", "known_app_1/suffix", app1, true},
		{"calls fallback from non-path suffix", "known_app_1suffix", fallback, true},
		{"calls fallback from extra prefix", "prefix/known_app_1", fallback, true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertNoCall(t, calls, "premature")

			setApp(tc.app)
			branchAction()

			if tc.wantOk {
				gotCall := <-calls
				assert.Equal(t, tc.wantCall, gotCall)
			} else {
				assertNoCall(t, calls, "nil")
			}

			assertNoCall(t, calls, "subsequent")
		})
	}
}

func assertNoCall(t *testing.T, calls <-chan string, callDesc string) {
	select {
	case call := <-calls:
		t.Errorf("Do not want any %s branch action calls, got %s.", callDesc, call)
	default:
	}
}
