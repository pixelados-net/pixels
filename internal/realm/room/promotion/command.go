package promotion

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancel "github.com/niflaot/pixels/networking/outbound/room/promotion/cancel"
	outeligibility "github.com/niflaot/pixels/networking/outbound/room/promotion/eligibility"
	outerror "github.com/niflaot/pixels/networking/outbound/room/promotion/error"
	outpurchase "github.com/niflaot/pixels/networking/outbound/room/promotion/purchase"
	"go.uber.org/zap/zapcore"
)

// CommandName identifies room promotion commands.
const CommandName command.Name = "room.promotion"

// Action identifies one promotion operation.
type Action uint8

const (
	// PurchaseInfo requests player-owned eligible rooms.
	PurchaseInfo Action = iota
	// PurchaseAd buys or extends one room promotion.
	PurchaseAd
	// EditEvent changes active promotion copy.
	EditEvent
	// CancelEvent acknowledges the renderer composer without revoking purchased time.
	CancelEvent
	// Telemetry is an explicit NOOP because Nitro React has no shipped caller and no domain state consumes the renderer-only analytics packets.
	Telemetry
)

// Command contains one promotion packet request.
type Command struct {
	// Connection stores the request transport.
	Connection netconn.Context
	// Action identifies requested behavior.
	Action Action
	// Purchase stores room-ad purchase fields.
	Purchase PurchaseParams
	// Edit stores event edit fields.
	Edit EditParams
}

// CommandHandler executes promotion commands.
type CommandHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings maps authenticated connections.
	Bindings *binding.Registry
	// Promotions manages room promotions.
	Promotions *Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return CommandName }

// MarshalLogObject writes safe promotion command metadata.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddUint8("action", uint8(input.Action))
	encoder.AddInt64("room_id", input.Purchase.RoomID)
	return nil
}

// Handle executes one promotion command.
func (handler CommandHandler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	if envelope.Command.Action == Telemetry {
		return nil
	}
	if envelope.Command.Action == CancelEvent {
		packet, err := outcancel.Encode()
		return sendPacket(ctx, envelope.Command.Connection, packet, err)
	}
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	switch envelope.Command.Action {
	case PurchaseInfo:
		return handler.sendPurchaseInfo(ctx, envelope.Command.Connection, player)
	case PurchaseAd:
		params := envelope.Command.Purchase
		params.PlayerID = player.ID()
		params.PlayerName = player.Username()
		params.HasClub = catalogsession.HasClub(player)
		_, err = handler.Promotions.Purchase(ctx, params)
	case EditEvent:
		params := envelope.Command.Edit
		params.PlayerID = player.ID()
		_, err = handler.Promotions.Edit(ctx, params)
	default:
		return codec.ErrUnexpectedHeader
	}
	if err != nil {
		return handler.sendError(ctx, envelope.Command.Connection, err)
	}
	return nil
}

// sendPurchaseInfo projects eligible owned rooms and creation eligibility.
func (handler CommandHandler) sendPurchaseInfo(ctx context.Context, connection netconn.Context, player *playerlive.Player) error {
	rooms, active, err := handler.Promotions.EligibleRooms(ctx, player.ID())
	if err != nil {
		return err
	}
	entries := make([]outpurchase.Room, len(rooms))
	for index := range rooms {
		_, promoted := active[rooms[index].ID]
		entries[index] = outpurchase.Room{ID: int32(rooms[index].ID), Name: rooms[index].Name, Promoted: promoted}
	}
	packet, err := outpurchase.Encode(catalogsession.HasClub(player), entries)
	if err = sendPacket(ctx, connection, packet, err); err != nil {
		return err
	}
	packet, err = outeligibility.Encode(len(rooms) > 0, 0)
	return sendPacket(ctx, connection, packet, err)
}

// sendError maps expected domain rejections without disconnecting the session.
func (handler CommandHandler) sendError(ctx context.Context, connection netconn.Context, cause error) error {
	code := int32(1)
	if errors.Is(cause, ErrNoRights) {
		code = 2
	}
	if errors.Is(cause, ErrInvalidCopy) {
		code = 3
	}
	packet, err := outerror.Encode(code, "")
	return sendPacket(ctx, connection, packet, err)
}

// sendPacket writes one successfully encoded promotion packet.
func sendPacket(ctx context.Context, connection netconn.Context, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}
