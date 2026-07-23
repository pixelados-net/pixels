package core

import (
	"context"
	"encoding/json"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
)

// cachedCards returns one shared durable friend-card snapshot when available.
func (service *Service) cachedCards(ctx context.Context, playerID int64) ([]messengermodel.Card, bool) {
	if service.redis == nil || service.options.FriendCacheTTL <= 0 {
		return nil, false
	}
	encoded, found, err := service.redis.Find(ctx, friendCardsCacheKey(playerID))
	if err != nil || !found {
		return nil, false
	}
	var cards []messengermodel.Card
	if json.Unmarshal(encoded, &cards) != nil {
		return nil, false
	}
	return cards, true
}

// cacheCards stores one durable friend-card snapshot without volatile presence.
func (service *Service) cacheCards(ctx context.Context, playerID int64, cards []messengermodel.Card) {
	if service.redis == nil || service.options.FriendCacheTTL <= 0 {
		return
	}
	encoded, err := json.Marshal(cards)
	if err == nil {
		_ = service.redis.Set(ctx, friendCardsCacheKey(playerID), encoded, service.options.FriendCacheTTL)
	}
}

// invalidateCards removes durable friend-card snapshots for selected players.
func (service *Service) invalidateCards(ctx context.Context, playerIDs ...int64) {
	if service.redis == nil {
		return
	}
	for _, playerID := range playerIDs {
		_ = service.redis.Delete(ctx, friendCardsCacheKey(playerID))
	}
}

// InvalidateProfile removes cached public cards after a committed profile mutation.
func (service *Service) InvalidateProfile(ctx context.Context, playerID int64) {
	service.invalidateCards(ctx, playerID)
}

// friendCardsCacheKey returns one player-scoped durable friend-card key.
func friendCardsCacheKey(playerID int64) string {
	return "messenger:friends:cards:" + integerKey(playerID)
}

// projectCardPresence overlays current process-local presence on durable cards.
func (service *Service) projectCardPresence(cards []messengermodel.Card) {
	for index := range cards {
		cards[index].Online = false
		cards[index].FollowingAllowed = false
		if player, online := service.live.Find(cards[index].ID); online {
			cards[index].Online = true
			_, inRoom := player.CurrentRoom()
			cards[index].FollowingAllowed = inRoom && !cards[index].BlockFollowing
		}
	}
}
