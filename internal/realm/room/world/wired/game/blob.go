package game

import (
	"context"
	"strconv"
	"strings"

	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Blob consumes one active WIRED game pickup after a completed walk.
func (coordinator *Coordinator) Blob(ctx context.Context, payload furniturewalkedon.Payload) error {
	active, found := coordinator.rooms.Find(payload.RoomID)
	if !found {
		return nil
	}
	item, found := active.FurnitureItem(payload.ItemID)
	if !found || item.Definition.InteractionType != "wf_blob" || item.ExtraData != "0" {
		return nil
	}
	points, _ := blobParameters(item)
	previous, current, scored := coordinator.games.AddScore(payload.RoomID, payload.PlayerID, points)
	if !scored {
		return nil
	}
	if err := coordinator.updateItem(ctx, active, item, "1"); err != nil {
		return err
	}
	coordinator.schedule(active, trigger.Event{Kind: trigger.ScoreAchieved, RoomID: payload.RoomID, ActorKind: trigger.ActorPlayer, ActorID: payload.PlayerID, PlayerID: payload.PlayerID, SourceItem: payload.ItemID, SourceSprite: int32(item.Definition.SpriteID), PreviousScore: previous, Score: current})
	return nil
}

// updateBlobs applies game lifecycle state to eligible pickups.
func (coordinator *Coordinator) updateBlobs(ctx context.Context, active *roomlive.Room, resetOnly bool, next string) error {
	for _, item := range active.FurnitureByInteraction("wf_blob") {
		_, resets := blobParameters(item)
		if resetOnly && !resets {
			continue
		}
		if err := coordinator.updateItem(ctx, active, item, next); err != nil {
			return err
		}
	}
	return nil
}

// blobParameters parses audited points and reset behavior with safe defaults.
func blobParameters(item worldfurniture.Item) (int64, bool) {
	points, resets := int64(1), true
	parts := strings.Split(item.Definition.CustomParams, ",")
	if len(parts) > 0 {
		if value, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32); err == nil && value > 0 {
			points = value
		}
	}
	if len(parts) > 1 {
		if value, err := strconv.ParseBool(strings.TrimSpace(parts[1])); err == nil {
			resets = value
		}
	}
	return points, resets
}
