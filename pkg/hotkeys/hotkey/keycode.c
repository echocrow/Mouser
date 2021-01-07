#include <stdint.h>
#include "keycode.h"
#include "../base/os.h"

// Combine key name and key code.
struct MouserKeyName {
  const char* name;
  MouserKeyCode code;
};

// Map key names to key codes.
static const struct MouserKeyName KEYS[] = {
  { "f1",  K_F1 },
  { "f2",  K_F2 },
  { "f3",  K_F3 },
  { "f4",  K_F4 },
  { "f5",  K_F5 },
  { "f6",  K_F6 },
  { "f7",  K_F7 },
  { "f8",  K_F8 },
  { "f9",  K_F9 },
  { "f10", K_F10 },
  { "f11", K_F11 },
  { "f12", K_F12 },
  { "f13", K_F13 },
  { "f14", K_F14 },
  { "f15", K_F15 },
  { "f16", K_F16 },
  { "f17", K_F17 },
  { "f18", K_F18 },
  { "f19", K_F19 },
  { "f20", K_F20 },
  { "f21", K_F21 },
  { "f22", K_F22 },
  { "f23", K_F23 },
  { "f24", K_F24 },
};

// Convert a key int to the key name.
static const char* keyNameFromInt(uint32_t keyInt) {
  uint32_t b = 'z' - '0' + 1;

  uint32_t strLen = ceil(log(keyInt) / log(b));
  if (strLen == 0) {
    strLen = 1;
  }

  char* keyName = calloc(strLen + 1, sizeof(char));
  if (keyName == NULL) {
    return NULL;
  }

  uint32_t n = keyInt;
  for (uint32 i = 0; i < strLen; i++) {
    uint32_t d = n % b;
    char c = d + '0';
    keyName[i] = c;
    n /= b;
  }

  return keyName;
}

// Convert a key name to the key code.
static MouserKeyCode keyCodeFromName(const char* keyName) {
  if (keyName == NULL) {
    return K_NOT_A_KEY;
  }
  size_t arrSize = sizeof(KEYS) / sizeof(struct MouserKeyName);
  for (int i = 0; i < arrSize; i++) {
    if (strcmp(KEYS[i].name, keyName) == 0) {
      return KEYS[i].code;
    }
  }
  return K_NOT_A_KEY;
}

MouserKeyCode keyCodeFromInt(uint32_t keyInt) {
  const char* keyName = keyNameFromInt(keyInt);
  MouserKeyCode keyCode = keyCodeFromName(keyName);
  return keyCode;
}
