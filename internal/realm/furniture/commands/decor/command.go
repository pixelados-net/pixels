// Package decor coordinates room decorator furniture commands.
package decor

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	furnitureaccess "github.com/niflaot/pixels/internal/realm/furniture/access"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// decorationSoftError reports malformed or stale decorator requests that should not disconnect a client.
func decorationSoftError(err error) bool {
	return errors.Is(err, roomdecor.ErrInvalidSurface) ||
		errors.Is(err, roomdecor.ErrInvalidSurfaceValue) ||
		errors.Is(err, roomdecor.ErrInvalidWallPosition) ||
		errors.Is(err, roomdecor.ErrInvalidDimmerPreset) ||
		errors.Is(err, roomdecor.ErrDecorationUnavailable)
}

// Name identifies grouped room decorator commands.
const Name command.Name = "furniture.decor"

const (
	// bubbleKeyFurniturePlacementError identifies Nitro's furniture feedback bubble.
	bubbleKeyFurniturePlacementError = "furni_placement_error"
	// noRightsTranslationKey identifies localized furniture authorization feedback.
	noRightsTranslationKey i18n.Key = "session.bubble.furniture.no_rights"
)

// Kind identifies one decorator operation.
type Kind uint8

const (
	// KindPostItPlace places a post-it.
	KindPostItPlace Kind = iota + 1
	// KindPostItSave saves initial post-it data.
	KindPostItSave
	// KindPostItGet requests post-it data.
	KindPostItGet
	// KindPostItSet edits post-it data.
	KindPostItSet
	// KindSurface applies a room surface consumable.
	KindSurface
	// KindDimmerSettings requests presets.
	KindDimmerSettings
	// KindDimmerSave saves a preset.
	KindDimmerSave
	// KindDimmerToggle toggles the mood light.
	KindDimmerToggle
	// KindMannequinLook saves the actor's clothing.
	KindMannequinLook
	// KindMannequinName saves an outfit name.
	KindMannequinName
	// KindTonerApply saves background toner HSL values.
	KindTonerApply
)

// Command contains the union of decorator packet fields.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Kind identifies the operation.
	Kind Kind
	// ItemID identifies furniture.
	ItemID int64
	// WallPosition stores modern Nitro wall coordinates.
	WallPosition string
	// Color stores post-it or dimmer color.
	Color string
	// Text stores post-it text or mannequin name.
	Text string
	// PresetID identifies a dimmer slot.
	PresetID int32
	// Type stores a dimmer preset type.
	Type int32
	// First stores brightness or toner hue.
	First int32
	// Second stores toner saturation.
	Second int32
	// Third stores toner lightness.
	Third int32
	// Apply reports whether a dimmer preset becomes active.
	Apply bool
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handler handles grouped decorator commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores authenticated session bindings.
	Bindings *binding.Registry
	// Furniture reads furniture and definitions.
	Furniture furnitureservice.Manager
	// States persists compare-and-swap furniture state.
	States furnitureservice.StateUpdater
	// Decoration manages room surface and dimmer persistence.
	Decoration *roomdecor.Service
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Permissions resolves global furniture authority.
	Permissions permissionservice.Checker
	// Connections stores live transports.
	Connections *netconn.Registry
	// WordFilters applies room-local text filtering.
	WordFilters roomwordfilter.Manager
	// GlobalFilter applies hotel-wide text filtering.
	GlobalFilter *chatfilter.Service
	// PlayerAdmin persists mannequin look application.
	PlayerAdmin playerservice.AdminManager
	// Translations resolves hotel-facing decorator feedback.
	Translations i18n.Translator
	// Log records rejected decorator requests.
	Log *zap.Logger
}

// Handle handles one decorator command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, active, roomID, err := handler.actor(envelope.Command.Handler)
	if err != nil || active == nil {
		return err
	}
	switch envelope.Command.Kind {
	case KindPostItPlace, KindPostItSave, KindPostItGet, KindPostItSet:
		return handler.handlePostIt(ctx, player, active, roomID, envelope.Command)
	case KindSurface:
		return handler.handleSurface(ctx, player, active, roomID, envelope.Command)
	case KindDimmerSettings, KindDimmerSave, KindDimmerToggle:
		return handler.handleDimmer(ctx, player, active, roomID, envelope.Command)
	case KindMannequinLook, KindMannequinName:
		return handler.handleMannequin(ctx, player, active, roomID, envelope.Command)
	case KindTonerApply:
		return handler.handleToner(ctx, player, active, roomID, envelope.Command)
	default:
		return nil
	}
}

// actor resolves authenticated room state.
func (handler Handler) actor(connection netconn.Context) (*playerlive.Player, *roomlive.Room, int64, error) {
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return nil, nil, 0, err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return player, nil, 0, nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return player, nil, roomID, nil
	}
	return player, active, roomID, nil
}

// broadcast sends one packet to every current occupant.
func (handler Handler) broadcast(ctx context.Context, active *roomlive.Room, packet codec.Packet) error {
	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// furnitureState changes one placed item using compare-and-swap semantics.
func (handler Handler) furnitureState(ctx context.Context, item furnituremodel.Item, roomID int64, value string) (furnituremodel.Item, error) {
	return handler.States.UpdateState(ctx, furnitureservice.StateParams{ItemID: item.ID, RoomID: roomID, Expected: item.ExtraData, Next: value})
}

// canManage reports whether the actor can alter room furniture.
func (handler Handler) canManage(ctx context.Context, active *roomlive.Room, playerID int64) (bool, error) {
	return furnitureaccess.CanManage(ctx, handler.Permissions, active, playerID)
}

// sendNoRights sends localized decorator authorization feedback.
func (handler Handler) sendNoRights(ctx context.Context, command Command) error {
	if handler.Log != nil {
		handler.Log.Warn("post-it placement rejected", zap.Int64("item_id", command.ItemID), zap.Error(roomlive.ErrNoFurnitureRights))
	}
	message := string(noRightsTranslationKey)
	if handler.Translations != nil {
		message = handler.Translations.Default(noRightsTranslationKey)
	}
	packet, err := outbubble.Encode(bubbleKeyFurniturePlacementError, message, outbubble.WithDisplayBubble())
	if err != nil {
		return err
	}

	return command.Handler.Send(ctx, packet)
}

// broadcastFloorUpdate sends specialized object data for one changed floor item.
func (handler Handler) broadcastFloorUpdate(ctx context.Context, active *roomlive.Room, item furnituremodel.Item, definition furnituremodel.Definition) error {
	if item.X == nil || item.Y == nil || item.Z == nil {
		return nil
	}
	record := outupdate.FloorItem{ID: item.ID, SpriteID: definition.SpriteID, X: *item.X, Y: *item.Y, Rotation: int(item.Rotation), Z: projection.FurnitureHeightValue(*item.Z), ExtraHeight: projection.ExtraHeightValue(definition), ExtraData: item.ExtraData, OwnerID: item.OwnerPlayerID}
	record.Data = projection.SpecializedObjectData(definition.InteractionType, item.ExtraData)
	packet, err := outupdate.Encode(record)
	if err != nil {
		return err
	}

	return handler.broadcast(ctx, active, packet)
}
