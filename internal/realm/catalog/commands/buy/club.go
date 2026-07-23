package buy

import (
	"context"
	"time"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outstatus "github.com/niflaot/pixels/networking/outbound/subscription/status"
)

const (
	// ClubLayout identifies Nitro's subscription purchase layout.
	ClubLayout = "vip_buy"
)

// ClubPurchaser buys subscription offers selected through the catalog.
type ClubPurchaser interface {
	// PurchaseClub buys an offer for the authenticated player.
	PurchaseClub(ctx context.Context, playerID int64, offerID int64, amount int32) (ClubPurchase, error)
}

// ClubPurchase contains committed subscription state needed by the client.
type ClubPurchase struct {
	// ExpiresAt stores the committed expiration instant.
	ExpiresAt time.Time
	// LifetimeActiveSeconds stores historical active club time.
	LifetimeActiveSeconds int64
	// LifetimeVIPSeconds stores historical active VIP time.
	LifetimeVIPSeconds int64
	// VIP reports whether the resulting tier is VIP.
	VIP bool
}

// handleClub purchases one subscription offer and acknowledges it to Nitro.
func (handler Handler) handleClub(ctx context.Context, connection netconn.Context, playerID int64, offerID int64, amount int32) error {
	if handler.Club == nil {
		return handler.sendError(ctx, connection, offerID, catalogservice.ErrOfferNotFound)
	}
	result, err := handler.Club.PurchaseClub(ctx, playerID, offerID, amount)
	if err != nil {
		return handler.sendError(ctx, connection, offerID, err)
	}
	packet, err := outok.Encode(offer.Offer{ID: int32(offerID)})
	if err != nil {
		return err
	}

	if err := connection.Send(ctx, packet); err != nil {
		return err
	}
	remaining := time.Until(result.ExpiresAt)
	if remaining < 0 {
		remaining = 0
	}
	packet, err = outstatus.Encode(outstatus.State{ProductName: "habbo_club",
		DaysToPeriodEnd: int32((remaining + 24*time.Hour - 1) / (24 * time.Hour)),
		ResponseType:    2, EverMember: true, VIP: result.VIP,
		PastClubDays:           int32(result.LifetimeActiveSeconds / 86_400),
		PastVIPDays:            int32(result.LifetimeVIPSeconds / 86_400),
		MinutesUntilExpiration: int32(remaining / time.Minute)})
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
