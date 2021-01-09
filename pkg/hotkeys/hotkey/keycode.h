#include <stdint.h>
#include "../base/os.h"

#ifdef MOUSER_OS_MACOS
  #include <Carbon/Carbon.h>

  typedef CGKeyCode MouserKeyCode;

  // Map mouser key codes to macOS key codes.
  // Key codes from "HIToolbox/Events.h".
  enum MouserKeyCode {
    K_NOT_A_KEY = 9999,
    K_F1 = kVK_F1,
    K_F2 = kVK_F2,
    K_F3 = kVK_F3,
    K_F4 = kVK_F4,
    K_F5 = kVK_F5,
    K_F6 = kVK_F6,
    K_F7 = kVK_F7,
    K_F8 = kVK_F8,
    K_F9 = kVK_F9,
    K_F10 = kVK_F10,
    K_F11 = kVK_F11,
    K_F12 = kVK_F12,
    K_F13 = kVK_F13,
    K_F14 = kVK_F14,
    K_F15 = kVK_F15,
    K_F16 = kVK_F16,
    K_F17 = kVK_F17,
    K_F18 = kVK_F18,
    K_F19 = kVK_F19,
    K_F20 = kVK_F20,
    K_F21 = K_NOT_A_KEY,
    K_F22 = K_NOT_A_KEY,
    K_F23 = K_NOT_A_KEY,
    K_F24 = K_NOT_A_KEY,
  };
#endif

// Represent a key by its internal index.
typedef uint32_t MouserKeyIndex;

// Convert a key index to the key code.
MouserKeyCode keyCodeFromIndex(MouserKeyIndex keyIndex);
