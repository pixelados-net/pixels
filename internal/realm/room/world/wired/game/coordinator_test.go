package game

import (
	"context"
	"math"
	"testing"
	"time"

	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// gameUpdater records durable furniture state transitions.
type gameUpdater struct {
	// calls stores ordered state changes.
	calls []furnitureservice.StateParams
}

// UpdateState records one state change.
func (updater *gameUpdater) UpdateState(_ context.Context, params furnitureservice.StateParams) (furnituremodel.Item, error) {
	updater.calls = append(updater.calls, params)
	return furnituremodel.Item{}, nil
}

// gameHighscores records board submissions and returns one protocol row.
type gameHighscores struct {
	// calls counts durable board updates.
	calls int
}

// RecordAndList returns a deterministic board row.
func (store *gameHighscores) RecordAndList(_ context.Context, _ int64, _ record.HighscoreMode, _ record.HighscorePeriod, _ *time.Time, results []record.HighscoreResult, _ int) ([]record.HighscoreEntry, error) {
	store.calls++
	if len(results) == 0 {
		return nil, nil
	}
	return []record.HighscoreEntry{{Score: results[0].Score, PlayerIDs: results[0].PlayerIDs, Usernames: []string{"demo"}}}, nil
}

// TestCoordinatorRunsGameBlobAndBoardLifecycle verifies real room furniture follows game state.
func TestCoordinatorRunsGameBlobAndBoardLifecycle(t *testing.T) {
	rooms, active := gameRoom(t)
	games := New()
	updater := &gameUpdater{}
	highscores := &gameHighscores{}
	coordinator := NewCoordinator(roomwired.Config{HighscoreTop: 10}, games, highscores, rooms, updater, nil, nil)
	join := &configuration.Node{Parameters: configuration.Parameters{Values: []int32{1}}}
	if _, err := games.ExecuteGame(context.Background(), effect.JoinTeam, join, trigger.Event{RoomID: active.ID(), PlayerID: 7}); err != nil {
		t.Fatal(err)
	}
	if err := coordinator.Start(context.Background(), active.ID()); err != nil {
		t.Fatal(err)
	}
	blob, _ := active.FurnitureItem(1)
	if blob.ExtraData != "0" {
		t.Fatalf("started blob state=%q", blob.ExtraData)
	}
	if err := coordinator.Blob(context.Background(), furniturewalkedon.Payload{RoomID: active.ID(), ItemID: 1, PlayerID: 7}); err != nil {
		t.Fatal(err)
	}
	state, _ := games.Snapshot(active.ID())
	if state.Scores[7] != 5 {
		t.Fatalf("blob score=%d, want 5", state.Scores[7])
	}
	blob, _ = active.FurnitureItem(1)
	if blob.ExtraData != "1" {
		t.Fatalf("consumed blob state=%q", blob.ExtraData)
	}
	if err := coordinator.End(context.Background(), active.ID()); err != nil {
		t.Fatal(err)
	}
	board, _ := active.FurnitureItem(2)
	if highscores.calls != 1 || board.ExtraData == "{}" {
		t.Fatalf("highscore calls=%d board=%q", highscores.calls, board.ExtraData)
	}
	if err := coordinator.Start(context.Background(), active.ID()); err != nil {
		t.Fatal(err)
	}
	if err := coordinator.Reset(context.Background(), active.ID()); err != nil {
		t.Fatal(err)
	}
	if _, found := games.Snapshot(active.ID()); found {
		t.Fatal("reset retained game state")
	}
	if len(updater.calls) < 4 {
		t.Fatalf("durable state calls=%d", len(updater.calls))
	}
}

// TestCoordinatorRejectsInvalidLifecycleAndBlobInputs verifies absent state remains a no-op.
func TestCoordinatorRejectsInvalidLifecycleAndBlobInputs(t *testing.T) {
	rooms := roomlive.NewRegistry(nil)
	coordinator := NewCoordinator(roomwired.Config{}, New(), &gameHighscores{}, rooms, &gameUpdater{}, nil, nil)
	if err := coordinator.Start(context.Background(), 999); err != ErrGameUnavailable {
		t.Fatalf("start error=%v", err)
	}
	if err := coordinator.End(context.Background(), 999); err != ErrGameUnavailable {
		t.Fatalf("end error=%v", err)
	}
	if err := coordinator.Reset(context.Background(), 999); err != ErrGameUnavailable {
		t.Fatalf("reset error=%v", err)
	}
	if err := coordinator.Blob(context.Background(), furniturewalkedon.Payload{RoomID: 999, ItemID: 1, PlayerID: 7}); err != nil {
		t.Fatal(err)
	}
	item := worldfurniture.Item{Definition: worldfurniture.Definition{CustomParams: "invalid,false"}}
	points, resets := blobParameters(item)
	if points != 1 || resets {
		t.Fatalf("blob defaults points=%d resets=%t", points, resets)
	}
}

// TestProjectedBoardClampsProtocolScores verifies rankings cannot overflow Nitro integers.
func TestProjectedBoardClampsProtocolScores(t *testing.T) {
	projection := projectedBoard(2, 0, []record.HighscoreEntry{{Score: math.MaxInt64}, {Score: math.MinInt64}})
	if projection.Entries[0].Score != math.MaxInt32 || projection.Entries[1].Score != math.MinInt32 {
		t.Fatalf("projected scores=%+v", projection.Entries)
	}
}

// gameRoom creates an active room with one resettable blob and one classic board.
func gameRoom(t *testing.T) (*roomlive.Registry, *roomlive.Room) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 66, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	items := []worldfurniture.Item{
		{ID: 1, Point: grid.MustPoint(1, 0), ExtraData: "1", Definition: worldfurniture.Definition{SpriteID: 5067, InteractionType: "wf_blob", CustomParams: "5,true", Width: 1, Length: 1}},
		{ID: 2, Point: grid.MustPoint(2, 0), ExtraData: "{}", Definition: worldfurniture.Definition{SpriteID: 5044, InteractionType: "wf_highscore", Width: 1, Length: 1}},
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: items, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	return rooms, active
}
