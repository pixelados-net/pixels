package wired

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	actionpacket "github.com/niflaot/pixels/networking/outbound/furniture/wired/action/definition"
	conditionpacket "github.com/niflaot/pixels/networking/outbound/furniture/wired/condition/definition"
	openedpacket "github.com/niflaot/pixels/networking/outbound/furniture/wired/opened"
	triggerpacket "github.com/niflaot/pixels/networking/outbound/furniture/wired/trigger/definition"
)

// HandleOpen validates and opens one WIRED editor.
func (handler Handler) HandleOpen(ctx context.Context, envelope command.Envelope[OpenCommand]) error {
	player, active, roomID, err := handler.actor(envelope.Command.Handler)
	if err != nil {
		return err
	}
	if err = handler.authorize(ctx, player.ID(), active); err != nil {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.no_rights")
	}
	stored, found, err := handler.Store.Find(ctx, roomID, envelope.Command.ItemID)
	if err != nil || !found {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.target_missing")
	}
	descriptor, found := handler.Registry.Resolve(stored.Interaction)
	if !found || (descriptor.Family != registry.FamilyTrigger && descriptor.Family != registry.FamilyEffect && descriptor.Family != registry.FamilyCondition) {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.unsupported_editor")
	}
	opened, err := openedpacket.Encode(stored.ItemID)
	if err != nil {
		return err
	}
	if err = envelope.Command.Handler.Send(ctx, opened); err != nil {
		return err
	}
	selected := selectedIDs(stored)
	selection := descriptor.Selection != registry.SelectionNone
	conflicts := handler.Engine.Conflicts(roomID, stored.ItemID)
	switch descriptor.Family {
	case registry.FamilyTrigger:
		packet, encodeErr := triggerpacket.Encode(selection, int32(handler.Config.Normalize().MaxSelection), selected, stored.SpriteID, stored.ItemID, stored.StringParam, stored.IntParams, stored.SelectionMode, descriptor.ClientCode, conflicts)
		if encodeErr != nil {
			return encodeErr
		}
		return envelope.Command.Handler.Send(ctx, packet)
	case registry.FamilyEffect:
		packet, encodeErr := actionpacket.Encode(selection, int32(handler.Config.Normalize().MaxSelection), selected, stored.SpriteID, stored.ItemID, stored.StringParam, stored.IntParams, stored.SelectionMode, descriptor.ClientCode, stored.DelayPulses, conflicts)
		if encodeErr != nil {
			return encodeErr
		}
		return envelope.Command.Handler.Send(ctx, packet)
	default:
		packet, encodeErr := conditionpacket.Encode(selection, int32(handler.Config.Normalize().MaxSelection), selected, stored.SpriteID, stored.ItemID, stored.StringParam, stored.IntParams, stored.SelectionMode, descriptor.ClientCode)
		if encodeErr != nil {
			return encodeErr
		}
		return envelope.Command.Handler.Send(ctx, packet)
	}
}

// selectedIDs copies protocol target identifiers.
func selectedIDs(stored record.Config) []int64 {
	result := make([]int64, len(stored.Targets))
	for index, target := range stored.Targets {
		result[index] = target.ItemID
	}
	return result
}
