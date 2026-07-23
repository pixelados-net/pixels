// Package game adapts room games to the furniture interaction boundary.
package game

import (
	"context"

	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	roomgames "github.com/niflaot/pixels/internal/realm/room/world/games"
)

// Service adapts specialized furniture requests to room games.
type Service struct {
	// games owns authoritative mechanics.
	games *roomgames.Service
}

// New creates a furniture game adapter.
func New(games *roomgames.Service) *Service { return &Service{games: games} }

// UseFurniture forwards one matching furniture interaction.
func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error) {
	if service == nil || service.games == nil {
		return false, nil
	}
	return service.games.UseFurniture(ctx, roomgames.UseRequest{PlayerID: request.PlayerID, Room: request.Room, Item: request.Item, State: request.State})
}

// Register attaches the game adapter to specialized furniture dispatch.
func Register(essentials *essential.Service, service *Service) {
	if essentials != nil {
		essentials.AddExternal(service)
	}
}

var externalAssertion essential.External = (*Service)(nil)
