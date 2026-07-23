package record

import sharedmodel "github.com/niflaot/pixels/pkg/model"

// Preference stores navigator window preferences.
type Preference struct {
	// PlayerID identifies the player.
	PlayerID int64

	// WindowX stores the navigator x coordinate.
	WindowX int

	// WindowY stores the navigator y coordinate.
	WindowY int

	// WindowWidth stores the navigator width.
	WindowWidth int

	// WindowHeight stores the navigator height.
	WindowHeight int

	// LeftPanelHidden reports whether the left panel is hidden.
	LeftPanelHidden bool

	// ResultsMode stores the default search result mode.
	ResultsMode int16

	// Timestamps contains durable record timestamps.
	sharedmodel.Timestamps
}

// CategoryPreference stores navigator result-list display state.
type CategoryPreference struct {
	// PlayerID identifies the player.
	PlayerID int64

	// Code stores the result-list code.
	Code string

	// Collapsed reports whether the list is collapsed.
	Collapsed bool

	// ListMode stores the result-list mode.
	ListMode int16

	// Timestamps contains durable record timestamps.
	sharedmodel.Timestamps
}
