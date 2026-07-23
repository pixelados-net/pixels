package identity

import "context"

// Store commits atomic identity replacements and denormalized room owners.
type Store interface {
	// Rename commits one one-shot username replacement and audit row.
	Rename(context.Context, int64, string) (RenameResult, error)
}
