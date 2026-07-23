package games

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	gamesconfig "github.com/niflaot/pixels/internal/realm/room/world/games/config"
	"github.com/niflaot/pixels/internal/realm/room/world/games/freeze"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	wiredgame "github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestMergeFootballKitPreservesIdentity verifies only football clothing slots are replaced.
func TestMergeFootballKitPreservesIdentity(t *testing.T) {
	actual := mergeFootballKit("hd-180-1.hr-828-45.ch-210-66.lg-270-82.sh-290-80.ha-1002-61", "ch-255-66.ca-1808-66.lg-275-66.sh-295-66")
	expected := "hd-180-1.hr-828-45.ha-1002-61.ch-255-66.ca-1808-66.lg-275-66.sh-295-66"
	if actual != expected {
		t.Fatalf("figure=%q expected=%q", actual, expected)
	}
}

// TestFreezeMatchOverRequiresLastParticipatingTeam verifies automatic match completion.
func TestFreezeMatchOverRequiresLastParticipatingTeam(t *testing.T) {
	players := map[int64]*freeze.Player{1: {Team: 1, Lives: 1}, 2: {Team: 2, Lives: 1}}
	if freezeMatchOver(players) {
		t.Fatal("two live teams ended")
	}
	players[2].Lives = 0
	if !freezeMatchOver(players) {
		t.Fatal("last team did not end")
	}
	delete(players, 2)
	if freezeMatchOver(players) {
		t.Fatal("single-team match ended as last-team-standing")
	}
}

// TestFootballCounterModes verifies increment, decrement, reset, and wrap.
func TestFootballCounterModes(t *testing.T) {
	if footballCounterNext(99, 0) != 0 || footballCounterNext(0, 1) != 99 || footballCounterNext(42, 2) != 0 {
		t.Fatal("invalid football counter mode")
	}
}

// TestMergeFootballKitIgnoresIdentityInjection verifies kits cannot replace identity parts.
func TestMergeFootballKitIgnoresIdentityInjection(t *testing.T) {
	actual := mergeFootballKit("hd-180-1.hr-828-45", "hd-999-9.ch-255-66")
	if actual != "hd-180-1.hr-828-45.ch-255-66" {
		t.Fatalf("unexpected figure %q", actual)
	}
}

// TestFootballKitGateTogglesTheRoomProjection verifies repeated crossings equip then restore.
func TestFootballKitGateTogglesTheRoomProjection(t *testing.T) {
	registry := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := registry.Activate(roomlive.Snapshot{ID: 5, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	gate := worldfurniture.Item{ID: 9, Point: grid.MustPoint(1, 0), Definition: worldfurniture.Definition{InteractionType: "football_gate", CustomParams: "ch-255-66.lg-275-66", Width: 1, Length: 1, AllowWalk: true}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{gate}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	original := "hd-180-1.hr-828-45.ch-210-66.lg-270-82"
	if _, err = registry.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, Figure: original, Gender: "M", ConnectionID: netconn.ID("7"), ConnectionKind: netconn.Kind("test")}); err != nil {
		t.Fatal(err)
	}
	service := &Service{config: gamesconfig.Config{Enabled: true}, rooms: registry, states: make(map[int64]*roomState)}
	if err = service.WalkedOn(context.Background(), bus.Event{Payload: furniturewalkedon.Payload{RoomID: active.ID(), ItemID: gate.ID, PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if occupant, _ := active.Occupant(7); occupant.Figure == original {
		t.Fatal("kit was not projected")
	}
	if err = service.WalkedOff(context.Background(), bus.Event{Payload: furniturewalkedoff.Payload{RoomID: active.ID(), ItemID: gate.ID, PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if occupant, _ := active.Occupant(7); occupant.Figure == original {
		t.Fatal("walking off the gate removed the kit")
	}
	if err = service.WalkedOn(context.Background(), bus.Event{Payload: furniturewalkedon.Payload{RoomID: active.ID(), ItemID: gate.ID, PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if occupant, _ := active.Occupant(7); occupant.Figure != original {
		t.Fatalf("restored figure=%q", occupant.Figure)
	}
	_, _, _ = registry.Close(context.Background(), active.ID())
}

// TestWinningTeamRejectsTies verifies match rewards require one sole team.
func TestWinningTeamRejectsTies(t *testing.T) {
	if team := winningTeam(wiredState(map[int64]int32{1: 1, 2: 2}, map[int32]int64{1: 4, 2: 4})); team != 0 {
		t.Fatalf("tie winner=%d", team)
	}
	if team := winningTeam(wiredState(map[int64]int32{1: 1, 2: 2}, map[int32]int64{1: 5, 2: 4})); team != 1 {
		t.Fatalf("winner=%d", team)
	}
}

// wiredState builds one minimal shared game snapshot.
func wiredState(teams map[int64]int32, scores map[int32]int64) wiredgame.State {
	state := wiredgame.State{Teams: teams}
	for team, score := range scores {
		state.TeamScores[team] = score
	}
	return state
}

// scoreCapture records completed room-game persistence.
type scoreCapture struct {
	// entries stores the last persisted match.
	entries []Score
}

// Save records one completed match.
func (capture *scoreCapture) Save(_ context.Context, entries []Score) error {
	capture.entries = append([]Score(nil), entries...)
	return nil
}

// List returns an empty history page.
func (capture *scoreCapture) List(context.Context, int64, int64, int) (ScorePage, error) {
	return ScorePage{}, nil
}

// TestBanzaiLifecycleRunsEndToEnd verifies gates, timer, tile completion, score persistence, and metrics.
func TestBanzaiLifecycleRunsEndToEnd(t *testing.T) {
	registry := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := registry.Activate(roomlive.Snapshot{ID: 8, OwnerPlayerID: 1, MaxUsers: 4})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00000\n00000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	items := []worldfurniture.Item{
		{ID: 10, Point: grid.MustPoint(1, 0), Definition: worldfurniture.Definition{InteractionType: "battlebanzai_gate_r", Width: 1, Length: 1, AllowWalk: true}},
		{ID: 11, Point: grid.MustPoint(2, 0), Definition: worldfurniture.Definition{InteractionType: "battlebanzai_gate_g", Width: 1, Length: 1, AllowWalk: true}},
		{ID: 12, Point: grid.MustPoint(3, 0), Definition: worldfurniture.Definition{InteractionType: "game_timer", CustomParams: "1", Width: 1, Length: 1, AllowWalk: true}},
		{ID: 13, Point: grid.MustPoint(1, 1), Definition: worldfurniture.Definition{InteractionType: "battlebanzai_tile", Width: 1, Length: 1, AllowWalk: true}},
		{ID: 14, Point: grid.MustPoint(2, 1), Definition: worldfurniture.Definition{InteractionType: "football", Width: 1, Length: 1, AllowWalk: true}},
		{ID: 15, Point: grid.MustPoint(3, 1), Definition: worldfurniture.Definition{InteractionType: "battlebanzai_random_teleport", Width: 1, Length: 1, AllowWalk: true}},
		{ID: 16, Point: grid.MustPoint(4, 1), Definition: worldfurniture.Definition{InteractionType: "battlebanzai_random_teleport", Width: 1, Length: 1, AllowWalk: true}},
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: items, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	for _, playerID := range []int64{1, 2} {
		if _, err = registry.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: playerID, Figure: "hd-180-1", ConnectionID: netconn.ID("player"), ConnectionKind: netconn.Kind("test")}); err != nil {
			t.Fatal(err)
		}
	}
	shared := wiredgame.New()
	coordinator := wiredgame.NewCoordinator(roomwired.Config{}, shared, nil, registry, nil, connections, nil)
	scores := &scoreCapture{}
	local := bus.New()
	metrics := NewMetrics()
	service := New(gamesconfig.Config{Enabled: true, Banzai: gamesconfig.Banzai{PointsLock: 1}}, registry, shared, coordinator, connections, nil, scores, bus.NewPublisher(local), metrics)
	for _, entry := range []struct {
		// itemID identifies the crossed gate.
		itemID int64
		// playerID identifies the entrant.
		playerID int64
	}{{10, 1}, {10, 2}} {
		if err = service.WalkedOn(context.Background(), bus.Event{Name: "test", Payload: furniturewalkedon.Payload{RoomID: active.ID(), ItemID: entry.itemID, PlayerID: entry.playerID}}); err != nil {
			t.Fatal(err)
		}
	}
	timerItem, _ := active.FurnitureItem(12)
	if _, err = service.UseFurniture(context.Background(), UseRequest{PlayerID: 1, Room: active, Item: timerItem}); err != nil {
		t.Fatal(err)
	}
	if err = service.WalkedOn(context.Background(), bus.Event{Name: "test", Payload: furniturewalkedon.Payload{RoomID: active.ID(), ItemID: 11, PlayerID: 1}}); err != nil {
		t.Fatal(err)
	}
	if team, _ := shared.Team(active.ID(), 1); team != 1 {
		t.Fatalf("running gate changed team to %d", team)
	}
	if err = service.scheduleBanzaiTeleport(active, 1, 15); err != nil {
		t.Fatal(err)
	}
	if _, moveErr := active.MoveTo(1, grid.MustPoint(4, 0)); !errors.Is(moveErr, roomlive.ErrUnitExiting) {
		t.Fatalf("teleport freeze move error=%v", moveErr)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	if unit, found := active.Unit(1); !found || unit.Position.Point != grid.MustPoint(4, 1) {
		t.Fatalf("teleported unit=%+v found=%t", unit, found)
	}
	ball, _ := active.FurnitureItem(14)
	var kicks sync.WaitGroup
	for _, playerID := range []int64{1, 2} {
		kicks.Add(1)
		go func() {
			defer kicks.Done()
			for range 100 {
				if kickErr := service.kickFootball(UseRequest{PlayerID: playerID, Room: active, Item: ball}); kickErr != nil {
					t.Error(kickErr)
				}
			}
		}()
	}
	kicks.Wait()
	errors := make(chan error, 6)
	var wait sync.WaitGroup
	for _, playerID := range []int64{1, 2} {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 3 {
				errors <- service.WalkedOn(context.Background(), bus.Event{Name: "test", Payload: furniturewalkedon.Payload{RoomID: active.ID(), ItemID: 13, PlayerID: playerID}})
			}
		}()
	}
	wait.Wait()
	close(errors)
	for eventErr := range errors {
		if eventErr != nil {
			t.Fatal(eventErr)
		}
	}
	if len(scores.entries) != 2 || scores.entries[0].Kind != "banzai" {
		t.Fatalf("scores=%+v", scores.entries)
	}
	snapshot := metrics.Snapshot()
	if snapshot.Started["banzai"] != 1 || snapshot.Ended["banzai"] != 1 || snapshot.TilesLocked != 1 {
		t.Fatalf("metrics=%+v", snapshot)
	}
	service.Close(active.ID())
	_, _, _ = registry.Close(context.Background(), active.ID())
}
