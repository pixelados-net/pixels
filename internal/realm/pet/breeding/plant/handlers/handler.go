// Package handlers adapts monsterplant packets to commands.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incompost "github.com/niflaot/pixels/networking/inbound/room/pet/plant/compost"
	inharvest "github.com/niflaot/pixels/networking/inbound/room/pet/plant/harvest"
	insupplement "github.com/niflaot/pixels/networking/inbound/room/pet/supplement"
	"go.uber.org/zap"
)

// NewSupplement creates the plant supplement adapter.
func NewSupplement(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.SupplementName, func(packet codec.Packet) (petcommands.Command, error) {
		payload, err := insupplement.Decode(packet)
		return petcommands.Command{PetID: payload.PetID, SupplementType: payload.Type}, err
	})
}

// NewHarvest creates the plant harvest adapter.
func NewHarvest(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return idAdapter(handler, log, petcommands.HarvestName, inharvest.Decode)
}

// NewCompost creates the plant compost adapter.
func NewCompost(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return idAdapter(handler, log, petcommands.CompostName, incompost.Decode)
}

// decoder decodes one monsterplant request.
type decoder func(codec.Packet) (petcommands.Command, error)

// idDecoder decodes one monsterplant identifier.
type idDecoder func(codec.Packet) (int64, error)

// idAdapter creates one identifier-only lifecycle adapter.
func idAdapter(handler petcommands.Handler, log *zap.Logger, action command.Name, decode idDecoder) netconn.Handler {
	return adapter(handler, log, action, func(packet codec.Packet) (petcommands.Command, error) {
		petID, err := decode(packet)
		return petcommands.Command{PetID: petID}, err
	})
}

// adapter creates one lifecycle command dispatcher.
func adapter(handler petcommands.Handler, log *zap.Logger, action command.Name, decode decoder) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		value, err := decode(packet)
		if err != nil {
			return err
		}
		value.Handler, value.Action = connection, action
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.Command]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register installs every monsterplant adapter.
func Register(registry *netconn.HandlerRegistry, supplement netconn.Handler, harvest netconn.Handler, compost netconn.Handler) {
	_ = registry.Register(insupplement.Header, supplement)
	_ = registry.Register(inharvest.Header, harvest)
	_ = registry.Register(incompost.Header, compost)
}
