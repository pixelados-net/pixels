// Package realm contains emulator-only pixel realm behavior.
package realm

// Realm describes a pixel-protocol realm served by the emulator.
type Realm struct {
	// Name identifies the realm.
	Name string
}

// New creates a realm with the provided name.
func New(name string) Realm {
	return Realm{Name: name}
}

// Ready reports whether the realm can accept sessions.
func (realm Realm) Ready() bool {
	return realm.Name != ""
}
