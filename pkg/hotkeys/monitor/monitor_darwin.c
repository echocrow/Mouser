#include "monitor_darwin.h"

#ifdef CGO
  #include "_cgo_export.h"
#endif

OSStatus handleHotkeyEventDown(
  EventHandlerCallRef nextCallRef,
  EventRef event,
  void *context
) {
  goHandleHotkeyEvent(event, true);
  return CallNextEventHandler(nextCallRef, event);
}

OSStatus handleHotkeyEventUp(
  EventHandlerCallRef nextCallRef,
  EventRef event,
  void *context
) {
  goHandleHotkeyEvent(event, false);
  return CallNextEventHandler(nextCallRef, event);
}

CGEventRef handleMouseButtonEvent(
  CGEventTapProxy proxy,
  CGEventType eventType,
  CGEventRef event,
  void *context
) {
  return goHandleMouseEvent(event, eventType);
}
