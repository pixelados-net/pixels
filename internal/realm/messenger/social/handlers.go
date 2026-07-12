package social

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	infollow "github.com/niflaot/pixels/networking/inbound/messenger/social/follow"
	ininvite "github.com/niflaot/pixels/networking/inbound/messenger/social/invite"
	inprivate "github.com/niflaot/pixels/networking/inbound/messenger/social/privatechat"
	inprivacy "github.com/niflaot/pixels/networking/inbound/user/settings/roominvites"
	"go.uber.org/zap"
)

// RegisterHandlers registers messenger social packet adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler, log *zap.Logger) {
	inviteDispatcher, _ := command.NewDispatcher(inviteHandler{handler: handler})
	inviteDispatcher.WithLogger(log)
	_ = registry.Register(ininvite.Header, func(connection netconn.Context, packet codec.Packet) error {
		payload, err := ininvite.Decode(packet)
		if err != nil {
			return err
		}
		return inviteDispatcher.Dispatch(context.Background(), command.Envelope[InviteCommand]{Command: InviteCommand{Connection: connection, PlayerIDs: payload.PlayerIDs, Message: payload.Message}, Metadata: metadata(connection)})
	})
	registerTarget(registry, infollow.Header, FollowName, func(packet codec.Packet) (int64, string, error) {
		playerID, err := infollow.Decode(packet)
		return playerID, "", err
	}, handler.HandleFollow, log)
	registerTarget(registry, inprivate.Header, PrivateName, func(packet codec.Packet) (int64, string, error) {
		payload, err := inprivate.Decode(packet)
		return payload.PlayerID, payload.Message, err
	}, handler.HandlePrivate, log)
	privacyDispatcher, _ := command.NewDispatcher(privacyHandler{handler: handler})
	privacyDispatcher.WithLogger(log)
	_ = registry.Register(inprivacy.Header, func(connection netconn.Context, packet codec.Packet) error {
		blocked, err := inprivacy.Decode(packet)
		if err != nil {
			return err
		}
		return privacyDispatcher.Dispatch(context.Background(), command.Envelope[PrivacyCommand]{Command: PrivacyCommand{Connection: connection, RoomInvitesBlocked: blocked}, Metadata: metadata(connection)})
	})
}

// inviteHandler adapts invite commands to domain behavior.
type inviteHandler struct {
	// handler stores social command behavior.
	handler Handler
}

// Handle executes one invite command.
func (handler inviteHandler) Handle(ctx context.Context, envelope command.Envelope[InviteCommand]) error {
	return handler.handler.HandleInvite(ctx, envelope.Command)
}

// targetHandler adapts target commands to domain behavior.
type targetHandler struct {
	// execute performs the selected target action.
	execute func(context.Context, TargetCommand) error
}

// privacyHandler adapts privacy commands to domain behavior.
type privacyHandler struct {
	// handler stores social command behavior.
	handler Handler
}

// Handle executes one privacy command.
func (handler privacyHandler) Handle(ctx context.Context, envelope command.Envelope[PrivacyCommand]) error {
	return handler.handler.HandlePrivacy(ctx, envelope.Command)
}

// Handle executes one target command.
func (handler targetHandler) Handle(ctx context.Context, envelope command.Envelope[TargetCommand]) error {
	return handler.execute(ctx, envelope.Command)
}

// targetDecoder decodes a target id and optional message.
type targetDecoder func(codec.Packet) (int64, string, error)

// registerTarget registers one target-based social command.
func registerTarget(registry *netconn.HandlerRegistry, header uint16, name command.Name, decode targetDecoder, execute func(context.Context, TargetCommand) error, log *zap.Logger) {
	dispatcher, _ := command.NewDispatcher(targetHandler{execute: execute})
	dispatcher.WithLogger(log)
	_ = registry.Register(header, func(connection netconn.Context, packet codec.Packet) error {
		playerID, message, err := decode(packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[TargetCommand]{Command: TargetCommand{Connection: connection, PlayerID: playerID, Message: message, Name: name}, Metadata: metadata(connection)})
	})
}

// metadata creates command tracing metadata for one connection.
func metadata(connection netconn.Context) command.Metadata {
	return command.Metadata{ConnectionID: string(connection.ConnectionID)}
}
