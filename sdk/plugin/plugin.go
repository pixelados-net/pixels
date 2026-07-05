// Package plugin contains the controlled plugin-facing SDK primitives.
package plugin

// Metadata describes a plugin without granting runtime capabilities.
type Metadata struct {
	// Name identifies the plugin.
	Name string
	// Version identifies the plugin release.
	Version string
}

// Valid reports whether metadata is complete enough to identify a plugin.
func (metadata Metadata) Valid() bool {
	return metadata.Name != "" && metadata.Version != ""
}
