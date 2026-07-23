package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	// createLimitedUnitsSQL creates one numbered LTD series.
	createLimitedUnitsSQL = `insert into catalog_item_limited_units (catalog_item_id, unit_number) select $1, value from generate_series(1, $2) as value`

	// syncLimitedUnitsSQL reconciles unsold LTD units while preserving completed sales.
	syncLimitedUnitsSQL = `with removed as (
	delete from catalog_item_limited_units where catalog_item_id=$1 and owner_player_id is null and unit_number>$2
	)
	insert into catalog_item_limited_units (catalog_item_id, unit_number)
	select $1, value from generate_series(1, $2) as value on conflict (catalog_item_id, unit_number) do nothing`

	// reserveLimitedUnitSQL atomically claims the lowest available LTD unit.
	reserveLimitedUnitSQL = `update catalog_item_limited_units set owner_player_id=$2, sold_at=now()
where id=(select id from catalog_item_limited_units where catalog_item_id=$1 and owner_player_id is null order by unit_number for update skip locked limit 1)
returning unit_number`

	// completeLimitedUnitSQL links the granted furniture and advances LTD inventory.
	completeLimitedUnitSQL = `with completed as (
update catalog_item_limited_units set furniture_item_id=$4
where catalog_item_id=$1 and unit_number=$2 and owner_player_id=$3 and furniture_item_id is null returning id
), advanced as (
update catalog_items set limited_sells=limited_sells+1, enabled=(limited_sells+1 < limited_stack), updated_at=now(), version=version+1
where id=$1 and limited_stack>limited_sells and exists(select 1 from completed) returning id
)
select exists(select 1 from advanced)`
)

// CreateLimitedUnits creates numbered units for an LTD offer.
func (repository *Repository) CreateLimitedUnits(ctx context.Context, catalogItemID int64, quantity int32) error {
	_, err := repository.executorFor(ctx).Exec(ctx, createLimitedUnitsSQL, catalogItemID, quantity)
	if err != nil {
		return fmt.Errorf("create %d limited units for catalog item %d: %w", quantity, catalogItemID, err)
	}

	return nil
}

// SyncLimitedUnits reconciles unsold numbered units with an LTD stack size.
func (repository *Repository) SyncLimitedUnits(ctx context.Context, catalogItemID int64, quantity int32) error {
	_, err := repository.executorFor(ctx).Exec(ctx, syncLimitedUnitsSQL, catalogItemID, quantity)
	if err != nil {
		return fmt.Errorf("sync %d limited units for catalog item %d: %w", quantity, catalogItemID, err)
	}

	return nil
}

// ReserveLimitedUnit atomically reserves the lowest available LTD number.
func (repository *Repository) ReserveLimitedUnit(ctx context.Context, catalogItemID int64, playerID int64) (int32, bool, error) {
	var number int32
	err := repository.executorFor(ctx).QueryRow(ctx, reserveLimitedUnitSQL, catalogItemID, playerID).Scan(&number)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("reserve limited unit for catalog item %d: %w", catalogItemID, err)
	}

	return number, true, nil
}

// CompleteLimitedUnit links a reservation and advances the offer sale count.
func (repository *Repository) CompleteLimitedUnit(ctx context.Context, catalogItemID int64, unitNumber int32, playerID int64, furnitureItemID int64) (bool, error) {
	var completed bool
	err := repository.executorFor(ctx).QueryRow(ctx, completeLimitedUnitSQL, catalogItemID, unitNumber, playerID, furnitureItemID).Scan(&completed)
	if err != nil {
		return false, fmt.Errorf("complete limited unit %d for catalog item %d: %w", unitNumber, catalogItemID, err)
	}

	return completed, nil
}
