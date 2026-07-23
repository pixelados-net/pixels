// Package handlers adapts pet equipment packets to commands.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/equipment/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbreed "github.com/niflaot/pixels/networking/inbound/room/pet/breeding/toggle"
	inhanditem "github.com/niflaot/pixels/networking/inbound/room/pet/handitem"
	inmount "github.com/niflaot/pixels/networking/inbound/room/pet/mount"
	inproduct "github.com/niflaot/pixels/networking/inbound/room/pet/product/use"
	inride "github.com/niflaot/pixels/networking/inbound/room/pet/riding/toggle"
	insaddle "github.com/niflaot/pixels/networking/inbound/room/pet/saddle/remove"
	"go.uber.org/zap"
)

// NewProduct creates typed product use adapter.
func NewProduct(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.ProductName, func(packet codec.Packet) (petcommands.Command, error) {
		payload, err := inproduct.Decode(packet)
		return petcommands.Command{PetID: payload.PetID, ItemID: payload.ItemID}, err
	})
}

// NewHandItem creates virtual hand-item feeding adapter.
func NewHandItem(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.HandItemName, func(packet codec.Packet) (petcommands.Command, error) {
		unitID, err := inhanditem.Decode(packet)
		return petcommands.Command{PetID: unitID}, err
	})
}

// NewMount creates mount and dismount adapter.
func NewMount(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.MountName, func(packet codec.Packet) (petcommands.Command, error) {
		payload, err := inmount.Decode(packet)
		return petcommands.Command{PetID: payload.PetID, Flag: payload.Mount}, err
	})
}

// NewSaddle creates saddle removal adapter.
func NewSaddle(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return idAdapter(handler, log, petcommands.SaddleName, insaddle.Decode)
}

// NewRiding creates public-riding toggle adapter.
func NewRiding(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return idAdapter(handler, log, petcommands.RidingName, inride.Decode)
}

// NewBreeding creates public-breeding toggle adapter.
func NewBreeding(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return idAdapter(handler, log, petcommands.BreedingName, inbreed.Decode)
}

// decoder decodes one equipment request.
type decoder func(codec.Packet) (petcommands.Command, error)

// idDecoder decodes one pet identifier.
type idDecoder func(codec.Packet) (int64, error)

// idAdapter creates one identifier-only adapter.
func idAdapter(handler petcommands.Handler, log *zap.Logger, action command.Name, decode idDecoder) netconn.Handler {
	return adapter(handler, log, action, func(packet codec.Packet) (petcommands.Command, error) {
		petID, err := decode(packet)
		return petcommands.Command{PetID: petID}, err
	})
}

// adapter creates one equipment command dispatcher.
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

// Register installs every equipment packet adapter.
func Register(registry *netconn.HandlerRegistry, product netconn.Handler, handItem netconn.Handler, mount netconn.Handler, saddle netconn.Handler, riding netconn.Handler, breeding netconn.Handler) {
	_ = registry.Register(inproduct.Header, product)
	_ = registry.Register(inhanditem.Header, handItem)
	_ = registry.Register(inmount.Header, mount)
	_ = registry.Register(insaddle.Header, saddle)
	_ = registry.Register(inride.Header, riding)
	_ = registry.Register(inbreed.Header, breeding)
}
