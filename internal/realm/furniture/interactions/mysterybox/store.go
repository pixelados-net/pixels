package mysterybox

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Keys stores one account's visible box and key colors.
type Keys struct {
	// BoxColor stores the current box color identifier.
	BoxColor string
	// KeyColor stores the current key color identifier.
	KeyColor string
}

// Store persists mystery-box account state.
type Store interface {
	// FindKeys finds or returns the neutral key tracker state.
	FindKeys(context.Context, int64) (Keys, error)
}

// Repository implements mystery-box account persistence.
type Repository struct{ executor postgres.Executor }

// NewRepository creates mystery-box account persistence.
func NewRepository(executor postgres.Executor) *Repository { return &Repository{executor: executor} }

// FindKeys finds or returns the neutral key tracker state.
func (repository *Repository) FindKeys(ctx context.Context, playerID int64) (Keys, error) {
	var value Keys
	err := postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, `select box_color,key_color from player_mysterybox_keys where player_id=$1`, playerID).Scan(&value.BoxColor, &value.KeyColor)
	if errors.Is(err, pgx.ErrNoRows) {
		return Keys{BoxColor: "default", KeyColor: "default"}, nil
	}
	return value, err
}
