#include <stdbool.h>
#include <stdint.h>
#include "monitor.h"
#include "../base/os.h"
#include "../base/vars.h"

#ifdef CGO
  #include "_cgo_export.h"
#endif

static bool isRunning = false;

#ifdef MOUSER_OS_MACOS
  #include <Carbon/Carbon.h>

  // Carbon: CarbonEvents.h
  extern void RunApplicationEventLoop(void);
  extern void QuitApplicationEventLoop(void);

  static void handleHotkeyEvent(EventRef event, bool isDown) {
    EventHotKeyID eventID;
    OSStatus status = GetEventParameter(event, kEventParamDirectObject, typeEventHotKeyID, NULL, sizeof(eventID), NULL, &eventID);
    if (status == noErr && eventID.signature == MOUSER_HOTKEY_SIG()) {
      goHandleHotkeyEvent(eventID.id, isDown);
    }
  }

  static OSStatus handleHotkeyEventDown(EventHandlerCallRef nextCallRef, EventRef event, void *context) {
    handleHotkeyEvent(event, true);
    return CallNextEventHandler(nextCallRef, event);
  }

  static OSStatus handleHotkeyEventUp(EventHandlerCallRef nextCallRef, EventRef event, void *context) {
    handleHotkeyEvent(event, false);
    return CallNextEventHandler(nextCallRef, event);
  }

  static EventHandlerRef evHandlerRefDown;
  static EventHandlerRef evHandlerRefUp;
#endif

bool initMonitor() {
  #ifdef MOUSER_OS_MACOS
    OSStatus status;
    EventHandlerUPP handlerUPP;

    if (!evHandlerRefDown) {
      EventTypeSpec evSpecDown = { kEventClassKeyboard, kEventHotKeyPressed};
      handlerUPP = NewEventHandlerUPP(handleHotkeyEventDown);
      status = InstallEventHandler(GetEventDispatcherTarget(), handlerUPP, 1, &evSpecDown, NULL, &evHandlerRefDown);
      if (status != noErr) {
        return false;
      }
    }

    if (!evHandlerRefUp) {
      EventTypeSpec evSpecUp = { kEventClassKeyboard, kEventHotKeyReleased };
      handlerUPP = NewEventHandlerUPP(handleHotkeyEventUp);
      status = InstallEventHandler(GetEventDispatcherTarget(), handlerUPP, 1, &evSpecUp, NULL, &evHandlerRefUp);
      if (status != noErr) {
        return false;
      }
    }
  #endif

  return true;
}

void startMonitor() {
  if (isRunning) return;
  isRunning = true;

  #ifdef MOUSER_OS_MACOS
    RunApplicationEventLoop();
  #endif
}

bool deinitMonitor() {
  bool success = true;

  #ifdef MOUSER_OS_MACOS
    OSStatus status;

    if (evHandlerRefDown) {
      status = RemoveEventHandler(evHandlerRefDown);
      evHandlerRefDown = NULL;
      success = success && (status == noErr);
    }

    if (evHandlerRefUp) {
      status = RemoveEventHandler(evHandlerRefUp);
      evHandlerRefUp = NULL;
      success = success && (status == noErr);
    }
  #endif

  return success;
}

void stopMonitor() {
  #ifdef MOUSER_OS_MACOS
    QuitApplicationEventLoop();
  #endif

  isRunning = false;
}
