package core

import (
	"context"
	"testing"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
)

// cardsStoreFixture records optimized durable-card projection calls.
type cardsStoreFixture struct {
	*fakeStore
	// cards stores the projected result.
	cards []messengermodel.Card
	// calls counts optimized projection reads.
	calls int
}

// ListCards returns the configured optimized projection.
func (store *cardsStoreFixture) ListCards(context.Context, int64) ([]messengermodel.Card, error) {
	store.calls++
	return store.cards, nil
}

// TestLoadDurableCardsUsesOptimizedProjection verifies production repositories avoid per-friend reads.
func TestLoadDurableCardsUsesOptimizedProjection(t *testing.T) {
	store := &cardsStoreFixture{
		fakeStore: newFakeStore(),
		cards:     []messengermodel.Card{{ID: 2, Username: "alice"}},
	}
	service := &Service{store: store}

	cards, err := service.loadDurableCards(context.Background(), 1)
	if err != nil || store.calls != 1 || len(cards) != 1 || cards[0].Username != "alice" {
		t.Fatalf("unexpected cards=%#v calls=%d err=%v", cards, store.calls, err)
	}
}
