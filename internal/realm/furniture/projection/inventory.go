// Package projection maps furniture mutations to Nitro inventory packets.
package projection

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	netconn "github.com/niflaot/pixels/networking/connection"
	outadd "github.com/niflaot/pixels/networking/outbound/inventory/furniture/add"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/furniture/list"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
)

// Inventory sends committed removals, one addition, and a refresh marker.
func Inventory(ctx context.Context, connection netconn.Context, removed []int64, granted furnituremodel.Item, definition furnituremodel.Definition) error {
	for _, itemID := range removed {
		packet, err := outremove.Encode(itemID)
		if err != nil {
			return err
		}
		if err = connection.Send(ctx, packet); err != nil {
			return err
		}
	}
	packet, err := outadd.Encode(itemRecord(granted, definition))
	if err != nil {
		return err
	}
	if err = connection.Send(ctx, packet); err != nil {
		return err
	}
	refresh, err := outrefresh.Encode()
	if err != nil {
		return err
	}
	return connection.Send(ctx, refresh)
}

// Removed sends one committed removal and a refresh marker.
func Removed(ctx context.Context, connection netconn.Context, itemID int64) error {
	packet, err := outremove.Encode(itemID)
	if err != nil {
		return err
	}
	if err = connection.Send(ctx, packet); err != nil {
		return err
	}
	refresh, err := outrefresh.Encode()
	if err != nil {
		return err
	}
	return connection.Send(ctx, refresh)
}

// itemRecord maps one furniture instance to its inventory wire record.
func itemRecord(item furnituremodel.Item, definition furnituremodel.Definition) outlist.Item {
	kind := outlist.KindFloor
	if definition.Kind == furnituremodel.KindWall {
		kind = outlist.KindWall
	}
	box, ribbon := int32(0), int32(0)
	if item.GiftWrapBoxID != nil {
		box = *item.GiftWrapBoxID
	}
	if item.GiftWrapRibbonID != nil {
		ribbon = *item.GiftWrapRibbonID
	}
	return outlist.Item{ID: item.ID, SpriteID: definition.SpriteID, Kind: kind, ExtraData: item.ExtraData, AllowInventoryStack: definition.AllowInventoryStack, GiftWrapped: item.GiftWrapped, AllowTrade: definition.AllowTrade, AllowMarketplace: definition.AllowMarketplaceSale, AllowRecycle: definition.AllowRecycle, LimitedEditionNumber: item.LimitedEditionNumber, GiftBoxID: box, GiftRibbonID: ribbon}
}
