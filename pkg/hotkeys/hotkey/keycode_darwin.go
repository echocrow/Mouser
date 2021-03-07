package hotkey

// #cgo darwin LDFLAGS: -framework Carbon
// #include <Carbon/Carbon.h>
import "C"

// KeyCode represents keys as platform-specific code.
type KeyCode C.CGKeyCode

// NotAKeyCode is the empty KeyCode value.
const NotAKeyCode KeyCode = 0

var keyCodes = map[KeyName]KeyCode{
	"f1":  C.kVK_F1,
	"f2":  C.kVK_F2,
	"f3":  C.kVK_F3,
	"f4":  C.kVK_F4,
	"f5":  C.kVK_F5,
	"f6":  C.kVK_F6,
	"f7":  C.kVK_F7,
	"f8":  C.kVK_F8,
	"f9":  C.kVK_F9,
	"f10": C.kVK_F10,
	"f11": C.kVK_F11,
	"f12": C.kVK_F12,
	"f13": C.kVK_F13,
	"f14": C.kVK_F14,
	"f15": C.kVK_F15,
	"f16": C.kVK_F16,
	"f17": C.kVK_F17,
	"f18": C.kVK_F18,
	"f19": C.kVK_F19,
	"f20": C.kVK_F20,
}

// NameToCode converts a key to a key code.
func NameToCode(key KeyName) (KeyCode, error) {
	if keyCode, ok := keyCodes[key]; ok {
		return keyCode, nil
	}
	return NotAKeyCode, ErrInvalidKeyName
}
