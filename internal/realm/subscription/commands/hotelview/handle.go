package hotelview

import (
	"context"
	"math"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	"github.com/niflaot/pixels/networking/codec"
	outseconds "github.com/niflaot/pixels/networking/outbound/progression/competition/seconds"
	outbonus "github.com/niflaot/pixels/networking/outbound/subscription/bonusrare/info"
)

const countdownLayout = "2006-01-02 15:04"

// Handle executes one hotel-view request.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	switch envelope.Command.Action {
	case StartCampaign:
		return nil
	case Countdown:
		return handler.sendCountdown(ctx, envelope.Command)
	case BonusRare:
		return handler.sendBonusRare(ctx, envelope.Command)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// sendBonusRare reads and projects configured currency progress.
func (handler Handler) sendBonusRare(ctx context.Context, input Command) error {
	player, err := catalogsession.Player(input.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	progress, err := handler.Subscriptions.BonusRareInfo(ctx, player.ID())
	if err != nil {
		return err
	}
	packet, err := outbonus.Encode(progress.ProductType, progress.ProductClassID, progress.Threshold, progress.Remaining)
	return send(ctx, input.Connection, packet, err)
}

// sendCountdown parses an absolute UTC minute and projects non-negative seconds.
func (handler Handler) sendCountdown(ctx context.Context, input Command) error {
	remaining := secondsUntil(input.Value, time.Now().UTC())
	packet, encodeErr := outseconds.Encode(input.Value, clampSeconds(remaining))
	return send(ctx, input.Connection, packet, encodeErr)
}

// secondsUntil returns whole non-negative seconds until one UTC minute.
func secondsUntil(value string, now time.Time) int64 {
	target, err := time.ParseInLocation(countdownLayout, value, time.UTC)
	if err != nil {
		return 0
	}
	remaining := int64(target.Sub(now).Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// clampSeconds converts a countdown duration to protocol range.
func clampSeconds(value int64) int32 {
	if value <= 0 {
		return 0
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}

// send writes one encoded hotel-view packet.
func send(ctx context.Context, connection codecSender, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// codecSender writes encoded packets.
type codecSender interface {
	// Send queues one packet for the active connection.
	Send(context.Context, codec.Packet) error
}
