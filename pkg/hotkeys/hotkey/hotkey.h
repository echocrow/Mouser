#include <stdbool.h>
#include <stdint.h>
#include "keycode.h"
#include "../base/os.h"

// Represent a hotkey ID.
typedef uint8_t MouserHotKeyID;

// Register a hotkey.
bool registerHotkey(MouserHotKeyID id, MouserKeyIndex keyIndex);

// Unregister a hotkey.
void unregisterHotkey(MouserHotKeyID id);
