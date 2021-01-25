package monitor

// Engine describes a hotkeys monitor engine.
type Engine interface {
	InitMonitor() (ok bool)
	StartMonitor(m *Monitor) error
	StopMonitor()
	DeinitMonitor() (ok bool)
}
