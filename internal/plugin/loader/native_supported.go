//go:build linux || darwin || freebsd

package loader

import stdplugin "plugin"

// nativeOpener loads plugins through Go's platform-native plugin package.
type nativeOpener struct{}

// nativeObject adapts the standard library symbol type to Object.
type nativeObject struct {
	// plugin stores the opened standard library object.
	plugin *stdplugin.Plugin
}

// NewNative creates the platform native object opener.
func NewNative() Opener { return nativeOpener{} }

// Open loads one Go plugin shared object.
func (nativeOpener) Open(path string) (Object, error) {
	opened, err := stdplugin.Open(path)
	if err != nil {
		return nil, err
	}

	return nativeObject{plugin: opened}, nil
}

// Lookup resolves one exported native symbol.
func (object nativeObject) Lookup(name string) (any, error) {
	return object.plugin.Lookup(name)
}
