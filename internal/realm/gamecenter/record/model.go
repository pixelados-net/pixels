// Package record defines Game Center persistence contracts.
package record

import "context"

// LaunchKind identifies the external launch packet shape.
type LaunchKind string

const (
	// LaunchURL uses Nitro's simple URL launcher.
	LaunchURL LaunchKind = "url"
	// LaunchParameters uses Nitro's parameterized launcher.
	LaunchParameters LaunchKind = "params"
)

// Game describes one external game registration.
type Game struct {
	// ID identifies the game type.
	ID int32
	// Name stores its display key.
	Name string
	// BackgroundColor stores a six-digit RGB value.
	BackgroundColor string
	// TextColor stores a six-digit RGB value.
	TextColor string
	// AssetURL stores lobby artwork.
	AssetURL string
	// SupportURL stores support documentation.
	SupportURL string
	// LaunchURL stores the external game entry point.
	LaunchURL string
	// LaunchKind stores the selected launcher shape.
	LaunchKind LaunchKind
	// Enabled reports whether players may launch the game.
	Enabled bool
	// Version provides optimistic concurrency control.
	Version int64
}

// Store persists Game Center state.
type Store interface {
	// ListGames returns all games in stable id order.
	ListGames(context.Context, bool) ([]Game, error)
	// FindGame returns one game by id.
	FindGame(context.Context, int32) (Game, bool, error)
	// UpsertScore stores one player's best weekly score.
	UpsertScore(context.Context, int32, int64, int32, int32, int64) error
}
