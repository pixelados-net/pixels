// Package handlers adapts pet care packets to commands.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/care/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrespect "github.com/niflaot/pixels/networking/inbound/room/pet/respect"
	intraining "github.com/niflaot/pixels/networking/inbound/room/pet/training"
	"go.uber.org/zap"
)

// NewRespect creates the pet respect packet adapter.
func NewRespect(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.RespectName, inrespect.Decode)
}

// NewTraining creates the pet training packet adapter.
func NewTraining(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.TrainingName, intraining.Decode)
}

// decoder decodes one care request pet identifier.
type decoder func(codec.Packet) (int64, error)

// adapter creates one care command dispatcher.
func adapter(handler petcommands.Handler, log *zap.Logger, action command.Name, decode decoder) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		petID, err := decode(packet)
		if err != nil {
			return err
		}
		value := petcommands.Command{Handler: connection, PetID: petID, Action: action}
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.Command]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register installs pet care packet handlers.
func Register(registry *netconn.HandlerRegistry, respect netconn.Handler, training netconn.Handler) {
	_ = registry.Register(inrespect.Header, respect)
	_ = registry.Register(intraining.Header, training)
}
