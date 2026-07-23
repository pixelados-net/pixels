// Package handlers adapts pet catalog packets to commands.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/catalog/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbreeds "github.com/niflaot/pixels/networking/inbound/catalog/pet/breeds"
	inpackage "github.com/niflaot/pixels/networking/inbound/room/pet/package/open"
	inapprove "github.com/niflaot/pixels/networking/inbound/user/name/approve"
	"go.uber.org/zap"
)

// NewBreeds creates the pet palette adapter.
func NewBreeds(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, func(packet codec.Packet) (petcommands.Command, error) {
		value, err := inbreeds.Decode(packet)
		return petcommands.Command{Action: petcommands.BreedsName, ProductCode: value}, err
	})
}

// NewNameApproval creates the pet name adapter.
func NewNameApproval(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, func(packet codec.Packet) (petcommands.Command, error) {
		value, err := inapprove.Decode(packet)
		if err == nil && value.Type != 1 {
			return petcommands.Command{}, nil
		}
		return petcommands.Command{Action: petcommands.NameApprovalName, Name: value.Name}, err
	})
}

// NewPackage creates the pet package adapter.
func NewPackage(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, func(packet codec.Packet) (petcommands.Command, error) {
		value, err := inpackage.Decode(packet)
		return petcommands.Command{Action: petcommands.PackageName, ObjectID: value.ObjectID, Name: value.Name}, err
	})
}

// decoder decodes one pet catalog request.
type decoder func(codec.Packet) (petcommands.Command, error)

// adapter creates one pet catalog dispatcher.
func adapter(handler petcommands.Handler, log *zap.Logger, decode decoder) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		value, err := decode(packet)
		if err != nil {
			return err
		}
		value.Handler = connection
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.Command]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register installs every pet catalog adapter.
func Register(registry *netconn.HandlerRegistry, breeds netconn.Handler, approve netconn.Handler, open netconn.Handler) {
	_ = registry.Register(inbreeds.Header, breeds)
	_ = registry.Register(inapprove.Header, approve)
	_ = registry.Register(inpackage.Header, open)
}
