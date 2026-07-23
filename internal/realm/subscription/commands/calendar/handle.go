package calendar

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	"github.com/niflaot/pixels/internal/realm/subscription/access"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	outdata "github.com/niflaot/pixels/networking/outbound/subscription/calendar/data"
	outopened "github.com/niflaot/pixels/networking/outbound/subscription/calendar/opened"
	outseasonal "github.com/niflaot/pixels/networking/outbound/subscription/calendar/seasonal"
)

// Handle executes one calendar command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if envelope.Command.Action == Seasonal {
		return handler.sendSeasonal(ctx, envelope.Command, player.ID())
	}
	staff := false
	if envelope.Command.Action == OpenStaff && handler.Permissions != nil {
		staff, err = handler.Permissions.HasPermission(ctx, player.ID(), access.CalendarStaffBypass)
		if err != nil {
			return err
		}
	}
	day, err := handler.Subscriptions.OpenCalendarDoor(ctx, player.ID(), envelope.Command.CampaignName, envelope.Command.DayNumber, staff)
	if err != nil {
		packet, encodeErr := outopened.Encode(false, "", "", "")
		return send(ctx, envelope.Command.Connection, packet, encodeErr)
	}
	productName, furnitureName := rewardNames(ctx, handler.Furniture, day.ProductDefinitionID)
	packet, err := outopened.Encode(true, productName, day.CustomImage, furnitureName)
	if err := send(ctx, envelope.Command.Connection, packet, err); err != nil {
		return err
	}

	return handler.sendCalendarData(ctx, envelope.Command, player.ID())
}

// sendCalendarData sends current opened and missed campaign days.
func (handler Handler) sendCalendarData(ctx context.Context, input Command, playerID int64) error {
	campaign, _, opened, err := handler.Subscriptions.CalendarData(ctx, playerID, input.CampaignName)
	if err != nil {
		return err
	}
	current := int32(time.Since(campaign.StartDate) / (24 * time.Hour))
	missed := missedDays(current, opened)
	packet, err := outdata.Encode(campaign.Name, campaign.Image, current, campaign.DayCount, opened, missed)
	return send(ctx, input.Connection, packet, err)
}

// sendSeasonal sends today's separate seasonal catalog offer.
func (handler Handler) sendSeasonal(ctx context.Context, input Command, playerID int64) error {
	seasonal, found, err := handler.Subscriptions.SeasonalOffer(ctx, time.Now())
	if err != nil || !found {
		return err
	}
	item, found := handler.Catalog.Item(seasonal.CatalogItemID)
	if !found {
		return core.ErrOfferNotFound
	}
	mapped, err := mapOffer(ctx, handler.Catalog, item)
	if err != nil {
		return err
	}
	packet, err := outseasonal.Encode(int32(seasonal.CatalogPageID), mapped)
	return send(ctx, input.Connection, packet, err)
}
