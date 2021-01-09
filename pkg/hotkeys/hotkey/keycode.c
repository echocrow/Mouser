#include <stdint.h>
#include "keycode.h"
#include "../base/os.h"

// Map key indeices to key codes.
static const MouserKeyCode KEY_CODES[] = {
  K_NOT_A_KEY,
  K_F1,
  K_F2,
  K_F3,
  K_F4,
  K_F5,
  K_F6,
  K_F7,
  K_F8,
  K_F9,
  K_F10,
  K_F11,
  K_F12,
  K_F13,
  K_F14,
  K_F15,
  K_F16,
  K_F17,
  K_F18,
  K_F19,
  K_F20,
  K_F21,
  K_F22,
  K_F23,
  K_F24,
};
static const size_t KEY_CODES_LEN = sizeof(KEY_CODES) / sizeof(MouserKeyCode);

MouserKeyCode keyCodeFromIndex(MouserKeyIndex keyIndex) {
  return (keyIndex < KEY_CODES_LEN) ? KEY_CODES[keyIndex] : K_NOT_A_KEY;
}
