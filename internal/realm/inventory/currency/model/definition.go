package model

// Definition describes one configured currency type.
type Definition struct {
	// Type identifies the protocol currency type.
	Type int32 `json:"type"`

	// Key identifies the localized currency name.
	Key string `json:"key"`

	// Ledger reports whether balance mutations require audit entries.
	Ledger bool `json:"ledger"`

	// Color reserves an optional client presentation color.
	Color string `json:"color,omitempty"`
}
