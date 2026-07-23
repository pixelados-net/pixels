package session

import (
	"context"

	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// PlayerFinder resolves exact active usernames.
type PlayerFinder interface {
	// FindByUsername returns one exact case-insensitive player record.
	FindByUsername(context.Context, string) (playerservice.Record, bool, error)
}

// HandleProfileByName resolves an exact active username through the shared profile workflow.
func (handler Handler) HandleProfileByName(ctx context.Context, connection netconn.Context, username string) error {
	if handler.Players == nil {
		return nil
	}
	record, found, err := handler.Players.FindByUsername(ctx, username)
	if err != nil || !found {
		return err
	}
	return handler.HandleProfile(ctx, ProfileCommand{Connection: connection, PlayerID: record.Player.ID, OpenWindow: true})
}
