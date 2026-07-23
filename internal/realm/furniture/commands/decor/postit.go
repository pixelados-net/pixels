package decor

import (
	"context"
	"strings"
	"unicode/utf8"

	postitevent "github.com/niflaot/pixels/internal/realm/furniture/events/postitplaced"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	outplaced "github.com/niflaot/pixels/networking/outbound/inventory/furniture/postitplaced"
	outinvremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outitemdata "github.com/niflaot/pixels/networking/outbound/room/furniture/itemdata"
	outopen "github.com/niflaot/pixels/networking/outbound/room/furniture/postitopen"
	outwalladd "github.com/niflaot/pixels/networking/outbound/room/furniture/walladd"
	outwallupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/wallupdate"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// maximumPostItText stores the maximum persisted Unicode text length.
	maximumPostItText = 512
)

// handlePostIt handles placement, initial save, reads, and edits.
func (handler Handler) handlePostIt(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, command Command) error {
	item, found, err := handler.Furniture.FindItemByID(ctx, command.ItemID)
	if err != nil || !found {
		return err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found || definition.InteractionType != "postit" {
		return err
	}
	switch command.Kind {
	case KindPostItPlace:
		return handler.placePostIt(ctx, player, active, roomID, item, definition, command)
	case KindPostItGet:
		if item.RoomID == nil || *item.RoomID != roomID {
			return nil
		}
		packet, encodeErr := outitemdata.Encode(item.ID, postItData(item.ExtraData))
		if encodeErr != nil {
			return encodeErr
		}
		return command.Handler.Send(ctx, packet)
	case KindPostItSave, KindPostItSet:
		return handler.savePostIt(ctx, player, active, roomID, item, definition, command)
	default:
		return nil
	}
}

// placePostIt moves an inventory post-it to the wall and opens its editor.
func (handler Handler) placePostIt(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, item furnituremodel.Item, definition furnituremodel.Definition, command Command) error {
	allowed, err := handler.canManage(ctx, active, player.ID())
	if err != nil {
		return err
	}
	if !allowed && !hasStickyPole(active) {
		return handler.sendNoRights(ctx, command)
	}
	if err = handler.Decoration.PlacePostIt(ctx, item.ID, player.ID(), roomID, command.WallPosition); err != nil {
		if decorationSoftError(err) {
			return nil
		}
		return err
	}
	if handler.Events != nil {
		ownerID := active.Snapshot().OwnerPlayerID
		_ = handler.Events.Publish(ctx, bus.Event{Name: postitevent.Name, Payload: postitevent.Payload{PlayerID: player.ID(), RoomOwnerID: ownerID, RoomID: roomID, ItemID: item.ID}})
	}
	packet, err := outinvremove.Encode(item.ID)
	if err != nil {
		return err
	}
	if err = command.Handler.Send(ctx, packet); err != nil {
		return err
	}
	packet, err = outwalladd.Encode(outwalladd.Item{ID: item.ID, SpriteID: definition.SpriteID, WallPosition: command.WallPosition, ExtraData: roomdecor.DefaultPostItColor, OwnerID: item.OwnerPlayerID, OwnerName: player.Username()})
	if err != nil {
		return err
	}
	if err = handler.broadcast(ctx, active, packet); err != nil {
		return err
	}
	packet, err = handler.postItPlacedPacket(ctx, player.ID(), item)
	if err != nil {
		return err
	}
	if err = command.Handler.Send(ctx, packet); err != nil {
		return err
	}
	packet, err = outopen.Encode(item.ID, command.WallPosition)
	if err != nil {
		return err
	}
	return command.Handler.Send(ctx, packet)
}

// savePostIt validates filtered content and broadcasts the changed wall item.
func (handler Handler) savePostIt(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, item furnituremodel.Item, definition furnituremodel.Definition, command Command) error {
	if item.RoomID == nil || *item.RoomID != roomID || item.WallPosition == nil {
		return nil
	}
	allowed, err := handler.canManage(ctx, active, player.ID())
	if err != nil {
		return err
	}
	if !allowed && item.OwnerPlayerID != player.ID() {
		return nil
	}
	color := strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(command.Color), "#"))
	if !validPostItColor(color) || utf8.RuneCountInString(command.Text) > maximumPostItText {
		return nil
	}
	text := strings.ReplaceAll(command.Text, "\t", "")
	if handler.GlobalFilter != nil {
		text, _ = handler.GlobalFilter.Censor(text)
	}
	if handler.WordFilters != nil {
		text, _, err = handler.WordFilters.Censor(ctx, roomID, text)
		if err != nil {
			return err
		}
	}
	updated, err := handler.furnitureState(ctx, item, roomID, color+" "+text)
	if err != nil {
		return err
	}
	packet, err := outwallupdate.Encode(updated.ID, definition.SpriteID, *updated.WallPosition, roomdecor.PostItColor(updated.ExtraData), 0, updated.OwnerPlayerID)
	if err != nil {
		return err
	}
	if err = handler.broadcast(ctx, active, packet); err != nil {
		return err
	}
	return nil
}

// postItPlacedPacket reports the remaining inventory count after durable placement.
func (handler Handler) postItPlacedPacket(ctx context.Context, playerID int64, item furnituremodel.Item) (codec.Packet, error) {
	inventory, err := handler.Furniture.ListInventory(ctx, playerID)
	if err != nil {
		return codec.Packet{}, err
	}
	left := int32(0)
	for _, candidate := range inventory {
		if candidate.DefinitionID == item.DefinitionID {
			left++
		}
	}

	return outplaced.Encode(item.ID, left)
}

// postItData supplies the default yellow color for empty legacy rows.
func postItData(value string) string {
	if len(value) < 6 {
		return roomdecor.DefaultPostItColor
	}
	return value
}

// validPostItColor reports whether Nitro supports the requested note color.
func validPostItColor(value string) bool {
	return value == "9CCEFF" || value == "FF9CFF" || value == "9CFF9C" || value == roomdecor.DefaultPostItColor
}

// hasStickyPole reports whether a placed sticky pole permits guest note placement.
func hasStickyPole(active *roomlive.Room) bool {
	for _, item := range active.FurnitureItems() {
		if item.Definition.InteractionType == "sticky_pole" {
			return true
		}
	}
	return false
}
