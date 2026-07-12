package session

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inignoreid "github.com/niflaot/pixels/networking/inbound/user/ignore/id"
	inignorelist "github.com/niflaot/pixels/networking/inbound/user/ignore/list"
	inignorename "github.com/niflaot/pixels/networking/inbound/user/ignore/name"
	inignoreremove "github.com/niflaot/pixels/networking/inbound/user/ignore/remove"
	inrelationships "github.com/niflaot/pixels/networking/inbound/user/relationships"
	"go.uber.org/zap"
)

// RegisterHandlers registers ignored-user and relationship packet adapters.
func RegisterPrivacyHandlers(registry *netconn.HandlerRegistry, handler PrivacyHandler, log *zap.Logger) {
	registerPrivacy(registry, inignorelist.Header, func(packet codec.Packet) (PrivacyCommand, error) {
		_, err := inignorelist.Decode(packet)
		return PrivacyCommand{Name: ListName}, err
	}, handler, log)
	registerPrivacy(registry, inignorename.Header, func(packet codec.Packet) (PrivacyCommand, error) {
		username, err := inignorename.Decode(packet)
		return PrivacyCommand{Name: IgnoreName, Username: username}, err
	}, handler, log)
	registerPrivacy(registry, inignoreid.Header, func(packet codec.Packet) (PrivacyCommand, error) {
		playerID, err := inignoreid.Decode(packet)
		return PrivacyCommand{Name: IgnoreName, PlayerID: playerID}, err
	}, handler, log)
	registerPrivacy(registry, inignoreremove.Header, func(packet codec.Packet) (PrivacyCommand, error) {
		username, err := inignoreremove.Decode(packet)
		return PrivacyCommand{Name: UnignoreName, Username: username}, err
	}, handler, log)
	registerPrivacy(registry, inrelationships.Header, func(packet codec.Packet) (PrivacyCommand, error) {
		playerID, err := inrelationships.Decode(packet)
		return PrivacyCommand{Name: RelationshipsName, PlayerID: playerID}, err
	}, handler, log)
}

// commandHandler adapts privacy commands to domain behavior.
type commandHandler struct {
	// handler stores privacy behavior.
	handler PrivacyHandler
}

// Handle executes one privacy command envelope.
func (handler commandHandler) Handle(ctx context.Context, envelope command.Envelope[PrivacyCommand]) error {
	return handler.handler.Handle(ctx, envelope.Command)
}

// decoder decodes one privacy packet.
type privacyDecoder func(codec.Packet) (PrivacyCommand, error)

// register installs one privacy command adapter.
func registerPrivacy(registry *netconn.HandlerRegistry, header uint16, decode privacyDecoder, handler PrivacyHandler, log *zap.Logger) {
	dispatcher, _ := command.NewDispatcher(commandHandler{handler: handler})
	dispatcher.WithLogger(log)
	_ = registry.Register(header, func(connection netconn.Context, packet codec.Packet) error {
		input, err := decode(packet)
		if err != nil {
			return err
		}
		input.Connection = connection
		return dispatcher.Dispatch(context.Background(), command.Envelope[PrivacyCommand]{Command: input, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	})
}
