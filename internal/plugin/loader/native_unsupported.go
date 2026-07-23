//go:build !linux && !darwin && !freebsd

package loader

import "fmt"

// unsupportedOpener rejects native plugins on unsupported operating systems.
type unsupportedOpener struct{}

// NewNative creates an opener that reports platform incompatibility.
func NewNative() Opener { return unsupportedOpener{} }

// Open reports that Go native plugins are unavailable on this platform.
func (unsupportedOpener) Open(string) (Object, error) {
	return nil, fmt.Errorf("native Go plugins are unsupported on this platform")
}
