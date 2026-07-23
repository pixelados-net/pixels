package friend

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaccept "github.com/niflaot/pixels/networking/inbound/messenger/friend/accept"
	indecline "github.com/niflaot/pixels/networking/inbound/messenger/friend/decline"
	inrelation "github.com/niflaot/pixels/networking/inbound/messenger/friend/relation"
	inremove "github.com/niflaot/pixels/networking/inbound/messenger/friend/remove"
	inrequest "github.com/niflaot/pixels/networking/inbound/messenger/friend/request"
	"go.uber.org/zap"
)

// RegisterHandlers registers messenger friendship packet adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler, log *zap.Logger) {
	requestDispatcher, _ := command.NewDispatcher(requestHandler{handler: handler})
	requestDispatcher.WithLogger(log)
	_ = registry.Register(inrequest.Header, func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrequest.Decode(packet)
		if err != nil {
			return err
		}
		return requestDispatcher.Dispatch(context.Background(), command.Envelope[RequestCommand]{Command: RequestCommand{Connection: connection, Username: payload.Username}, Metadata: metadata(connection)})
	})
	registerBatch(registry, inaccept.Header, AcceptName, func(packet codec.Packet) ([]int64, bool, error) {
		ids, err := inaccept.Decode(packet)
		return ids, false, err
	}, handler.HandleAccept, log)
	registerBatch(registry, indecline.Header, DeclineName, func(packet codec.Packet) ([]int64, bool, error) {
		payload, err := indecline.Decode(packet)
		return payload.PlayerIDs, payload.All, err
	}, handler.HandleDecline, log)
	registerBatch(registry, inremove.Header, RemoveName, func(packet codec.Packet) ([]int64, bool, error) {
		ids, err := inremove.Decode(packet)
		return ids, false, err
	}, handler.HandleRemove, log)
	relationDispatcher, _ := command.NewDispatcher(relationHandler{handler: handler})
	relationDispatcher.WithLogger(log)
	_ = registry.Register(inrelation.Header, func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrelation.Decode(packet)
		if err != nil {
			return err
		}
		return relationDispatcher.Dispatch(context.Background(), command.Envelope[RelationCommand]{Command: RelationCommand{Connection: connection, PlayerID: payload.PlayerID, Relation: messengermodel.Relation(payload.Relation)}, Metadata: metadata(connection)})
	})
}

// requestHandler adapts request commands to domain behavior.
type requestHandler struct {
	// handler stores friendship command behavior.
	handler Handler
}

// Handle executes one friend-request command.
func (handler requestHandler) Handle(ctx context.Context, envelope command.Envelope[RequestCommand]) error {
	return handler.handler.HandleRequest(ctx, envelope.Command)
}

// batchHandler adapts batch commands to domain behavior.
type batchHandler struct {
	// execute performs the selected batch mutation.
	execute func(context.Context, BatchCommand) error
}

// Handle executes one friendship batch command.
func (handler batchHandler) Handle(ctx context.Context, envelope command.Envelope[BatchCommand]) error {
	return handler.execute(ctx, envelope.Command)
}

// relationHandler adapts relation commands to domain behavior.
type relationHandler struct {
	// handler stores friendship command behavior.
	handler Handler
}

// Handle executes one relation command.
func (handler relationHandler) Handle(ctx context.Context, envelope command.Envelope[RelationCommand]) error {
	return handler.handler.HandleRelation(ctx, envelope.Command)
}

// batchDecoder decodes one bounded id collection.
type batchDecoder func(codec.Packet) ([]int64, bool, error)

// registerBatch registers one batch mutation adapter.
func registerBatch(registry *netconn.HandlerRegistry, header uint16, name command.Name, decode batchDecoder, execute func(context.Context, BatchCommand) error, log *zap.Logger) {
	dispatcher, _ := command.NewDispatcher(batchHandler{execute: execute})
	dispatcher.WithLogger(log)
	_ = registry.Register(header, func(connection netconn.Context, packet codec.Packet) error {
		ids, all, err := decode(packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[BatchCommand]{Command: BatchCommand{Connection: connection, PlayerIDs: ids, All: all, Name: name}, Metadata: metadata(connection)})
	})
}

// metadata creates command tracing metadata for one connection.
func metadata(connection netconn.Context) command.Metadata {
	return command.Metadata{ConnectionID: string(connection.ConnectionID)}
}
