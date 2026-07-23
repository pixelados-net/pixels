package game

import (
	"context"
	"errors"
	"time"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

// Coordinator owns WIRED game lifecycle, blobs, boards, and trigger projection.
type Coordinator struct {
	// games stores ephemeral team and score state.
	games *Service
	// highscores persists normalized board rows.
	highscores record.HighscoreStore
	// rooms resolves active lifecycle owners.
	rooms *roomlive.Registry
	// furniture persists blob and board object data.
	furniture furnitureservice.StateUpdater
	// connections broadcasts specialized item updates.
	connections *netconn.Registry
	// engine executes lifecycle and score triggers.
	engine *wiredruntime.Engine
	// now supplies UTC period boundaries.
	now func() time.Time
	// highscoreTop bounds protocol and persistence ranking work.
	highscoreTop int
}

// NewCoordinator creates the authoritative WIRED game coordinator.
func NewCoordinator(config roomwired.Config, games *Service, highscores record.HighscoreStore, rooms *roomlive.Registry, furniture furnitureservice.StateUpdater, connections *netconn.Registry, engine *wiredruntime.Engine) *Coordinator {
	return &Coordinator{games: games, highscores: highscores, rooms: rooms, furniture: furniture, connections: connections, engine: engine, now: time.Now, highscoreTop: config.Normalize().HighscoreTop}
}

// Start initializes one active room game and resets configured blobs.
func (coordinator *Coordinator) Start(ctx context.Context, roomID int64) error {
	active, found := coordinator.rooms.Find(roomID)
	if !found || !coordinator.games.Start(roomID) {
		return ErrGameUnavailable
	}
	if err := coordinator.updateBlobs(ctx, active, true, "0"); err != nil {
		coordinator.games.Reset(roomID)
		return err
	}
	coordinator.schedule(active, trigger.Event{Kind: trigger.GameStarted, RoomID: roomID})
	return nil
}

// End persists boards, emits team outcomes, and closes one active game.
func (coordinator *Coordinator) End(ctx context.Context, roomID int64) error {
	active, found := coordinator.rooms.Find(roomID)
	state, running := coordinator.games.Snapshot(roomID)
	if !found || !running || !state.Running {
		return ErrGameUnavailable
	}
	winners := winningTeams(state)
	if err := coordinator.projectBoards(ctx, active, state, winners); err != nil {
		return err
	}
	if !coordinator.games.End(roomID) {
		return ErrGameUnavailable
	}
	if err := coordinator.updateBlobs(ctx, active, false, "1"); err != nil {
		return err
	}
	coordinator.scheduleOutcomes(active, state, winners)
	return nil
}

// Reset clears one room game and restores all blobs to their used state.
func (coordinator *Coordinator) Reset(ctx context.Context, roomID int64) error {
	active, found := coordinator.rooms.Find(roomID)
	if !found {
		return ErrGameUnavailable
	}
	coordinator.games.Reset(roomID)
	return coordinator.updateBlobs(ctx, active, false, "1")
}

// AddScore applies authoritative game score and schedules threshold triggers.
func (coordinator *Coordinator) AddScore(_ context.Context, roomID int64, playerID int64, amount int64) bool {
	previous, current, changed := coordinator.games.AddScore(roomID, playerID, amount)
	if !changed {
		return false
	}
	active, found := coordinator.rooms.Find(roomID)
	if found {
		coordinator.schedule(active, trigger.Event{Kind: trigger.ScoreAchieved, RoomID: roomID, ActorKind: trigger.ActorPlayer, ActorID: playerID, PlayerID: playerID, PreviousScore: previous, Score: current})
	}
	return true
}

// NotifyScore schedules a threshold crossing from a game-owned counter.
func (coordinator *Coordinator) NotifyScore(roomID int64, playerID int64, previous int64, current int64) {
	active, found := coordinator.rooms.Find(roomID)
	if found {
		coordinator.schedule(active, trigger.Event{Kind: trigger.ScoreAchieved, RoomID: roomID, ActorKind: trigger.ActorPlayer, ActorID: playerID, PlayerID: playerID, PreviousScore: previous, Score: current})
	}
}

// updateItem persists, swaps, and broadcasts one specialized furniture state.
func (coordinator *Coordinator) updateItem(ctx context.Context, active *roomlive.Room, item worldfurniture.Item, next string) error {
	if item.ExtraData == next {
		return nil
	}
	if _, err := coordinator.furniture.UpdateState(ctx, furnitureservice.StateParams{ItemID: item.ID, RoomID: active.ID(), Expected: item.ExtraData, Next: next}); err != nil {
		return err
	}
	updated, err := active.UpdateFurnitureState(item.ID, next, false)
	if err != nil {
		return err
	}
	packet, err := outupdate.Encode(outupdate.FloorItem{ID: updated.ID, SpriteID: updated.Definition.SpriteID, X: int(updated.Point.X), Y: int(updated.Point.Y), Rotation: int(updated.Rotation), Z: updated.Z.String(), ExtraHeight: updated.Top().String(), ExtraData: updated.ExtraData, Data: projection.SpecializedObjectData(updated.Definition.InteractionType, updated.ExtraData), OwnerID: updated.OwnerPlayerID})
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, coordinator.connections, active, packet, 0)
}

// ErrGameUnavailable reports an absent or invalid lifecycle transition.
var ErrGameUnavailable = errors.New("WIRED game unavailable")
