package core

import (
	"context"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
)

// durableCardsStore loads complete durable friend projections in one query.
type durableCardsStore interface {
	// ListCards returns one player's durable friend-card projections.
	ListCards(context.Context, int64) ([]messengermodel.Card, error)
}

// loadDurableCards uses the optimized repository projection with a compatibility fallback.
func (service *Service) loadDurableCards(ctx context.Context, playerID int64) ([]messengermodel.Card, error) {
	if store, supported := service.store.(durableCardsStore); supported {
		return store.ListCards(ctx, playerID)
	}
	friendships, err := service.store.ListFriends(ctx, playerID)
	if err != nil {
		return nil, err
	}
	cards := make([]messengermodel.Card, 0, len(friendships))
	for _, friendship := range friendships {
		record, found, findErr := service.players.FindByID(ctx, friendship.FriendPlayerID)
		if findErr != nil {
			return nil, findErr
		}
		if found {
			cards = append(cards, durableCardFromRecord(record, friendship.Relation, friendship.CategoryID))
		}
	}
	return cards, nil
}
