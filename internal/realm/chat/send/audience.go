package send

import (
	"context"
	"strings"
	"time"

	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	whisperedevent "github.com/niflaot/pixels/internal/realm/chat/events/whispered"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/networking/codec"
)

// sendTalk routes normal speech to occupants within the configured squared radius.
func (service *Service) sendTalk(ctx context.Context, active *roomlive.Room, speaker roomlive.UnitSnapshot, packet codec.Packet) {
	distance := chatconfig.AudienceDistance(active.Snapshot().ChatDistance)
	maximum := distance * distance
	for _, presence := range active.Presences() {
		if withinDistance(speaker, presence.Unit, maximum) {
			service.sendPresence(ctx, presence.Occupant, packet)
		}
	}
}

// withinDistance reports whether two units fit one squared tile radius.
func withinDistance(source roomlive.UnitSnapshot, target roomlive.UnitSnapshot, maximum int64) bool {
	dx := int64(target.Position.Point.X) - int64(source.Position.Point.X)
	dy := int64(target.Position.Point.Y) - int64(source.Position.Point.Y)

	return dx*dx+dy*dy <= maximum
}

// sendAll routes one packet to every current room occupant.
func (service *Service) sendAll(ctx context.Context, active *roomlive.Room, packet codec.Packet) {
	for _, occupant := range active.Occupants() {
		service.sendPresence(ctx, occupant, packet)
	}
}

// sendWhisper routes private speech to sender, target, and authorized observers.
func (service *Service) sendWhisper(ctx context.Context, active *roomlive.Room, senderID int64, recipient string, packet codec.Packet, message string, censored bool, createdAt time.Time) error {
	presences := active.Presences()
	targetID := int64(0)
	for _, presence := range presences {
		if strings.EqualFold(presence.Occupant.Username, recipient) {
			targetID = presence.Occupant.PlayerID
			break
		}
	}
	if targetID == 0 {
		binding, found := service.bindings.FindByPlayer(senderID)
		if !found {
			return nil
		}
		connection, found := service.connections.Get(binding.ConnectionKind, binding.ConnectionID)
		if !found {
			return nil
		}

		return service.sendAlertConnection(ctx, connection, "chat.error.whisper_target")
	}
	for _, presence := range presences {
		deliver := presence.Occupant.PlayerID == senderID || presence.Occupant.PlayerID == targetID
		if !deliver {
			allowed, err := service.permissions.HasPermission(ctx, presence.Occupant.PlayerID, service.nodes.WhisperObserveAny)
			if err != nil {
				return err
			}
			deliver = allowed
		}
		if deliver {
			service.sendPresence(ctx, presence.Occupant, packet)
		}
	}
	service.publish(ctx, whisperedevent.Name, whisperedevent.Payload{RoomID: active.ID(), PlayerID: senderID, TargetPlayerID: targetID, Message: message, Censored: censored, CreatedAt: createdAt})

	return nil
}

// sendPresence performs one best-effort occupant delivery.
func (service *Service) sendPresence(ctx context.Context, occupant roomlive.Occupant, packet codec.Packet) {
	connection, found := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
	if found {
		_ = connection.Send(ctx, packet)
	}
}
