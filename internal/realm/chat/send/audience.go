package send

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	whisperedevent "github.com/niflaot/pixels/internal/realm/chat/events/whispered"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	outwhisper "github.com/niflaot/pixels/networking/outbound/chat/whisper"
	"github.com/niflaot/pixels/pkg/i18n"
)

// sendTalk routes normal speech to occupants within the configured squared radius.
func (service *Service) sendTalk(ctx context.Context, active *roomlive.Room, speaker roomlive.UnitSnapshot, packet codec.Packet) {
	distance := chatconfig.AudienceDistance(active.Snapshot().ChatDistance)
	maximum := distance * distance
	for _, presence := range active.Presences() {
		if withinDistance(speaker, presence.Unit, maximum) && service.canReceive(presence.Occupant.PlayerID, speaker.PlayerID) {
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
func (service *Service) sendAll(ctx context.Context, active *roomlive.Room, senderID int64, packet codec.Packet) {
	for _, occupant := range active.Occupants() {
		if service.canReceive(occupant.PlayerID, senderID) {
			service.sendPresence(ctx, occupant, packet)
		}
	}
}

// sendWhisper routes private speech to sender, target, and authorized observers.
func (service *Service) sendWhisper(ctx context.Context, active *roomlive.Room, senderID int64, unitID int32, recipient string, packet codec.Packet, message string, styleID int32, censored bool, createdAt time.Time) error {
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
	var observerPacket codec.Packet
	observerReady := false
	for _, presence := range presences {
		if presence.Occupant.PlayerID == senderID || presence.Occupant.PlayerID == targetID {
			if service.canReceive(presence.Occupant.PlayerID, senderID) {
				service.sendPresence(ctx, presence.Occupant, packet)
			}
			continue
		}
		allowed, permissionErr := service.permissions.HasPermission(ctx, presence.Occupant.PlayerID, service.nodes.WhisperObserveAny)
		if permissionErr != nil {
			return permissionErr
		}
		if allowed && service.canReceive(presence.Occupant.PlayerID, senderID) {
			if !observerReady {
				encoded, encodeErr := service.observerWhisper(unitID, recipient, message, styleID)
				if encodeErr != nil {
					return encodeErr
				}
				observerPacket = encoded
				observerReady = true
			}
			service.sendPresence(ctx, presence.Occupant, observerPacket)
		}
	}
	service.publish(ctx, whisperedevent.Name, whisperedevent.Payload{RoomID: active.ID(), PlayerID: senderID, TargetPlayerID: targetID, Message: message, Censored: censored, CreatedAt: createdAt})

	return nil
}

// canReceive reports whether one live recipient accepts communication from a sender.
func (service *Service) canReceive(recipientID int64, senderID int64) bool {
	if recipientID == senderID {
		return true
	}
	recipient, found := service.players.Find(recipientID)
	return !found || !recipient.IsIgnoring(senderID)
}

// observerWhisper creates the recipient-aware packet shown to authorized observers.
func (service *Service) observerWhisper(unitID int32, recipient string, message string, styleID int32) (codec.Packet, error) {
	visible := "To " + recipient + ": " + message
	if service.translations != nil {
		visible = service.translations.Default("chat.whisper.observer", i18n.Params{"recipient": recipient, "message": message})
	}

	return outwhisper.Encode(unitID, visible, gesture(message), styleID, int32(utf8.RuneCountInString(visible)))
}

// sendPresence performs one best-effort occupant delivery.
func (service *Service) sendPresence(ctx context.Context, occupant roomlive.Occupant, packet codec.Packet) {
	connection, found := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
	if found {
		_ = connection.Send(ctx, packet)
	}
}
