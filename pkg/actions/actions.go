// Package actions provides system actions.
package actions

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/go-vgo/robotgo"
)

// Action describes an action function.
type Action func()

// ActionCreator describes a function that creates an Action.
type ActionCreator func(args ...interface{}) (Action, error)

// Actions errors raised by package actions.
var (
	ErrInvalidActionName = errors.New("action name is invalid")
	ErrInvalidActionArgs = errors.New("action arguments are invalid")
)

// New creates an action.
func New(actionName string, args ...interface{}) (Action, error) {
	if action, ok := basicActions[actionName]; ok {
		if len(args) != 0 {
			return nil, ErrInvalidActionArgs
		}
		return action, nil
	}
	if actionCreator, ok := actionCreators[actionName]; ok {
		return actionCreator(args...)
	}
	return nil, ErrInvalidActionName
}

var basicActions = map[string]Action{
	// vol:down decreases the audio volume level.
	"vol:down": func() { robotgo.KeyTap("audio_vol_down") },
	// vol:up increases the audio volume level.
	"vol:up": func() { robotgo.KeyTap("audio_vol_up") },
	// vol:mute toggles between muting and unmuting audio.
	"vol:mute": func() { robotgo.KeyTap("audio_mute") },

	// media:toggle toggles between playing and pausing the current media.
	"media:toggle": func() { robotgo.KeyTap("audio_play") },
	// media:prev rewindes the current or jumps back to the previous media record.
	"media:prev": func() { robotgo.KeyTap("audio_prev") },
	// media:prev forwards to the next media record.
	"media:next": func() { robotgo.KeyTap("audio_next") },

	// os:close-window closes the current window.
	"os:close-window": func() { robotgo.CloseWindow() },

	// misc:none does nothing.
	"misc:none": func() {},
}

var actionCreators = map[string]ActionCreator{
	// io:tap triggers a short key press & release.
	// Arguments:
	// - modifiers ...string: Optional modifiers to hold during the key tap, e.g.
	//   "shift", "command", etc.
	// - key string: The name of the key to tap, e.g. "f1", "a", "enter" etc.
	"io:tap": func(args ...interface{}) (Action, error) {
		l := len(args)
		if l < 1 {
			return nil, ErrInvalidActionArgs
		} else if key, ok := stringifySingle(args[l-1]); !ok {
			return nil, ErrInvalidActionArgs
		} else {
			modifiers, ok := stringify(args[:l-1])
			if !ok {
				return nil, ErrInvalidActionArgs
			}
			tap := func() { robotgo.KeyTap(key, destringify(modifiers)...) }
			return tap, nil
		}
	},
	// io:type writes out the given text.
	// Arguments:
	// - text string: The text to type out.
	"io:type": func(args ...interface{}) (Action, error) {
		if len(args) != 1 {
			return nil, ErrInvalidActionArgs
		} else if text, ok := stringifySingle(args[0]); !ok {
			return nil, ErrInvalidActionArgs
		} else {
			write := func() { robotgo.TypeStr(text) }
			return write, nil
		}
	},
	// io:scroll triggers a scroll event.
	// Arguments:
	// - x int: The distance in pixels to scroll horizontally (left to right).
	// - y int: The distance in pixels to scroll vertically (top to bottom).
	"io:scroll": func(args ...interface{}) (Action, error) {
		if len(args) != 2 {
			return nil, ErrInvalidActionArgs
		} else if x, ok := args[0].(int); !ok {
			return nil, ErrInvalidActionArgs
		} else if y, ok := args[1].(int); !ok {
			return nil, ErrInvalidActionArgs
		} else {
			scroll := func() { robotgo.Scroll(-x, -y) }
			return scroll, nil
		}
	},

	// os:open opens a file or application.
	// Arguments:
	// - file string: The path to the file or application to open.
	// - openArgs ...string: List of extra arguments to pass to the open command.
	"os:open": func(args ...interface{}) (Action, error) {
		if len(args) < 1 {
			return nil, ErrInvalidActionArgs
		} else if app, ok := stringifySingle(args[0]); !ok {
			return nil, ErrInvalidActionArgs
		} else {
			openArgs, ok := stringify(args[1:])
			if !ok {
				return nil, ErrInvalidActionArgs
			}
			cmdArgs := make([]string, len(args))
			copy(cmdArgs, openArgs)
			cmdArgs[len(openArgs)] = app
			cmd := exec.Command("open", cmdArgs...)
			open := func() {
				c := *cmd
				c.Start()
			}
			return open, nil
		}
	},
	// os:cmd runs a custom command.
	// Arguments:
	// - cmd string: The command name or path.
	// - cmdArgs ...string: List of extra arguments to pass to the command.
	"os:cmd": func(args ...interface{}) (Action, error) {
		if len(args) < 1 {
			return nil, ErrInvalidActionArgs
		} else if cmdName, ok := stringifySingle(args[0]); !ok {
			return nil, ErrInvalidActionArgs
		} else {
			cmdArgs, ok := stringify(args[1:])
			if !ok {
				return nil, ErrInvalidActionArgs
			}
			cmd := exec.Command(cmdName, cmdArgs...)
			run := func() {
				c := *cmd
				c.Start()
			}
			return run, nil
		}
	},

	// misc:sleep pauses action execution for a given time.
	// Arguments:
	// - duration int|uint: The duration of the pause in milliseconds > 0.
	"misc:sleep": func(args ...interface{}) (Action, error) {
		if len(args) != 1 {
			return nil, ErrInvalidActionArgs
		}
		var ms uint
		switch t := args[0].(type) {
		case uint:
			ms = t
		case int:
			if t < 0 {
				return nil, ErrInvalidActionArgs
			}
			ms = uint(t)
		default:
			return nil, ErrInvalidActionArgs
		}
		if ms == 0 {
			return nil, ErrInvalidActionArgs
		}
		d := time.Duration(ms) * time.Millisecond
		sleep := func() { time.Sleep(d) }
		return sleep, nil
	},
}

func stringify(args []interface{}) ([]string, bool) {
	strs := make([]string, len(args))
	for i, arg := range args {
		if str, ok := stringifySingle(arg); ok {
			strs[i] = str
		} else {
			return nil, false
		}
	}
	return strs, true
}

func stringifySingle(arg interface{}) (string, bool) {
	switch str := arg.(type) {
	case string:
		return str, true
	case int, uint, float64:
		return fmt.Sprint(str), true
	default:
		return "", false
	}
}

func destringify(strs []string) []interface{} {
	out := make([]interface{}, len(strs))
	for i, str := range strs {
		out[i] = str
	}
	return out
}
