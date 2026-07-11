package pair

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
)

const (
	// CommandName identifies the teleport pair command.
	CommandName command.Name = "furniture.teleport.pair"
)

// Command requests one durable teleport relationship.
type Command struct {
	// ActorPlayerID identifies the owner creating the relationship.
	ActorPlayerID int64
	// FirstItemID identifies one teleport item.
	FirstItemID int64
	// SecondItemID identifies the other teleport item.
	SecondItemID int64
}

// Handler executes teleport pair commands.
type Handler struct {
	// Service manages durable relationships.
	Service *Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return CommandName
}

// Handle validates and stores one teleport pair.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	_, err := handler.Service.Pair(ctx, envelope.Command.ActorPlayerID, envelope.Command.FirstItemID, envelope.Command.SecondItemID)

	return err
}
