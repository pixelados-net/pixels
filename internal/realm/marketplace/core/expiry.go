package core

import (
	"context"
	marketexpired "github.com/niflaot/pixels/internal/realm/marketplace/events/expired"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// Expire returns expired listings to seller inventories in bounded transactions.
func (service *Service) Expire(ctx context.Context) (int, error) {
	count := 0
	for {
		var batch int
		var expired []marketrecord.Listing
		err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
			listings, err := service.store.ExpireListings(txCtx, 100)
			if err != nil {
				return err
			}
			batch = len(listings)
			expired = listings
			for _, listing := range listings {
				if err := service.furniture.ReleaseFromMarketplace(txCtx, listing.FurnitureItemID, listing.SellerPlayerID); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return count, err
		}
		count += batch
		if service.events != nil {
			for _, listing := range expired {
				_ = service.events.Publish(ctx, bus.Event{Name: marketexpired.Name, Payload: marketexpired.Payload{ListingID: listing.ID, SellerPlayerID: listing.SellerPlayerID, FurnitureItemID: listing.FurnitureItemID}})
			}
		}
		if batch < 100 {
			break
		}
	}
	if count > 0 {
		service.invalidate(ctx)
	}
	return count, nil
}
