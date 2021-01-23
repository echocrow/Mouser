package hotkey

// IDProvider describes a hotkey ID provider.
//go:generate mockery --name "IDProvider"
type IDProvider interface {
	NextID() ID
}

// Engine describes a hotkey registry engine.
//go:generate mockery --name "Engine"
type Engine interface {
	Register(id ID, keyIndex KeyIndex) (ok bool)
	Unregister(id ID)
}

// Registrar describes a hotkey registrar.
//go:generate mockery --name "Registrar"
type Registrar interface {
	Add(key KeyName) (ID, error)
	Remove(id ID)
}
