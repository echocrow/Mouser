#include <Carbon/Carbon.h>

extern void RunApplicationEventLoop(void);
extern void QuitApplicationEventLoop(void);

OSStatus handleHotkeyEventDown(
  EventHandlerCallRef nextCallRef,
  EventRef event,
  void* context
);

OSStatus handleHotkeyEventUp(
  EventHandlerCallRef nextCallRef,
  EventRef event,
  void* context
);
