package hotkey

import (
	"errors"
)

// Keycode errors raised by package hotkey.
var (
	ErrInvalidKeyName = errors.New("key name is invalid or unsupported")
)

// KeyName represents keys by string names, e.g. "f1", "a", etc.
type KeyName string
