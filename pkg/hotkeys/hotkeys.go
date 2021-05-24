// Package hotkeys provides functionality to create hotkeys and monitor their
// events. Hotkeys are registered keyboard buttons that trigger an action.
// Hotkey events include key-down and key-up events.
package hotkeys

import (
	"sync"

	"github.com/echocrow/Mouser/pkg/hotkeys/hotkey"
	"github.com/echocrow/Mouser/pkg/hotkeys/monitor"
	"github.com/echocrow/Mouser/pkg/log"
)

var (
	sharedMonitor     *monitor.Monitor
	sharedMonitorOnce sync.Once
)

// DefaultMonitor constructs a shared default hotkey monitor.
func DefaultMonitor(initWithLog bool) *monitor.Monitor {
	sharedMonitorOnce.Do(func() {
		sharedMonitor = defaultMonitor(initWithLog)
	})
	return sharedMonitor
}

// defaultMonitor constructs a default hotkey monitor.
func defaultMonitor(withLog bool) *monitor.Monitor {
	hkReg := defaultRegistry()
	monitor := monitor.New(hkReg, nil)
	if withLog {
		monitor.SetLogCb(log.NewCallback("Monitor"))
	}
	return monitor
}

// DefaultRegistry constructs a default hotkey registry.
func defaultRegistry() hotkey.Registry {
	return hotkey.NewRegistry(nil, nil)
}
