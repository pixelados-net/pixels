package core

import (
	"context"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
)

// Ignored returns durable ignored users and refreshes the live projection.
func (service *Service) Ignored(ctx context.Context, playerID int64) ([]messengermodel.IgnoredPlayer, error) {
	items, err := service.store.ListIgnored(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if player, found := service.live.Find(playerID); found {
		ids := make([]int64, len(items))
		for index := range items {
			ids[index] = items[index].PlayerID
		}
		player.ReplaceIgnored(ids)
	}
	return items, nil
}

// IgnoreByName persists and projects one username ignore.
func (service *Service) IgnoreByName(ctx context.Context, playerID int64, username string) (messengermodel.IgnoredPlayer, bool, error) {
	record, found, err := service.players.FindByUsername(ctx, username)
	if err != nil || !found || record.Player.ID == playerID {
		return messengermodel.IgnoredPlayer{}, false, err
	}
	return service.ignore(ctx, playerID, record.Player.ID, record.Player.Username)
}

// IgnoreByID persists and projects one player-id ignore.
func (service *Service) IgnoreByID(ctx context.Context, playerID int64, targetID int64) (messengermodel.IgnoredPlayer, bool, error) {
	record, found, err := service.players.FindByID(ctx, targetID)
	if err != nil || !found || targetID == playerID {
		return messengermodel.IgnoredPlayer{}, false, err
	}
	return service.ignore(ctx, playerID, targetID, record.Player.Username)
}

// UnignoreByName removes and projects one username ignore.
func (service *Service) UnignoreByName(ctx context.Context, playerID int64, username string) (messengermodel.IgnoredPlayer, bool, error) {
	record, found, err := service.players.FindByUsername(ctx, username)
	if err != nil || !found {
		return messengermodel.IgnoredPlayer{}, false, err
	}
	removed, err := service.store.RemoveIgnored(ctx, playerID, record.Player.ID)
	if err == nil {
		if player, online := service.live.Find(playerID); online {
			player.Unignore(record.Player.ID)
		}
	}
	return messengermodel.IgnoredPlayer{PlayerID: record.Player.ID, Username: record.Player.Username}, removed, err
}

// Relationships returns public relationship summaries assigned by one player.
func (service *Service) Relationships(ctx context.Context, playerID int64) ([]messengermodel.RelationshipSummary, error) {
	return service.store.RelationshipSummaries(ctx, playerID)
}

// ViewProfile records one live player's currently opened public profile.
func (service *Service) ViewProfile(viewerID int64, playerID int64) bool {
	viewer, found := service.live.Find(viewerID)
	return found && viewer.ViewProfile(playerID)
}

// RelationshipViewers returns online players currently observing one profile.
func (service *Service) RelationshipViewers(playerID int64) []int64 {
	players := service.live.Snapshot()
	viewers := make([]int64, 0)
	for _, player := range players {
		viewedID, found := player.ViewedProfile()
		if found && viewedID == playerID {
			viewers = append(viewers, player.ID())
		}
	}
	return viewers
}

// ignore applies one durable and live ignore mutation.
func (service *Service) ignore(ctx context.Context, playerID int64, targetID int64, username string) (messengermodel.IgnoredPlayer, bool, error) {
	added, err := service.store.AddIgnored(ctx, playerID, targetID)
	if err == nil {
		if player, online := service.live.Find(playerID); online {
			player.Ignore(targetID)
		}
	}
	return messengermodel.IgnoredPlayer{PlayerID: targetID, Username: username}, added, err
}
