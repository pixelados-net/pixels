package repository

import (
	"context"
	"fmt"
)

const (
	// cloneRoomItemsSQL clones room furniture in one server-side statement.
	cloneRoomItemsSQL = `
with inserted as (
    insert into furniture_items
        (definition_id, owner_player_id, room_id, x, y, z, rotation, wall_position,
         extra_data, limited_edition_number, marketplace_reserved, gift_wrapped,
         gift_wrap_sprite_id, gift_wrap_box_id, gift_wrap_ribbon_id,
         gift_sender_player_id, gift_message, metadata)
    select definition_id, $3, $2, x, y, z, rotation, wall_position,
           extra_data, null, false, false, null, null, null, null, null, metadata
    from furniture_items
    where room_id = $1 and deleted_at is null
    returning 1
)
select count(*) from inserted`

	// listRoomBundleProductsSQL groups template furniture for Nitro previews.
	listRoomBundleProductsSQL = `select definition_id, count(*)::integer from furniture_items where room_id=$1 and deleted_at is null group by definition_id order by min(id)`
)

// CloneRoomItems copies active room items while clearing unique and transient state.
func (repository *Repository) CloneRoomItems(ctx context.Context, sourceRoomID int64, targetRoomID int64, targetOwnerID int64) (int, error) {
	var count int
	if err := repository.executorFor(ctx).QueryRow(ctx, cloneRoomItemsSQL, sourceRoomID, targetRoomID, targetOwnerID).Scan(&count); err != nil {
		return 0, fmt.Errorf("clone furniture from room %d: %w", sourceRoomID, err)
	}
	return count, nil
}

// ListRoomBundleProducts groups active room items by definition.
func (repository *Repository) ListRoomBundleProducts(ctx context.Context, roomID int64) ([]RoomBundleProduct, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, listRoomBundleProductsSQL, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room bundle products: %w", err)
	}
	defer rows.Close()
	products := make([]RoomBundleProduct, 0)
	for rows.Next() {
		var product RoomBundleProduct
		if err := rows.Scan(&product.DefinitionID, &product.Quantity); err != nil {
			return nil, fmt.Errorf("scan room bundle product: %w", err)
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

// roomBundleStoreAssertion verifies Repository implements bundle persistence.
var roomBundleStoreAssertion RoomBundleStore = (*Repository)(nil)
