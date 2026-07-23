// Package profile refreshes public messenger profiles observed by live players.
package profile

import (
	"context"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	"github.com/niflaot/pixels/networking/codec"
	"go.uber.org/zap"
)

// Reader supplies relationship summaries and active profile observers.
type Reader interface {
	// Relationships returns public relationship summaries assigned by one player.
	Relationships(context.Context, int64) ([]messengermodel.RelationshipSummary, error)
	// RelationshipViewers returns online players observing one public profile.
	RelationshipViewers(int64) []int64
}

// Sender delivers one packet to an online player.
type Sender interface {
	// Send sends one packet and reports whether delivery occurred.
	Send(context.Context, int64, codec.Packet) (bool, error)
}

// RelationshipBroadcaster refreshes relationship summaries for active profile observers.
type RelationshipBroadcaster struct {
	// reader supplies current relationship and viewer state.
	reader Reader
	// sender delivers refresh packets.
	sender Sender
	// log records observer refresh failures.
	log *zap.Logger
}

// New creates a profile relationship broadcaster.
func NewRelationships(reader Reader, sender Sender, log *zap.Logger) *RelationshipBroadcaster {
	if log == nil {
		log = zap.NewNop()
	}
	return &RelationshipBroadcaster{reader: reader, sender: sender, log: log}
}

// Refresh sends one relationship snapshot to every current profile observer.
func (broadcaster *RelationshipBroadcaster) Refresh(ctx context.Context, playerID int64) error {
	items, err := broadcaster.reader.Relationships(ctx, playerID)
	if err != nil {
		return err
	}
	packet, err := delivery.RelationshipPacket(playerID, items)
	if err != nil {
		return err
	}
	for _, viewerID := range broadcaster.reader.RelationshipViewers(playerID) {
		if _, sendErr := broadcaster.sender.Send(ctx, viewerID, packet); sendErr != nil {
			broadcaster.log.Warn("messenger profile relationship refresh failed", zap.Int64("player_id", playerID), zap.Int64("viewer_id", viewerID), zap.Error(sendErr))
		}
	}
	return nil
}
