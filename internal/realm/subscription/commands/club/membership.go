package club

import (
	"context"
	"errors"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/networking/codec"
	outbalance "github.com/niflaot/pixels/networking/outbound/subscription/balance"
	outbuilders "github.com/niflaot/pixels/networking/outbound/subscription/builders/count"
	outbuildersstatus "github.com/niflaot/pixels/networking/outbound/subscription/builders/status"
	outkickback "github.com/niflaot/pixels/networking/outbound/subscription/kickback"
	outoffers "github.com/niflaot/pixels/networking/outbound/subscription/offers"
	outextended "github.com/niflaot/pixels/networking/outbound/subscription/offers/extended"
	outsms "github.com/niflaot/pixels/networking/outbound/subscription/sms"
	outstatus "github.com/niflaot/pixels/networking/outbound/subscription/status"
)

// Handle executes one club command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	if envelope.Command.Action == BuildersCount {
		packet, err := outbuilders.Encode()
		if err := send(ctx, envelope.Command.Connection, packet, err); err != nil {
			return err
		}
		status, err := outbuildersstatus.Encode()
		return send(ctx, envelope.Command.Connection, status, err)
	}
	if envelope.Command.Action == SMS {
		packet, err := outsms.Encode()
		return send(ctx, envelope.Command.Connection, packet, err)
	}
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	switch envelope.Command.Action {
	case Status:
		membership, _, readErr := handler.Subscriptions.Membership(ctx, player.ID())
		packet, encodeErr := outstatus.Encode(statusProjection(membership, 1, time.Now(), envelope.Command.ProductName))
		if readErr != nil {
			return readErr
		}
		return send(ctx, envelope.Command.Connection, packet, encodeErr)
	case Offers:
		return handler.sendOffers(ctx, envelope.Command.Connection, player.ID(), false)
	case Extension:
		return handler.sendExtension(ctx, envelope.Command.Connection, player.ID(), envelope.Command.VIP)
	case PurchaseHC, PurchaseVIP:
		membership, purchaseErr := handler.Subscriptions.PurchaseExtensionOffer(ctx, player.ID(), envelope.Command.OfferID, envelope.Command.Action == PurchaseVIP)
		if errors.Is(purchaseErr, currencyservice.ErrInsufficientBalance) {
			packet, encodeErr := outbalance.Encode(true, false, -1)
			return send(ctx, envelope.Command.Connection, packet, encodeErr)
		}
		if purchaseErr != nil {
			return purchaseErr
		}
		player.SetClub(clubProjection(membership))
		packet, encodeErr := outstatus.Encode(statusProjection(membership, 2, time.Now(), ""))
		return send(ctx, envelope.Command.Connection, packet, encodeErr)
	case Kickback:
		return handler.sendKickback(ctx, envelope.Command.Connection, player.ID())
	case GiftInfo, SelectGift:
		return handler.handleGift(ctx, envelope.Command, player.ID())
	default:
		return nil
	}
}

// sendOffers sends one filtered club offer list.
func (handler Handler) sendOffers(ctx context.Context, connection interface {
	Send(context.Context, codec.Packet) error
}, playerID int64, deals bool) error {
	offers, err := handler.Subscriptions.Offers(ctx, deals)
	if err != nil {
		return err
	}
	membership, _, err := handler.Subscriptions.Membership(ctx, playerID)
	if err != nil {
		return err
	}
	now := time.Now()
	mapped := make([]outoffers.Offer, 0, len(offers))
	for _, offer := range offers {
		mapped = append(mapped, offerProjection(offer, membership, now))
	}
	packet, err := outoffers.Encode(mapped)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendExtension sends the first extension deal matching a tier.
func (handler Handler) sendExtension(ctx context.Context, connection interface {
	Send(context.Context, codec.Packet) error
}, playerID int64, vip bool) error {
	offers, err := handler.Subscriptions.Offers(ctx, true)
	if err != nil {
		return err
	}
	for _, offer := range offers {
		if offer.VIP == vip {
			membership, _, membershipErr := handler.Subscriptions.Membership(ctx, playerID)
			if membershipErr != nil {
				return membershipErr
			}
			now := time.Now()
			projected := offerProjection(offer, membership, now)
			packet, encodeErr := outextended.Encode(projected, projected.PriceCredits,
				projected.PricePoints, projected.PointsType, membershipDaysLeft(membership, now))
			if encodeErr != nil {
				return encodeErr
			}
			return connection.Send(ctx, packet)
		}
	}

	return handler.sendOffers(ctx, connection, playerID, false)
}

// membershipDaysLeft calculates whole or partial remaining membership days.
func membershipDaysLeft(membership record.Membership, now time.Time) int32 {
	if membership.ExpiresAt == nil || !membership.ExpiresAt.After(now) {
		return 0
	}
	return int32((membership.ExpiresAt.Sub(now) + 24*time.Hour - 1) / (24 * time.Hour))
}

// sendKickback sends one current payday projection.
func (handler Handler) sendKickback(ctx context.Context, connection interface {
	Send(context.Context, codec.Packet) error
}, playerID int64) error {
	info, err := handler.Subscriptions.CurrentPaydayInfo(ctx, playerID)
	if err != nil {
		return err
	}
	first := ""
	if info.Membership.StartedAt != nil {
		first = info.Membership.StartedAt.Format("2006-01-02")
	}
	packet, err := outkickback.Encode(outkickback.Info{Streak: info.StreakDays, FirstSubscriptionDate: first,
		Percentage: handler.Subscriptions.KickbackPercentage(), CreditsSpent: clampInt32(info.CreditsSpent),
		StreakBonus: clampInt32(info.StreakBonus), MonthlyBonus: clampInt32(info.MonthlyBonus), MinutesUntilPayday: info.MinutesUntilPayday})
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// offerProjection maps one durable club offer.
func offerProjection(offer record.Offer, membership record.Membership, now time.Time) outoffers.Offer {
	anchor := now
	if membership.ExpiresAt != nil && membership.ExpiresAt.After(anchor) {
		anchor = *membership.ExpiresAt
	}
	expires := anchor.Add(time.Duration(offer.DayCount) * 24 * time.Hour)
	remaining := expires.Sub(now)
	days := int32((remaining + 24*time.Hour - 1) / (24 * time.Hour))

	return outoffers.Offer{ID: int32(offer.ID), Name: offer.Name,
		PriceCredits: clampInt32(offer.PriceCredits), PricePoints: clampInt32(offer.PricePoints),
		PointsType: offer.PointsType, VIP: offer.VIP, Months: offer.DayCount / 31,
		ExtraDays: offer.DayCount % 31, DaysLeftAfterPurchase: days,
		Year: int32(expires.Year()), Month: int32(expires.Month()), Day: int32(expires.Day())}
}
