package games

import (
	"context"
	"testing"
	"time"

	furnituremoved "github.com/niflaot/pixels/internal/realm/furniture/events/moved"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	gamesconfig "github.com/niflaot/pixels/internal/realm/room/world/games/config"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	wiredgame "github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// footballMover records authoritative ball placements.
type footballMover struct {
	// moves stores each accepted placement.
	moves []furnitureservice.MoveParams
}

// TestFurnitureMovedUpdatesFootballKickoff verifies room editing replaces a stale load-time origin.
func TestFurnitureMovedUpdatesFootballKickoff(t *testing.T) {
	registry := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := registry.Activate(roomlive.Snapshot{ID: 17, OwnerPlayerID: 1, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	ball := worldfurniture.Item{ID: 50, OwnerPlayerID: 1, Point: grid.MustPoint(1, 0), Definition: worldfurniture.Definition{InteractionType: "football", Width: 1, Length: 1, AllowWalk: true}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{ball}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	service := &Service{config: gamesconfig.Config{Enabled: true}, rooms: registry, states: make(map[int64]*roomState)}
	service.mutex.Lock()
	state := service.stateLocked(active)
	state.footballs[ball.ID] = &footballBall{remaining: 4}
	service.mutex.Unlock()
	ball.Point = grid.MustPoint(2, 0)
	if _, err = active.ReloadFurniture(ball.ID, &ball); err != nil {
		t.Fatal(err)
	}
	event := bus.Event{Name: furnituremoved.Name, Payload: furnituremoved.Payload{RoomID: active.ID(), ItemID: ball.ID, PlayerID: 1, X: 2, Y: 0}}
	if err = service.FurnitureMoved(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	service.mutex.Lock()
	origin := state.footballOrigins[ball.ID]
	remaining := state.footballs[ball.ID].remaining
	service.mutex.Unlock()
	if origin != ball.Point || remaining != 0 {
		t.Fatalf("origin=%+v remaining=%d", origin, remaining)
	}
	_, _, _ = registry.Close(context.Background(), active.ID())
}

// Move records one accepted ball placement.
func (mover *footballMover) Move(_ context.Context, params furnitureservice.MoveParams) (furnituremodel.Item, error) {
	mover.moves = append(mover.moves, params)
	return furnituremodel.Item{}, nil
}

// TestResolveFootballMoveUsesUnitsAndGoalFace verifies dynamic rebounds and directional goal entry.
func TestResolveFootballMoveUsesUnitsAndGoalFace(t *testing.T) {
	registry := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := registry.Activate(roomlive.Snapshot{ID: 19, OwnerPlayerID: 1, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00000\n00000\n00000\n00000\n00000", grid.WithDoor(3, 2))
	if err != nil {
		t.Fatal(err)
	}
	ball := worldfurniture.Item{ID: 60, OwnerPlayerID: 1, Point: grid.MustPoint(2, 2), Definition: worldfurniture.Definition{InteractionType: "football", Width: 1, Length: 1, AllowStack: true, AllowWalk: true}}
	goal := worldfurniture.Item{ID: 61, OwnerPlayerID: 1, Point: grid.MustPoint(4, 1), Rotation: 6, Definition: worldfurniture.Definition{InteractionType: "football_goal_blue", Width: 3, Length: 1}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{ball, goal}, Door: worldpath.Position{Point: grid.MustPoint(3, 2)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = registry.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 1, Figure: "hd-180-1", ConnectionID: "one", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	service := &Service{}
	target, direction, valid := service.resolveFootballMove(active, ball, 2)
	if !valid || target != grid.MustPoint(1, 2) || direction != 6 {
		t.Fatalf("cardinal rebound target=%+v direction=%d valid=%t", target, direction, valid)
	}
	if _, err = active.TeleportUnit(1, grid.MustPoint(3, 1), worldunit.RotationSouth, false); err != nil {
		t.Fatal(err)
	}
	target, direction, valid = service.resolveFootballMove(active, ball, 1)
	if !valid || target != grid.MustPoint(1, 1) || direction != 7 {
		t.Fatalf("diagonal rebound target=%+v direction=%d valid=%t", target, direction, valid)
	}
	if service.blockedFootball(active, grid.MustPoint(4, 2), 0, 2) {
		t.Fatal("front goal entry was blocked")
	}
	if !service.blockedFootball(active, grid.MustPoint(4, 2), 0, 0) {
		t.Fatal("side goal entry was accepted")
	}
	_, _, _ = registry.Close(context.Background(), active.ID())
}

// TestFootballGoalMovesScoresAndResets verifies the complete native football flow.
func TestFootballGoalMovesScoresAndResets(t *testing.T) {
	registry := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := registry.Activate(roomlive.Snapshot{ID: 12, OwnerPlayerID: 1, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000000000\n000000000\n000000000", grid.WithDoor(3, 1))
	if err != nil {
		t.Fatal(err)
	}
	ball := worldfurniture.Item{ID: 40, OwnerPlayerID: 1, Point: grid.MustPoint(4, 1), Definition: worldfurniture.Definition{InteractionType: "football", Width: 1, Length: 1, AllowWalk: true}}
	goal := worldfurniture.Item{ID: 41, OwnerPlayerID: 1, Point: grid.MustPoint(6, 0), Rotation: 6, Definition: worldfurniture.Definition{InteractionType: "football_goal_blue", Width: 3, Length: 1}}
	counter := worldfurniture.Item{ID: 42, OwnerPlayerID: 1, Point: grid.MustPoint(8, 0), ExtraData: "0", Definition: worldfurniture.Definition{InteractionType: "football_counter_blue", Width: 1, Length: 1}}
	timer := worldfurniture.Item{ID: 43, OwnerPlayerID: 1, Point: grid.MustPoint(8, 2), ExtraData: "30", Definition: worldfurniture.Definition{InteractionType: "game_timer", CustomParams: "30", Width: 1, Length: 1}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{ball, goal, counter, timer}, Door: worldpath.Position{Point: grid.MustPoint(3, 1)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = registry.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 1, Figure: "hd-180-1", ConnectionID: "one", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	shared := wiredgame.New()
	coordinator := wiredgame.NewCoordinator(roomwired.Config{}, shared, nil, registry, nil, connections, nil)
	mover := &footballMover{}
	metrics := NewMetrics()
	service := New(gamesconfig.Config{Enabled: true}, registry, shared, coordinator, connections, nil, nil, nil, metrics)
	service.furniture = mover
	if _, err = service.UseFurniture(context.Background(), UseRequest{PlayerID: 1, Room: active, Item: timer}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.UseFurniture(context.Background(), UseRequest{PlayerID: 1, Room: active, Item: ball}); err != nil {
		t.Fatal(err)
	}
	if err = service.Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err = service.Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatal(err)
	}
	updatedCounter, _ := active.FurnitureItem(counter.ID)
	scoredBall, _ := active.FurnitureItem(ball.ID)
	if updatedCounter.ExtraData != "1" || scoredBall.Point != grid.MustPoint(6, 1) || metrics.Snapshot().FootballGoals != 1 {
		t.Fatalf("counter=%q ball=%+v goals=%d", updatedCounter.ExtraData, scoredBall.Point, metrics.Snapshot().FootballGoals)
	}
	service.queueFootball(active, ball.ID, 6, 1)
	active.RunScheduled(time.Now().Add(time.Second))
	resetBall, _ := active.FurnitureItem(ball.ID)
	if resetBall.Point != ball.Point || len(mover.moves) != 3 {
		t.Fatalf("reset ball=%+v moves=%d", resetBall.Point, len(mover.moves))
	}
	service.mutex.Lock()
	resetState := service.states[active.ID()].footballs[ball.ID]
	service.mutex.Unlock()
	if resetState.remaining != 0 || resetState.resetting {
		t.Fatalf("reset queue=%+v", resetState)
	}
	_, _, _ = registry.Close(context.Background(), active.ID())
}
