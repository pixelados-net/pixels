package send

import (
	"context"
	"strings"
	"unicode/utf8"

	shoutedevent "github.com/niflaot/pixels/internal/realm/chat/events/shouted"
	talkedevent "github.com/niflaot/pixels/internal/realm/chat/events/talked"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	outshout "github.com/niflaot/pixels/networking/outbound/chat/shout"
	outtalk "github.com/niflaot/pixels/networking/outbound/chat/talk"
	outwhisper "github.com/niflaot/pixels/networking/outbound/chat/whisper"
	"github.com/niflaot/pixels/pkg/bus"
)

// deliver encodes and routes one validated message.
func (service *Service) deliver(ctx context.Context, active *roomlive.Room, player *playerlive.Player, request Request, message string, censored bool) error {
	unit, found := active.Unit(player.ID())
	if !found {
		return roomlive.ErrUnitNotFound
	}
	styleID := player.Snapshot().BubbleStyle
	length := int32(utf8.RuneCountInString(message))
	packet, err := encode(request.Kind, int32(unit.UnitID), message, styleID, length)
	if err != nil {
		return err
	}
	createdAt := service.now()
	switch request.Kind {
	case KindTalk:
		service.sendTalk(ctx, active, unit, packet)
		service.publish(ctx, talkedevent.Name, talkedevent.Payload{RoomID: active.ID(), PlayerID: player.ID(), Message: message, Censored: censored, CreatedAt: createdAt})
	case KindShout:
		service.sendAll(ctx, active, player.ID(), packet)
		service.publish(ctx, shoutedevent.Name, shoutedevent.Payload{RoomID: active.ID(), PlayerID: player.ID(), Message: message, Censored: censored, CreatedAt: createdAt})
	case KindWhisper:
		return service.sendWhisper(ctx, active, player.ID(), int32(unit.UnitID), request.Recipient, packet, message, styleID, censored, createdAt)
	}

	return nil
}

// encode creates the protocol packet for one chat mode.
func encode(kind Kind, unitID int32, message string, styleID int32, length int32) (codec.Packet, error) {
	switch kind {
	case KindTalk:
		return outtalk.Encode(unitID, message, gesture(message), styleID, length)
	case KindShout:
		return outshout.Encode(unitID, message, gesture(message), styleID, length)
	default:
		return outwhisper.Encode(unitID, message, gesture(message), styleID, length)
	}
}

// gesture returns the Nitro gesture code inferred from common emoticons.
func gesture(message string) int32 {
	switch {
	case strings.Contains(message, ":)") || strings.Contains(message, ":D"):
		return 1
	case strings.Contains(message, ":("):
		return 2
	case strings.Contains(message, ":o") || strings.Contains(message, ":O"):
		return 3
	case strings.Contains(message, ":@"):
		return 4
	default:
		return 0
	}
}

// publish emits one delivered message event without affecting live delivery.
func (service *Service) publish(ctx context.Context, name bus.Name, payload any) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}
