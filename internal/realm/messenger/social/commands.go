// Package social handles messenger room invites, following, and private chat.
package social

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	invitesent "github.com/niflaot/pixels/internal/realm/messenger/social/events/invitesent"
	messagesent "github.com/niflaot/pixels/internal/realm/messenger/social/events/messagesent"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfollowerror "github.com/niflaot/pixels/networking/outbound/messenger/social/followerror"
	outinvite "github.com/niflaot/pixels/networking/outbound/messenger/social/invite"
	outinviteerror "github.com/niflaot/pixels/networking/outbound/messenger/social/inviteerror"
	outprivate "github.com/niflaot/pixels/networking/outbound/messenger/social/privatechat"
	outprivateerror "github.com/niflaot/pixels/networking/outbound/messenger/social/privateerror"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	"github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// InviteName identifies room invitations.
	InviteName command.Name = "messenger.social.invite"
	// FollowName identifies friend following.
	FollowName command.Name = "messenger.social.follow"
	// PrivateName identifies private messages.
	PrivateName command.Name = "messenger.social.private_chat"
	// PrivacyName identifies room-invite privacy changes.
	PrivacyName command.Name = "messenger.social.privacy"
)

// InviteCommand requests one room invitation batch.
type InviteCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// PlayerIDs stores invitation recipients.
	PlayerIDs []int64
	// Message stores invitation text.
	Message string
}

// CommandName returns the invite command identity.
func (InviteCommand) CommandName() command.Name { return InviteName }

// TargetCommand carries one target and optional message.
type TargetCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// PlayerID identifies the target.
	PlayerID int64
	// Message stores private-message text.
	Message string
	// Name stores the command identity.
	Name command.Name
}

// PrivacyCommand changes Nitro's native room-invite preference.
type PrivacyCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// RoomInvitesBlocked stores the requested privacy state.
	RoomInvitesBlocked bool
}

// CommandName returns the privacy command identity.
func (PrivacyCommand) CommandName() command.Name { return PrivacyName }

// CommandName returns the configured target command identity.
func (input TargetCommand) CommandName() command.Name { return input.Name }

// Handler executes messenger social commands.
type Handler struct {
	// Messenger stores messenger behavior.
	Messenger *messengerservice.Service
	// Delivery sends packets and resolves actors.
	Delivery *delivery.Sender
	// Events publishes completed actions.
	Events bus.Publisher
	// Translations localizes hotel-facing feedback.
	Translations i18n.Translator
	// Directory resolves exact usernames for Navigator visit-user requests.
	Directory playerservice.Finder
}

// HandleVisitUser resolves a username and reuses the authoritative follow policy.
func (handler Handler) HandleVisitUser(ctx context.Context, connection netconn.Context, username string) error {
	if handler.Directory == nil {
		return nil
	}
	record, found, err := handler.Directory.FindByUsername(ctx, username)
	if err != nil || !found {
		return err
	}
	return handler.HandleFollow(ctx, TargetCommand{Connection: connection, PlayerID: record.Player.ID, Name: FollowName})
}

// HandleInvite validates and delivers one room invitation batch.
func (handler Handler) HandleInvite(ctx context.Context, input InviteCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	result, err := handler.Messenger.Invite(ctx, actorID, input.PlayerIDs, input.Message)
	if err != nil {
		return err
	}
	packet, err := outinvite.Encode(actorID, result.Message)
	if err != nil {
		return err
	}
	for _, playerID := range result.Delivered {
		_, _ = handler.Delivery.Send(ctx, playerID, packet)
	}
	if len(result.NotFriends) > 0 {
		failed, encodeErr := outinviteerror.Encode(1, result.NotFriends)
		if encodeErr != nil {
			return encodeErr
		}
		if err = input.Connection.Send(ctx, failed); err != nil {
			return err
		}
	}
	for _, playerID := range result.Delivered {
		handler.publish(ctx, invitesent.Name, invitesent.Payload{FromPlayerID: actorID, ToPlayerID: playerID, RoomID: result.RoomID})
	}
	return nil
}

// HandleFollow forwards the actor or returns Nitro's native follow error.
func (handler Handler) HandleFollow(ctx context.Context, input TargetCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	result, err := handler.Messenger.Follow(ctx, actorID, input.PlayerID)
	if err != nil {
		packet, encodeErr := outfollowerror.Encode(followErrorCode(err))
		if encodeErr != nil {
			return encodeErr
		}
		return input.Connection.Send(ctx, packet)
	}
	if result.SameRoom {
		message := handler.Translations.Default("messenger.follow.same_room")
		packet, encodeErr := bubblealert.Encode("messenger.follow.same_room", message, bubblealert.WithDisplayBubble())
		if encodeErr != nil {
			return encodeErr
		}
		return input.Connection.Send(ctx, packet)
	}
	packet, err := outforward.Encode(int32(result.RoomID))
	if err != nil {
		return err
	}
	return input.Connection.Send(ctx, packet)
}

// HandlePrivate validates and delivers one private message.
func (handler Handler) HandlePrivate(ctx context.Context, input TargetCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	result, err := handler.Messenger.SendPrivate(ctx, actorID, input.PlayerID, input.Message)
	if err != nil {
		packet, encodeErr := outprivateerror.Encode(1, input.PlayerID, "")
		if encodeErr != nil {
			return encodeErr
		}
		return input.Connection.Send(ctx, packet)
	}
	if result.Throttled {
		return nil
	}
	if result.Deliver {
		packet, encodeErr := outprivate.Encode(actorID, result.Message, 0)
		if encodeErr != nil {
			return encodeErr
		}
		_, err = handler.Delivery.Send(ctx, input.PlayerID, packet)
	}
	handler.publish(ctx, messagesent.Name, messagesent.Payload{FromPlayerID: actorID, ToPlayerID: input.PlayerID, Delivered: result.Deliver})
	return err
}

// HandlePrivacy persists Nitro's native room-invite blocking preference.
func (handler Handler) HandlePrivacy(ctx context.Context, input PrivacyCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	_, err = handler.Messenger.SetRoomInvites(ctx, actorID, input.RoomInvitesBlocked)
	return err
}

// followErrorCode maps follow failures to Nitro's native codes.
func followErrorCode(failure error) int32 {
	switch {
	case errors.Is(failure, messengerservice.ErrFriendOffline):
		return 1
	case errors.Is(failure, messengerservice.ErrFriendNotInRoom):
		return 2
	case errors.Is(failure, messengerservice.ErrFollowingBlocked):
		return 3
	default:
		return 0
	}
}

// publish emits one accepted social action.
func (handler Handler) publish(ctx context.Context, name bus.Name, payload any) {
	if handler.Events != nil {
		_ = handler.Events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}
