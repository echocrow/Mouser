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

// KeyIndex represents keys by an internal index, e.g. 1, 2, etc.
type KeyIndex uint32

// NotAKey represents a missing or invalid key.
const NotAKey = KeyIndex(0)

// KeyIndices maps key names to key indices.
var KeyIndices = map[KeyName]KeyIndex{
	"":    NotAKey,
	"f1":  1,
	"f2":  2,
	"f3":  3,
	"f4":  4,
	"f5":  5,
	"f6":  6,
	"f7":  7,
	"f8":  8,
	"f9":  9,
	"f10": 10,
	"f11": 11,
	"f12": 12,
	"f13": 13,
	"f14": 14,
	"f15": 15,
	"f16": 16,
	"f17": 17,
	"f18": 18,
	"f19": 19,
	"f20": 20,
	"f21": 21,
	"f22": 22,
	"f23": 23,
	"f24": 24,
}

// NameToIndex converts a key to a key int.
func NameToIndex(key KeyName) (KeyIndex, error) {
	keyIndex, ok := KeyIndices[key]
	if !ok || keyIndex == NotAKey {
		return NotAKey, ErrInvalidKeyName
	}
	return keyIndex, nil
}
