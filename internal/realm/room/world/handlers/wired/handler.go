// Package wired adapts Nitro WIRED editor packets into room commands.
package wired

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	wiredcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/wired"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaction "github.com/niflaot/pixels/networking/inbound/furniture/wired/action/save"
	incondition "github.com/niflaot/pixels/networking/inbound/furniture/wired/condition/save"
	inopen "github.com/niflaot/pixels/networking/inbound/furniture/wired/open"
	insnapshot "github.com/niflaot/pixels/networking/inbound/furniture/wired/snapshot/apply"
	intrigger "github.com/niflaot/pixels/networking/inbound/furniture/wired/trigger/save"
	"go.uber.org/zap"
)

// New creates a grouped WIRED packet handler.
func New(handler wiredcmd.Handler, log *zap.Logger) netconn.Handler {
	return func(connection netconn.Context, packet codec.Packet) error {
		ctx := context.Background()
		metadata := command.Metadata{ConnectionID: string(connection.ConnectionID)}
		switch packet.Header {
		case inopen.Header:
			payload, err := inopen.Decode(packet)
			if err != nil {
				return err
			}
			return handler.HandleOpen(ctx, command.Envelope[wiredcmd.OpenCommand]{Command: wiredcmd.OpenCommand{Handler: connection, ItemID: int64(payload.ItemID)}, Metadata: metadata})
		case intrigger.Header:
			payload, err := intrigger.Decode(packet)
			if err != nil {
				return err
			}
			return handler.HandleSave(ctx, command.Envelope[wiredcmd.SaveCommand]{Command: wiredcmd.SaveCommand{Handler: connection, Family: wiredcmd.TriggerFamily, Settings: payload}, Metadata: metadata})
		case inaction.Header:
			payload, err := inaction.Decode(packet)
			if err != nil {
				return err
			}
			return handler.HandleSave(ctx, command.Envelope[wiredcmd.SaveCommand]{Command: wiredcmd.SaveCommand{Handler: connection, Family: wiredcmd.EffectFamily, Settings: payload}, Metadata: metadata})
		case incondition.Header:
			payload, err := incondition.Decode(packet)
			if err != nil {
				return err
			}
			return handler.HandleSave(ctx, command.Envelope[wiredcmd.SaveCommand]{Command: wiredcmd.SaveCommand{Handler: connection, Family: wiredcmd.ConditionFamily, Settings: payload}, Metadata: metadata})
		case insnapshot.Header:
			payload, err := insnapshot.Decode(packet)
			if err != nil {
				return err
			}
			return handler.HandleSnapshot(ctx, command.Envelope[wiredcmd.SnapshotCommand]{Command: wiredcmd.SnapshotCommand{Handler: connection, ItemID: int64(payload.ItemID)}, Metadata: metadata})
		default:
			if log != nil {
				log.Debug("unsupported WIRED packet", zap.Uint16("header", packet.Header))
			}
			return codec.ErrUnexpectedHeader
		}
	}
}

// Register registers every WIRED editor packet header.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inopen.Header, handler)
	_ = registry.Register(intrigger.Header, handler)
	_ = registry.Register(inaction.Header, handler)
	_ = registry.Register(incondition.Header, handler)
	_ = registry.Register(insnapshot.Header, handler)
}
