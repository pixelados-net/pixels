package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

const altarColumns = `definition_id,enabled,created_at,updated_at,version`

// Altar finds one altar registration.
func (repository *Repository) Altar(ctx context.Context, definitionID int64, includeDisabled bool) (craftingrecord.Altar, bool, error) {
	query := `select ` + altarColumns + ` from crafting_altars where definition_id=$1 and ($2 or enabled)`
	var altar craftingrecord.Altar
	err := repository.executorFor(ctx).QueryRow(ctx, query, definitionID, includeDisabled).Scan(&altar.DefinitionID, &altar.Enabled, &altar.CreatedAt, &altar.UpdatedAt, &altar.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return craftingrecord.Altar{}, false, nil
	}
	return altar, err == nil, err
}

// ListAltars lists altar registrations with optional disabled rows.
func (repository *Repository) ListAltars(ctx context.Context, includeDisabled bool) ([]craftingrecord.Altar, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select `+altarColumns+` from crafting_altars where $1 or enabled order by definition_id`, includeDisabled)
	if err != nil {
		return nil, fmt.Errorf("list crafting altars: %w", err)
	}
	defer rows.Close()
	altars := make([]craftingrecord.Altar, 0, 8)
	for rows.Next() {
		var altar craftingrecord.Altar
		if err = rows.Scan(&altar.DefinitionID, &altar.Enabled, &altar.CreatedAt, &altar.UpdatedAt, &altar.Version); err != nil {
			return nil, err
		}
		altars = append(altars, altar)
	}
	return altars, rows.Err()
}

// UpsertAltar creates or re-enables an altar.
func (repository *Repository) UpsertAltar(ctx context.Context, definitionID int64) (craftingrecord.Altar, bool, error) {
	var altar craftingrecord.Altar
	var created bool
	query := `insert into crafting_altars(definition_id) values($1) on conflict(definition_id) do update set enabled=true,updated_at=now(),version=crafting_altars.version+1 returning ` + altarColumns + `,(xmax=0)`
	err := repository.executorFor(ctx).QueryRow(ctx, query, definitionID).Scan(&altar.DefinitionID, &altar.Enabled, &altar.CreatedAt, &altar.UpdatedAt, &altar.Version, &created)
	return altar, created, err
}

// DisableAltar disables an altar and its recipes.
func (repository *Repository) DisableAltar(ctx context.Context, definitionID int64) (bool, error) {
	changed := false
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		tag, err := repository.executorFor(txCtx).Exec(txCtx, `update crafting_altars set enabled=false,updated_at=now(),version=version+1 where definition_id=$1 and enabled`, definitionID)
		if err != nil {
			return err
		}
		changed = tag.RowsAffected() > 0
		if !changed {
			return nil
		}
		_, err = repository.executorFor(txCtx).Exec(txCtx, `update crafting_recipes set enabled=false,updated_at=now(),version=version+1 where altar_definition_id=$1 and enabled`, definitionID)
		return err
	})
	return changed, err
}
