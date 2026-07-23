// Package tests verifies room-game workflows through their public boundary.
package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	roomgames "github.com/niflaot/pixels/internal/realm/room/world/games"
	gamesconfig "github.com/niflaot/pixels/internal/realm/room/world/games/config"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	wiredgame "github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestFreezeBlockClickApproachesAndLaunches verifies Nitro's walk-then-use ordering.
func TestFreezeBlockClickApproachesAndLaunches(t *testing.T) {
	registry := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := registry.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00000\n00000\n00000\n00000", grid.WithDoor(0, 3))
	if err != nil {
		t.Fatal(err)
	}
	tile := worldfurniture.Item{ID: 20, Point: grid.MustPoint(3, 1), Definition: worldfurniture.Definition{InteractionType: "freeze_tile", Width: 1, Length: 1, AllowStack: true, AllowWalk: true}}
	block := worldfurniture.Item{ID: 21, Point: tile.Point, Z: 1, Definition: worldfurniture.Definition{InteractionType: "freeze_block", Width: 1, Length: 1, AllowStack: true}}
	timer := worldfurniture.Item{ID: 22, Point: grid.MustPoint(1, 1), Definition: worldfurniture.Definition{InteractionType: "game_timer", CustomParams: "30", Width: 1, Length: 1, AllowWalk: true}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{tile, block, timer}, Door: worldpath.Position{Point: grid.MustPoint(0, 3)}}); err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	if _, err = registry.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 1, Figure: "hd-180-1", ConnectionID: "one", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	shared := wiredgame.NewProjected(registry, connections)
	shared.JoinTeam(active.ID(), 1, 1)
	coordinator := wiredgame.NewCoordinator(roomwired.Config{}, shared, nil, registry, nil, connections, nil)
	metrics := roomgames.NewMetrics()
	service := roomgames.New(gamesconfig.Config{Enabled: true, Freeze: gamesconfig.Freeze{MaxLives: 3, MaxSnowballs: 5, FrozenDuration: time.Second, PowerupChance: 100}}, registry, shared, coordinator, connections, nil, nil, nil, metrics)
	if _, err = service.UseFurniture(context.Background(), roomgames.UseRequest{PlayerID: 1, Room: active, Item: timer}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.UseFurniture(context.Background(), roomgames.UseRequest{PlayerID: 1, Room: active, Item: block}); err != nil {
		t.Fatal(err)
	}
	for range 4 {
		active.Tick()
	}
	active.RunScheduled(time.Now().Add(5 * time.Second))
	updated, _ := active.FurnitureItem(tile.ID)
	if metrics.Snapshot().FreezeBalls != 1 || updated.ExtraData != "2000" {
		t.Fatalf("freeze balls=%d tile state=%q", metrics.Snapshot().FreezeBalls, updated.ExtraData)
	}
	if err = service.Cycle(context.Background(), active, time.Now().Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	active.RunScheduled(time.Now().Add(5 * time.Second))
	broken, _ := active.FurnitureItem(block.ID)
	brokenState, parseErr := strconv.Atoi(broken.ExtraData)
	if parseErr != nil || brokenState < 2000 || brokenState > 7000 {
		t.Fatalf("broken block state=%q error=%v", broken.ExtraData, parseErr)
	}
	if _, moveErr := active.MoveTo(1, block.Point); moveErr != nil {
		t.Fatalf("broken block remained impassable: %v", moveErr)
	}
	if err = service.WalkedOn(context.Background(), bus.Event{Payload: furniturewalkedon.Payload{RoomID: active.ID(), ItemID: block.ID, PlayerID: 1}}); err != nil {
		t.Fatal(err)
	}
	collected, _ := active.FurnitureItem(block.ID)
	collectedState, parseErr := strconv.Atoi(collected.ExtraData)
	if parseErr != nil || collectedState < 12000 || collectedState > 17000 {
		t.Fatalf("collected block state=%q error=%v", collected.ExtraData, parseErr)
	}
	if unit, found := active.Unit(1); !found || unit.ActiveEffectID == 0 {
		t.Fatalf("running Freeze effect=%d found=%v", unit.ActiveEffectID, found)
	}
	if err = service.Cycle(context.Background(), active, time.Now().Add(35*time.Second)); err != nil {
		t.Fatal(err)
	}
	active.RunScheduled(time.Now().Add(40 * time.Second))
	finishedUnit, found := active.Unit(1)
	resetBlock, _ := active.FurnitureItem(block.ID)
	if !found || finishedUnit.ActiveEffectID != 0 || resetBlock.ExtraData != "0" {
		t.Fatalf("finished effect=%d found=%v block=%q", finishedUnit.ActiveEffectID, found, resetBlock.ExtraData)
	}
	_, _, _ = registry.Close(context.Background(), active.ID())
}
