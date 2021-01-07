package hotkey

import (
	"errors"
	"math"
)

// Keycode errors raised by package hotkey.
var (
	ErrKeyCodeConvFailed = errors.New("key code conversion failed")
)

// KeyInt holds the integer ID of a key.
type KeyInt uint32

// KeyToInt converts a key to a key int.
func KeyToInt(key string) (KeyInt, error) {
	maxStrLen := 6
	if len(key) > maxStrLen {
		return 0, ErrKeyCodeConvFailed
	}

	keyInt := KeyInt(0)
	b := uint('z' - '0' + 1)
	for i, char := range key {
		if char < '0' || char > 'z' {
			return 0, ErrKeyCodeConvFailed
		}
		charInt := KeyInt(char - '0')
		keyInt += charInt * KeyInt(powUint(b, uint(i)))
	}

	return keyInt, nil
}

func powUint(b, p uint) uint {
	return uint(math.Pow(float64(b), float64(p)))
}
