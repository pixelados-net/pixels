package profile

import (
	"context"
	"testing"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/networking/codec"
	outrelationships "github.com/niflaot/pixels/networking/outbound/user/relationships"
)

// readerForTest supplies one relationship summary and two observers.
type readerForTest struct{}

// Relationships returns one relationship summary.
func (readerForTest) Relationships(context.Context, int64) ([]messengermodel.RelationshipSummary, error) {
	return []messengermodel.RelationshipSummary{{Relation: messengermodel.RelationHeart, Count: 1}}, nil
}

// RelationshipViewers returns two profile observers.
func (readerForTest) RelationshipViewers(int64) []int64 { return []int64{2, 3} }

// senderForTest captures profile refresh recipients.
type senderForTest struct {
	// playerIDs stores delivered observer ids.
	playerIDs []int64
	// packets stores delivered relationship packets.
	packets []codec.Packet
}

// Send captures one profile refresh.
func (sender *senderForTest) Send(_ context.Context, playerID int64, packet codec.Packet) (bool, error) {
	sender.playerIDs = append(sender.playerIDs, playerID)
	sender.packets = append(sender.packets, packet)
	return true, nil
}

// TestRefreshSendsCurrentSummaryOnlyToObservers verifies targeted live projection.
func TestRefreshSendsCurrentSummaryOnlyToObservers(t *testing.T) {
	sender := &senderForTest{}
	broadcaster := NewRelationships(readerForTest{}, sender, nil)
	if err := broadcaster.Refresh(context.Background(), 1); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if len(sender.playerIDs) != 2 || sender.playerIDs[0] != 2 || sender.playerIDs[1] != 3 {
		t.Fatalf("unexpected viewers=%v", sender.playerIDs)
	}
	if len(sender.packets) != 2 || sender.packets[0].Header != outrelationships.Header {
		t.Fatalf("unexpected packets=%v", sender.packets)
	}
}
