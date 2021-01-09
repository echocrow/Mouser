#pragma once
#ifndef MOUSER_VARS_H
#define MOUSER_VARS_H

#include "os.h"

#ifdef MOUSER_OS_MACOS
  #include <Carbon/Carbon.h>

  // Hotkey event signature.
  // This is NOT meant as a private const. However, VS Code's go launch process
  // fails when this is non-static. The "go build" command works either way.
  static inline const FourCharCode MOUSER_HOTKEY_SIG() {
    return 'MSER';
  }
#endif

#endif
