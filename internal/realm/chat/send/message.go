package send

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	muterejected "github.com/niflaot/pixels/internal/realm/chat/events/muterejected"
	roomcontrol "github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Kind names one protocol room chat mode.
type Kind uint8

const (
	// KindTalk delivers normal distance-limited speech.
	KindTalk Kind = iota + 1
	// KindShout delivers speech to the full room.
	KindShout
	// KindWhisper delivers private speech to one room occupant.
	KindWhisper
)

// Request contains one decoded room chat request.
type Request struct {
	// Handler stores the source connection context.
	Handler netconn.Context
	// Kind stores the protocol chat mode.
	Kind Kind
	// Message stores submitted text.
	Message string
	// Recipient stores a whisper target username.
	Recipient string
}

// Handle validates and delivers one room chat request.
func (service *Service) Handle(ctx context.Context, request Request) error {
	player, roomID, err := roomcontrol.Actor(request.Handler, service.bindings, service.players)
	if err != nil {
		return err
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return roomlive.ErrRoomNotFound
	}
	message := sanitize(request.Message)
	if message == "" {
		return nil
	}
	if remaining, muted := active.RemainingMute(player.ID(), service.now()); muted {
		service.publishMute(ctx, roomID, player.ID(), "mute")
		return service.sendMute(ctx, request.Handler, remaining)
	}
	privileged, err := service.muteAllBypass(ctx, active, player.ID())
	if err != nil {
		return err
	}
	if active.MuteAll() && !privileged {
		service.publishMute(ctx, roomID, player.ID(), "mute_all")
		return service.sendMute(ctx, request.Handler, 0)
	}
	if rejected, err := service.flooded(ctx, player.ID(), active.Snapshot().ChatProtection); err != nil {
		return err
	} else if rejected {
		return service.sendFlood(ctx, request.Handler, service.config.Tier(active.Snapshot().ChatProtection).Window)
	}
	if utf8.RuneCountInString(message) > service.config.MaxMessageRunes {
		unlimited, permissionErr := service.permissions.HasPermission(ctx, player.ID(), service.nodes.LengthUnlimited)
		if permissionErr != nil {
			return permissionErr
		}
		if !unlimited {
			return service.sendAlert(ctx, request.Handler, "chat.error.too_long")
		}
	}
	message, censored, err := service.censor(ctx, roomID, player.ID(), message)
	if err != nil {
		return err
	}

	return service.deliver(ctx, active, player, request, message, censored)
}

// flooded increments and evaluates one player's cross-room burst counter.
func (service *Service) flooded(ctx context.Context, playerID int64, protection int16) (bool, error) {
	immune, err := service.permissions.HasPermission(ctx, playerID, service.nodes.FloodImmune)
	if err != nil || immune {
		return false, err
	}
	tier := service.config.Tier(protection)
	count, err := service.counter.Increment(ctx, "chat:flood:"+strconv.FormatInt(playerID, 10), tier.Window)

	return count > tier.MaxMessages, err
}

// censor applies global and room filters unless the speaker is exempt.
func (service *Service) censor(ctx context.Context, roomID int64, playerID int64, message string) (string, bool, error) {
	immune, err := service.permissions.HasPermission(ctx, playerID, service.nodes.FilterImmune)
	if err != nil || immune {
		return message, false, err
	}
	message, globalChanged := service.globalFilter.Censor(message)
	message, roomChanged, err := service.roomFilter.Censor(ctx, roomID, message)

	return message, globalChanged || roomChanged, err
}

// muteAllBypass reports whether a speaker may talk through room mute-all.
func (service *Service) muteAllBypass(ctx context.Context, active *roomlive.Room, playerID int64) (bool, error) {
	if !active.MuteAll() || active.HasRights(playerID) {
		return true, nil
	}
	allowed, err := service.permissions.HasPermission(ctx, playerID, service.nodes.ModerationAnyMute)
	if err != nil || allowed {
		return allowed, err
	}

	return service.permissions.HasPermission(ctx, playerID, service.nodes.ModerationOwnMute)
}

// sanitize removes protocol-hostile line breaks and surrounding whitespace.
func sanitize(message string) string {
	message = strings.TrimSpace(message)
	if strings.IndexAny(message, "\r\n") < 0 {
		return message
	}

	return strings.Map(func(value rune) rune {
		if value == '\r' || value == '\n' {
			return ' '
		}
		return value
	}, message)
}

// publishMute emits non-critical mute rejection telemetry.
func (service *Service) publishMute(ctx context.Context, roomID int64, playerID int64, reason string) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: muterejected.Name, Payload: muterejected.Payload{RoomID: roomID, PlayerID: playerID, Reason: reason}})
	}
}
