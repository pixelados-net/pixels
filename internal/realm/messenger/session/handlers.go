package session

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininit "github.com/niflaot/pixels/networking/inbound/messenger/session/init"
	inrefresh "github.com/niflaot/pixels/networking/inbound/messenger/session/refresh"
	inrequests "github.com/niflaot/pixels/networking/inbound/messenger/session/requests"
	insearch "github.com/niflaot/pixels/networking/inbound/messenger/session/search"
	inupdates "github.com/niflaot/pixels/networking/inbound/messenger/session/updates"
	infind "github.com/niflaot/pixels/networking/inbound/messenger/social/findroom"
	inprofile "github.com/niflaot/pixels/networking/inbound/user/profile"
	inprofilebyname "github.com/niflaot/pixels/networking/inbound/user/profile/byname"
	"go.uber.org/zap"
)

// RegisterHandlers registers messenger session packet adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler, log *zap.Logger) {
	registerConnection(registry, ininit.Header, ininit.Decode, handler, InitName, handler.HandleInit, log)
	registerConnection(registry, inrefresh.Header, inrefresh.Decode, handler, RefreshName, handler.HandleRefresh, log)
	registerConnection(registry, inupdates.Header, inupdates.Decode, handler, RefreshName, handler.HandleRefresh, log)
	registerConnection(registry, inrequests.Header, inrequests.Decode, handler, RequestsName, handler.HandleRequests, log)
	registerConnection(registry, infind.Header, infind.Decode, handler, FindRoomName, handler.HandleFindRoom, log)
	searchDispatcher, _ := command.NewDispatcher(searchHandler{handler: handler})
	searchDispatcher.WithLogger(log)
	_ = registry.Register(insearch.Header, func(connection netconn.Context, packet codec.Packet) error {
		term, err := insearch.Decode(packet)
		if err != nil {
			return err
		}
		return searchDispatcher.Dispatch(context.Background(), command.Envelope[SearchCommand]{
			Command:  SearchCommand{Connection: connection, Term: term},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	})
	profileDispatcher, _ := command.NewDispatcher(profileHandler{handler: handler})
	profileDispatcher.WithLogger(log)
	_ = registry.Register(inprofile.Header, func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inprofile.Decode(packet)
		if err != nil {
			return err
		}
		return profileDispatcher.Dispatch(context.Background(), command.Envelope[ProfileCommand]{Command: ProfileCommand{Connection: connection, PlayerID: payload.PlayerID, OpenWindow: payload.OpenWindow}, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	})
	_ = registry.Register(inprofilebyname.Header, func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inprofilebyname.Decode(packet)
		if err != nil {
			return err
		}
		return handler.HandleProfileByName(context.Background(), connection, payload.Username)
	})
}

// connectionDecoder validates one fieldless inbound packet.
type connectionDecoder func(codec.Packet) error

// connectionExecutor executes one connection command.
type connectionExecutor func(context.Context, ConnectionCommand) error

// connectionHandler adapts connection commands to the typed command bus.
type connectionHandler struct {
	// execute performs the command behavior.
	execute connectionExecutor
}

// Handle executes one connection command.
func (handler connectionHandler) Handle(ctx context.Context, envelope command.Envelope[ConnectionCommand]) error {
	return handler.execute(ctx, envelope.Command)
}

// searchHandler adapts search commands to the typed command bus.
type searchHandler struct {
	// handler stores session command behavior.
	handler Handler
}

// profileHandler adapts profile commands to the typed command bus.
type profileHandler struct {
	// handler stores session command behavior.
	handler Handler
}

// Handle executes one profile command.
func (handler profileHandler) Handle(ctx context.Context, envelope command.Envelope[ProfileCommand]) error {
	return handler.handler.HandleProfile(ctx, envelope.Command)
}

// Handle executes one search command.
func (handler searchHandler) Handle(ctx context.Context, envelope command.Envelope[SearchCommand]) error {
	return handler.handler.HandleSearch(ctx, envelope.Command)
}

// registerConnection registers one fieldless messenger command adapter.
func registerConnection(registry *netconn.HandlerRegistry, header uint16, decode connectionDecoder, handler Handler, name command.Name, execute connectionExecutor, log *zap.Logger) {
	dispatcher, _ := command.NewDispatcher(connectionHandler{execute: execute})
	dispatcher.WithLogger(log)
	_ = registry.Register(header, func(connection netconn.Context, packet codec.Packet) error {
		if err := decode(packet); err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[ConnectionCommand]{
			Command:  ConnectionCommand{Connection: connection, Name: name},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	})
}
