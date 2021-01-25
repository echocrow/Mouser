package hotkey

// IDProvider describes a hotkey ID provider.
type IDProvider interface {
	NextID() ID
}

// Engine describes a hotkey registry engine.
type Engine interface {
	Register(id ID, keyIndex KeyIndex) (ok bool)
	Unregister(id ID)
}

// Registrar describes a hotkey registrar.
type Registrar interface {
	Add(key KeyName) (ID, error)
	Remove(id ID)
}
