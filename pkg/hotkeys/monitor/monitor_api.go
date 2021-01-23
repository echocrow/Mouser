package monitor

// Engine describes a hotkeys monitor engine.
//go:generate mockery --name "Engine"
type Engine interface {
	Init() (ok bool)
	Start(m *Monitor) error
	Stop()
	Deinit() (ok bool)
}
