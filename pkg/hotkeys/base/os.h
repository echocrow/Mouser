#pragma once
#ifndef MOUSER_OS_H
#define MOUSER_OS_H

// Detect OS.
// See: https://sourceforge.net/p/predef/wiki/OperatingSystems/
#if defined(__APPLE__) && defined(__MACH__)
  #define MOUSER_OS_MACOS
#else
  #error "This platform is not supported!"
#endif

#endif
