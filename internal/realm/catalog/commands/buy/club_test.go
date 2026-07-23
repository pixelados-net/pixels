package buy

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outstatus "github.com/niflaot/pixels/networking/outbound/subscription/status"
)

// clubPurchase records one catalog-routed subscription purchase.
type clubPurchase struct {
	// playerID stores the purchasing player.
	playerID int64
	// offerID stores the selected subscription offer.
	offerID int64
	// amount stores the selected quantity.
	amount int32
}

// PurchaseClub records one subscription purchase.
func (purchase *clubPurchase) PurchaseClub(_ context.Context, playerID int64, offerID int64, amount int32) (ClubPurchase, error) {
	purchase.playerID, purchase.offerID, purchase.amount = playerID, offerID, amount
	return ClubPurchase{ExpiresAt: time.Now().Add(31 * 24 * time.Hour), VIP: true}, nil
}

// TestHandleRoutesClubLayoutToSubscription verifies Nitro's normal catalog purchase path.
func TestHandleRoutesClubLayoutToSubscription(t *testing.T) {
	handler, connection, sent, manager := buyFixture(t)
	club := &clubPurchase{}
	manager.page = catalogmodel.Page{Layout: ClubLayout}
	handler.Club = club
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Connection: connection, PageID: 101, OfferID: 4, Amount: 2,
	}})
	if err != nil || manager.purchases != 0 || club.playerID != 7 || club.offerID != 4 ||
		club.amount != 2 || len(*sent) != 2 || (*sent)[0].Header != outok.Header || (*sent)[1].Header != outstatus.Header {
		t.Fatalf("club=%#v packets=%#v purchases=%d error=%v", club, *sent, manager.purchases, err)
	}
}
