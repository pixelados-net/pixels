package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

const (
	// definitionColumns contains the shared furniture definition select list.
	definitionColumns = `id, sprite_id, name, public_name, kind, width, length, stack_height::float8, allow_stack, allow_walk, allow_sit, allow_lay, allow_inventory_stack, allow_trade, allow_marketplace_sale, redeemable_credits, interaction_type, interaction_modes_count, multiheight, custom_params, metadata, created_at, updated_at, deleted_at, version`

	// findDefinitionByIDSQL reads one active furniture definition by id.
	findDefinitionByIDSQL = `select ` + definitionColumns + ` from furniture_definitions where id = $1 and deleted_at is null`

	// listDefinitionsSQL reads active furniture definitions.
	listDefinitionsSQL = `select ` + definitionColumns + ` from furniture_definitions where deleted_at is null order by id asc`
)

// FindDefinitionByID finds an active furniture definition by id.
func (repository *Repository) FindDefinitionByID(ctx context.Context, id int64) (furnituremodel.Definition, bool, error) {
	definition, err := scanDefinition(repository.executorFor(ctx).QueryRow(ctx, findDefinitionByIDSQL, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return furnituremodel.Definition{}, false, nil
	}
	if err != nil {
		return furnituremodel.Definition{}, false, err
	}

	return definition, true, nil
}

// ListDefinitions lists active furniture definitions.
func (repository *Repository) ListDefinitions(ctx context.Context) ([]furnituremodel.Definition, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, listDefinitionsSQL)
	if err != nil {
		return nil, fmt.Errorf("list furniture definitions: %w", err)
	}
	defer rows.Close()

	return scanDefinitions(rows)
}
