package service

import (
	"context"
	"testing"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// giftPlayerFinder resolves one configured recipient fixture.
type giftPlayerFinder struct {
	// record stores the configured recipient.
	record playerservice.Record
}

// FindByID supplies unused player lookup behavior.
func (finder giftPlayerFinder) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// FindByUsername resolves the configured gift recipient.
func (finder giftPlayerFinder) FindByUsername(_ context.Context, username string) (playerservice.Record, bool, error) {
	return finder.record, username == finder.record.Player.Username, nil
}

// TestPurchaseGiftChargesBuyerAndWrapsRecipientItem verifies durable gift ownership and metadata.
func TestPurchaseGiftChargesBuyerAndWrapsRecipientItem(t *testing.T) {
	item := itemForTest()
	item.Giftable = true
	fixture := newServiceFixture(t, item)
	receiver := playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 12}}, Username: "alice"}
	fixture.service.WithPlayers(giftPlayerFinder{record: playerservice.Record{Player: receiver}})

	result, err := fixture.service.PurchaseGift(context.Background(), GiftPurchaseParams{
		BuyerID: 7, ReceiverName: " alice ", CatalogItemID: item.ID, SpriteID: 3372, BoxID: 8, RibbonID: 10,
		Message: "Enjoy", ShowMyFace: true,
	})
	if err != nil || result.RecipientPlayerID != 12 || len(result.GrantedItems) != 1 {
		t.Fatalf("unexpected gift result %#v error %v", result, err)
	}
	if len(fixture.currency.calls) != 1 || fixture.currency.calls[0].PlayerID != 7 {
		t.Fatalf("unexpected buyer charge %#v", fixture.currency.calls)
	}
	if len(fixture.furniture.giftCalls) != 1 {
		t.Fatalf("unexpected gift grants %#v", fixture.furniture.giftCalls)
	}
	grant := fixture.furniture.giftCalls[0]
	if grant.OwnerPlayerID != 12 || grant.SpriteID != 3372 || grant.BoxID != 8 || grant.RibbonID != 10 || grant.SenderPlayerID == nil || *grant.SenderPlayerID != 7 {
		t.Fatalf("unexpected gift metadata %#v", grant)
	}
}
