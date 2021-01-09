#include <stdbool.h>
#include <stdint.h>
#include "hotkey.h"
#include "keycode.h"
#include "../base/os.h"
#include "../base/vars.h"

#ifdef MOUSER_OS_MACOS
  // Registered hotkey references.
  static EventHotKeyRef* hotkeyRefs = NULL;
  static size_t hotkeyRefsLen = 0;
  static const size_t INITIAL_HOTKEY_BUF_LEN = 8;
  static const size_t MAX_HOTKEY_BUF_LEN = 256;
  static const size_t HOTKEY_REF_SIZE = sizeof(EventHotKeyRef);
#endif

bool registerHotkey(uint8_t id, uint32_t keyInt) {
  if (!id) {
    return false;
  }

  MouserKeyCode keyCode = keyCodeFromInt(keyInt);

  #ifdef MOUSER_OS_MACOS
    uint32_t modifiers = 0;

    EventHotKeyID eventID = { MOUSER_HOTKEY_SIG(), id };

    EventHotKeyRef hotkeyRef;

    // Initialize hotkey references.
    if (hotkeyRefs == NULL) {
      hotkeyRefsLen = INITIAL_HOTKEY_BUF_LEN;
      hotkeyRefs = (EventHotKeyRef*) calloc(hotkeyRefsLen, HOTKEY_REF_SIZE);
    }
    // Grow hotkey references buffer.
    else if (id > hotkeyRefsLen) {
      size_t prevLen = hotkeyRefsLen;
      hotkeyRefsLen = MAX_HOTKEY_BUF_LEN;
      hotkeyRefs = realloc(hotkeyRefs, hotkeyRefsLen * HOTKEY_REF_SIZE);
      memset(hotkeyRefs + prevLen, 0, (hotkeyRefsLen - prevLen) * HOTKEY_REF_SIZE);
    }
    if (hotkeyRefs == NULL) {
      // Memory allocation failed.
      return false;
    }

    OSStatus status = RegisterEventHotKey(
      keyCode,
      modifiers,
      eventID,
      GetEventDispatcherTarget(),
      0,
      &hotkeyRef
    );
    if (status != noErr) {
      // Registration failed.
      return false;
    }

    hotkeyRefs[id - 1] = hotkeyRef;
  #endif

  return true;
}

void unregisterHotkey(uint8_t id) {
  #ifdef MOUSER_OS_MACOS
    EventHotKeyRef hotkeyRef = hotkeyRefs[id - 1];
    if (hotkeyRef) {
      hotkeyRefs[id - 1] = 0;
      UnregisterEventHotKey(hotkeyRef);
    }
  #endif
}
