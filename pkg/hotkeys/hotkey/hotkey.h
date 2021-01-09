#include <stdbool.h>
#include <stdint.h>
#include "../base/os.h"

// Register a hotkey.
bool registerHotkey(uint8_t id, uint32_t keyIndex);

// Unregister a hotkey.
void unregisterHotkey(uint8_t id);
