// Package wardrobe implements PostgreSQL player wardrobe persistence.
package wardrobe

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists wardrobe outfits.
type Repository struct {
	// pool owns transaction scopes.
	pool *postgres.Pool
}

// New creates a wardrobe repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Outfits returns all saved outfits ordered by slot.
func (repository *Repository) Outfits(ctx context.Context, playerID int64) ([]playerwardrobe.Outfit, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `select slot_id,figure,gender from player_wardrobe_outfits where player_id=$1 order by slot_id`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list wardrobe outfits: %w", err)
	}
	defer rows.Close()
	outfits := make([]playerwardrobe.Outfit, 0, playerwardrobe.MaxSlot)
	for rows.Next() {
		var outfit playerwardrobe.Outfit
		if err = rows.Scan(&outfit.SlotID, &outfit.Figure, &outfit.Gender); err != nil {
			return nil, fmt.Errorf("scan wardrobe outfit: %w", err)
		}
		outfits = append(outfits, outfit)
	}
	return outfits, rows.Err()
}

// SaveOutfit atomically upserts one outfit slot.
func (repository *Repository) SaveOutfit(ctx context.Context, playerID int64, outfit playerwardrobe.Outfit) error {
	_, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `insert into player_wardrobe_outfits(player_id,slot_id,figure,gender) values($1,$2,$3,$4) on conflict(player_id,slot_id) do update set figure=excluded.figure,gender=excluded.gender,updated_at=now()`, playerID, outfit.SlotID, outfit.Figure, outfit.Gender)
	if err != nil {
		return fmt.Errorf("save wardrobe outfit: %w", err)
	}
	return nil
}

// Clothing returns complete clothing unlock state.
func (repository *Repository) Clothing(ctx context.Context, playerID int64) (playerwardrobe.ClothingSnapshot, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `select figure_set_id,product_code from player_clothing_sets where player_id=$1 order by figure_set_id`, playerID)
	if err != nil {
		return playerwardrobe.ClothingSnapshot{}, fmt.Errorf("list clothing sets: %w", err)
	}
	defer rows.Close()
	snapshot := playerwardrobe.ClothingSnapshot{FigureSetIDs: make([]int32, 0), ProductCodes: make([]string, 0)}
	seen := make(map[string]struct{})
	for rows.Next() {
		var figureSetID int32
		var productCode string
		if err = rows.Scan(&figureSetID, &productCode); err != nil {
			return playerwardrobe.ClothingSnapshot{}, err
		}
		snapshot.FigureSetIDs = append(snapshot.FigureSetIDs, figureSetID)
		if _, found := seen[productCode]; !found {
			seen[productCode] = struct{}{}
			snapshot.ProductCodes = append(snapshot.ProductCodes, productCode)
		}
	}
	return snapshot, rows.Err()
}

// RedeemClothing consumes one inventory item only when it adds an unlock.
func (repository *Repository) RedeemClothing(ctx context.Context, playerID int64, itemID int64) (result playerwardrobe.RedeemResult, err error) {
	err = postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		var productCode string
		findErr := executor.QueryRow(txCtx, `select cp.product_code from furniture_items fi join clothing_products cp on cp.definition_id=fi.definition_id and cp.enabled where fi.id=$1 and fi.owner_player_id=$2 and fi.room_id is null and fi.deleted_at is null and not fi.marketplace_reserved for update of fi`, itemID, playerID).Scan(&productCode)
		if errors.Is(findErr, pgx.ErrNoRows) {
			return playerwardrobe.ErrInvalidClothingItem
		}
		if findErr != nil {
			return findErr
		}
		command, insertErr := executor.Exec(txCtx, `insert into player_clothing_sets(player_id,figure_set_id,product_code) select $1,figure_set_id,product_code from clothing_product_sets where product_code=$2 on conflict do nothing`, playerID, productCode)
		if insertErr != nil {
			return insertErr
		}
		if command.RowsAffected() == 0 {
			return nil
		}
		if _, insertErr = executor.Exec(txCtx, `update furniture_items set deleted_at=now(),updated_at=now(),version=version+1 where id=$1`, itemID); insertErr != nil {
			return insertErr
		}
		result.Applied = true
		return nil
	})
	if err != nil {
		return playerwardrobe.RedeemResult{}, fmt.Errorf("redeem clothing item: %w", err)
	}
	result.Snapshot, err = repository.Clothing(ctx, playerID)
	return result, err
}
