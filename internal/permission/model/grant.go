package model

import "github.com/niflaot/pixels/internal/permission"

// Grant stores one allowed or denied permission node.
type Grant struct {
	// Node identifies the granted capability or wildcard.
	Node permission.Node
	// Allowed reports whether the matching capability is allowed.
	Allowed bool
}
