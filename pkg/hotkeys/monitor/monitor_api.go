package monitor

// Engine describes a hotkeys monitor engine.
//go:generate mockery --name "Engine"
type Engine interface {
	InitMonitor() (ok bool)
	StartMonitor(m *Monitor) error
	StopMonitor()
	DeinitMonitor() (ok bool)
}
