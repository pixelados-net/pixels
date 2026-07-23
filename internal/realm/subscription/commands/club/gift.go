package club

import (
	"context"
	"math"
	"strconv"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outgiftinfo "github.com/niflaot/pixels/networking/outbound/subscription/gift/info"
	outnotification "github.com/niflaot/pixels/networking/outbound/subscription/gift/notification"
	outselected "github.com/niflaot/pixels/networking/outbound/subscription/gift/selected"
)

// handleGift sends gift information or claims one selected gift.
func (handler Handler) handleGift(ctx context.Context, input Command, playerID int64) error {
	membership, found, err := handler.Subscriptions.Membership(ctx, playerID)
	if err != nil {
		return err
	}
	offers, err := handler.clubGifts(ctx, playerID, membership)
	if err != nil {
		return err
	}
	if input.Action == GiftInfo {
		days, available := giftSummary(membership, found)
		packet, encodeErr := outgiftinfo.Encode(days, available, offers)
		return send(ctx, input.Connection, packet, encodeErr)
	}
	if !found {
		return core.ErrMembershipNotFound
	}
	for _, gift := range offers {
		if gift.Offer.LocalizationID != input.GiftName {
			continue
		}
		if !gift.Selectable {
			return core.ErrOfferNotFound
		}
		result, claimErr := handler.Subscriptions.ClaimClubGift(ctx, playerID, int64(gift.Offer.ID))
		if claimErr != nil {
			return claimErr
		}
		packet, encodeErr := outselected.Encode(input.GiftName, gift.Offer.Products)
		if err := send(ctx, input.Connection, packet, encodeErr); err != nil {
			return err
		}
		itemIDs := make([]int64, 0, len(result.GrantedItems))
		for _, item := range result.GrantedItems {
			itemIDs = append(itemIDs, item.ID)
		}
		packet, encodeErr = outunseen.EncodeOwned(itemIDs)
		if err := send(ctx, input.Connection, packet, encodeErr); err != nil {
			return err
		}
		packet, encodeErr = outrefresh.Encode()
		if err := send(ctx, input.Connection, packet, encodeErr); err != nil {
			return err
		}
		remaining := core.RemainingClubGifts(membership) - 1
		packet, encodeErr = outnotification.Encode(remaining)
		return send(ctx, input.Connection, packet, encodeErr)
	}

	return core.ErrOfferNotFound
}

// giftSummary returns Nitro's availability fields for one membership state.
func giftSummary(membership record.Membership, found bool) (int32, int32) {
	if !found || membership.Level == record.LevelNone {
		return 0, 0
	}

	return daysUntilGift(membership), core.RemainingClubGifts(membership)
}

// clubGifts maps every visible club-gift page offer.
func (handler Handler) clubGifts(ctx context.Context, playerID int64, membership record.Membership) ([]outgiftinfo.Gift, error) {
	pages, err := handler.Catalog.Pages(ctx, playerID, true)
	if err != nil {
		return nil, err
	}
	result := make([]outgiftinfo.Gift, 0)
	for _, page := range pages {
		if page.Layout != "club_gifts" {
			continue
		}
		_, items, readErr := handler.Catalog.Page(ctx, page.ID, playerID, true)
		if readErr != nil {
			return nil, readErr
		}
		for _, item := range items {
			mapped, mapErr := handler.mapGift(ctx, item)
			if mapErr != nil {
				return nil, mapErr
			}
			daysRequired, _ := strconv.ParseInt(item.ExtraData, 10, 32)
			lifetimeDays := int32(membership.LifetimeActiveSeconds / int64((24*time.Hour)/time.Second))
			result = append(result, outgiftinfo.Gift{Offer: mapped, VIP: item.ClubOnly,
				DaysRequired: int32(daysRequired), Selectable: lifetimeDays >= int32(daysRequired)})
		}
	}

	return result, nil
}

// mapGift maps one catalog item to its wire offer.
func (handler Handler) mapGift(ctx context.Context, item catalogmodel.Item) (catalogoffer.Offer, error) {
	products := handler.Catalog.Products(ctx, item.ID)
	if len(products) == 0 {
		products = []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}
	}
	definitions := make(map[int64]furnituremodel.Definition, len(products))
	for _, product := range products {
		definition, _, err := handler.Catalog.Definition(ctx, product.DefinitionID)
		if err != nil {
			return catalogoffer.Offer{}, err
		}
		definitions[product.DefinitionID] = definition
	}

	return catalogprojection.OfferProducts(item, products, definitions)
}

// daysUntilGift calculates the next monthly gift boundary.
func daysUntilGift(membership record.Membership) int32 {
	if core.RemainingClubGifts(membership) > 0 {
		return 0
	}
	remaining := core.ClubGiftPeriodSeconds - membership.LifetimeActiveSeconds%core.ClubGiftPeriodSeconds
	return int32(math.Ceil(float64(remaining) / float64((24*time.Hour)/time.Second)))
}
