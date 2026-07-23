// Package commands owns pet catalog protocol workflows.
package commands

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcatalog "github.com/niflaot/pixels/internal/realm/pet/catalog"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// BreedsName identifies pet palette requests.
	BreedsName command.Name = "pet.catalog.breeds"
	// NameApprovalName identifies pet name validation.
	NameApprovalName command.Name = "pet.catalog.name.approve"
	// PackageName identifies package opening.
	PackageName command.Name = "pet.package.open"
)

// Command stores one pet catalog request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Action identifies the catalog workflow.
	Action command.Name
	// ProductCode stores the requested species product code.
	ProductCode string
	// Name stores a proposed pet name.
	Name string
	// ObjectID identifies a furniture package.
	ObjectID int64
}

// CommandName returns the selected stable command name.
func (value Command) CommandName() command.Name { return value.Action }

// Handler executes pet catalog requests.
type Handler struct {
	// Service owns pet catalog behavior.
	Service *petcatalog.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connection bindings.
	Bindings *binding.Registry
}

// Handle executes one authenticated catalog workflow.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := petsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	switch envelope.Command.Action {
	case BreedsName:
		return handler.Service.SendBreeds(ctx, envelope.Command.Handler, envelope.Command.ProductCode)
	case NameApprovalName:
		return handler.Service.SendNameApproval(ctx, envelope.Command.Handler, envelope.Command.Name)
	case PackageName:
		roomID, found := player.CurrentRoom()
		if !found {
			return nil
		}
		return handler.Service.OpenPackage(ctx, envelope.Command.Handler, player.ID(), roomID, envelope.Command.ObjectID, envelope.Command.Name)
	default:
		return nil
	}
}
