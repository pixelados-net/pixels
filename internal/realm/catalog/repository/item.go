package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

const (
	// itemColumns stores the shared catalog offer projection.
	itemColumns = `id, page_id, definition_id, room_bundle_template_room_id, name, cost_credits, cost_points, points_type, amount, limited_stack, limited_sells, bundle_discount_enabled, giftable, club_only, order_num, enabled, extra_data, scheduled_at, created_at, updated_at, deleted_at, version`

	// listItemsSQL lists active catalog offers with an optional page filter.
	listItemsSQL = `select ` + itemColumns + ` from catalog_items where deleted_at is null and ($1::bigint is null or page_id=$1) order by page_id, order_num, id`

	// findItemSQL finds one active catalog offer.
	findItemSQL = `select ` + itemColumns + ` from catalog_items where id=$1 and deleted_at is null`

	// createItemSQL creates one catalog offer.
	createItemSQL = `insert into catalog_items (page_id, definition_id, room_bundle_template_room_id, name, cost_credits, cost_points, points_type, amount, limited_stack, limited_sells, bundle_discount_enabled, giftable, club_only, order_num, enabled, extra_data, scheduled_at)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) returning ` + itemColumns

	// updateItemSQL updates one active catalog offer using its version.
	updateItemSQL = `update catalog_items set page_id=$2, definition_id=$3, room_bundle_template_room_id=$4, name=$5, cost_credits=$6, cost_points=$7,
points_type=$8, amount=$9, limited_stack=$10, limited_sells=$11, bundle_discount_enabled=$12, giftable=$13,
club_only=$14, order_num=$15, enabled=$16, extra_data=$17, scheduled_at=$18, updated_at=now(), version=version+1
where id=$1 and version=$19 and deleted_at is null returning ` + itemColumns

	// softDeleteItemSQL soft deletes one active catalog offer using its version.
	softDeleteItemSQL = `update catalog_items set deleted_at=now(), updated_at=now(), version=version+1 where id=$1 and version=$2 and deleted_at is null`

	// sanitizeListSQL lists active definitions without active enabled offers.
	sanitizeListSQL = `select ` + definitionColumns + ` from furniture_definitions d left join catalog_items i on i.definition_id=d.id and i.deleted_at is null and i.enabled=true where d.deleted_at is null and i.id is null order by d.id`

	// countSanitizeSQL counts active definitions without active enabled offers.
	countSanitizeSQL = `select count(*) from furniture_definitions d left join catalog_items i on i.definition_id=d.id and i.deleted_at is null and i.enabled=true where d.deleted_at is null and i.id is null`
)

// ListItems lists active offers, optionally restricted to one page.
func (repository *Repository) ListItems(ctx context.Context, pageID *int64) ([]catalogmodel.Item, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, listItemsSQL, pageID)
	if err != nil {
		return nil, fmt.Errorf("list catalog items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}

// FindItemByID finds one active catalog offer.
func (repository *Repository) FindItemByID(ctx context.Context, id int64) (catalogmodel.Item, bool, error) {
	return repository.queryItem(ctx, findItemSQL, id)
}

// CreateItem creates one catalog offer.
func (repository *Repository) CreateItem(ctx context.Context, item catalogmodel.Item) (catalogmodel.Item, error) {
	created, err := scanItem(repository.executorFor(ctx).QueryRow(ctx, createItemSQL, itemValues(item)...))
	if err != nil {
		return catalogmodel.Item{}, fmt.Errorf("create catalog item %q: %w", item.Name, err)
	}

	return created, nil
}

// UpdateItem updates one offer using optimistic locking.
func (repository *Repository) UpdateItem(ctx context.Context, item catalogmodel.Item) (catalogmodel.Item, bool, error) {
	arguments := append([]any{item.ID}, itemValues(item)...)
	arguments = append(arguments, item.Version.Version)

	return repository.queryItem(ctx, updateItemSQL, arguments...)
}

// SoftDeleteItem soft deletes one offer using optimistic locking.
func (repository *Repository) SoftDeleteItem(ctx context.Context, id int64, version int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, softDeleteItemSQL, id, version)
	if err != nil {
		return false, fmt.Errorf("soft delete catalog item %d: %w", id, err)
	}

	return tag.RowsAffected() == 1, nil
}

// SanitizeList lists active furniture definitions without an active offer.
func (repository *Repository) SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, sanitizeListSQL)
	if err != nil {
		return nil, fmt.Errorf("list catalog sanitize definitions: %w", err)
	}
	defer rows.Close()

	return scanDefinitions(rows)
}

// CountEnabledDefinitionsWithoutOffer counts active definitions without enabled offers.
func (repository *Repository) CountEnabledDefinitionsWithoutOffer(ctx context.Context) (int64, error) {
	var count int64
	if err := repository.executorFor(ctx).QueryRow(ctx, countSanitizeSQL).Scan(&count); err != nil {
		return 0, fmt.Errorf("count catalog sanitize definitions: %w", err)
	}

	return count, nil
}

// queryItem scans one optional catalog offer.
func (repository *Repository) queryItem(ctx context.Context, query string, arguments ...any) (catalogmodel.Item, bool, error) {
	item, err := scanItem(repository.executorFor(ctx).QueryRow(ctx, query, arguments...))
	if errors.Is(err, pgx.ErrNoRows) {
		return catalogmodel.Item{}, false, nil
	}
	if err != nil {
		return catalogmodel.Item{}, false, err
	}

	return item, true, nil
}

// itemValues maps offer persistence values in statement order.
func itemValues(item catalogmodel.Item) []any {
	var definitionID any = item.DefinitionID
	if item.IsRoomBundle() {
		definitionID = nil
	}
	return []any{item.PageID, definitionID, item.RoomBundleTemplateRoomID, item.Name, item.CostCredits, item.CostPoints, item.PointsType, item.Amount, item.LimitedStack, item.LimitedSells, item.BundleDiscountEnabled, item.Giftable, item.ClubOnly, item.OrderNum, item.Enabled, item.ExtraData, item.ScheduledAt}
}
