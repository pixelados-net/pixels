package core

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/niflaot/pixels/internal/permission"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/redis"
)

// startPermissions provides one configurable restriction bypass.
type startPermissions struct{ bypass bool }

// HasPermission returns the configured bypass decision.
func (permissions startPermissions) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return permissions.bypass, nil
}

// startFixture owns one three-player active-room test graph.
type startFixture struct {
	service *Service
	players map[int64]*playerlive.Player
	room    *roomlive.Room
	rooms   *roomlive.Registry
	units   map[int64]int64
}

// newStartFixture creates a live room and direct-trade service.
func newStartFixture(t *testing.T, bypass bool, throttle *redis.Client) *startFixture {
	t.Helper()
	players := playerlive.NewRegistry()
	playerIndex := make(map[int64]*playerlive.Player, 3)
	for playerID := int64(1); playerID <= 3; playerID++ {
		peer, err := playerlive.NewSessionPeer(netconn.ID("trade-"+strconv.FormatInt(playerID, 10)), "websocket", time.Now())
		if err != nil {
			t.Fatal(err)
		}
		player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: playerID, Username: "player-" + strconv.FormatInt(playerID, 10), AllowTrade: true}, peer)
		if err != nil || players.Add(player) != nil {
			t.Fatalf("player=%d err=%v", playerID, err)
		}
		playerIndex[playerID] = player
	}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 5, TradeMode: 2})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil || active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}) != nil {
		t.Fatalf("world err=%v", err)
	}
	units := make(map[int64]int64, 3)
	for playerID := int64(1); playerID <= 3; playerID++ {
		_, err = rooms.Join(context.Background(), 9, roomlive.Occupant{PlayerID: playerID, Username: "player", ConnectionID: netconn.ID("trade-" + strconv.FormatInt(playerID, 10)), ConnectionKind: "websocket"})
		unit, found := active.Unit(playerID)
		if err != nil || !found {
			t.Fatalf("join player=%d err=%v", playerID, err)
		}
		units[playerID] = unit.UnitID
	}
	service := New(Options{Enabled: true, StartThrottle: 10 * time.Second, MaximumItems: 12}, traderuntime.NewRegistry(), players, rooms, nil, nil, nil, nil, startPermissions{bypass: bypass}, throttle, nil)
	fixture := &startFixture{service: service, players: playerIndex, room: active, rooms: rooms, units: units}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), 9) })
	return fixture
}

// hasTradeStatus reports whether one stable unit snapshot exposes trading.
func hasTradeStatus(unit roomlive.UnitSnapshot) bool {
	for _, status := range unit.Statuses {
		if status.Key == worldunit.StatusTrade {
			return true
		}
	}
	return false
}

// TestStartValidationRejectsEachParticipantBoundary verifies ordered start gates.
func TestStartValidationRejectsEachParticipantBoundary(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(*startFixture) int64
		want    error
	}{
		{name: "missing target", prepare: func(*startFixture) int64 { return 999 }, want: ErrUnavailable},
		{name: "global disabled", prepare: func(fixture *startFixture) int64 { fixture.service.config.Enabled = false; return fixture.units[2] }, want: ErrDisabled},
		{name: "room disabled", prepare: func(fixture *startFixture) int64 {
			fixture.room.UpdateCategoryAndTrade(nil, 0)
			return fixture.units[2]
		}, want: ErrRoomPolicy},
		{name: "ignored", prepare: func(fixture *startFixture) int64 { fixture.players[2].Ignore(1); return fixture.units[2] }, want: ErrIgnored},
		{name: "actor busy", prepare: func(fixture *startFixture) int64 {
			fixture.service.registry.Start(&traderuntime.Session{First: traderuntime.Participant{PlayerID: 1}, Second: traderuntime.Participant{PlayerID: 99}})
			return fixture.units[2]
		}, want: ErrUnavailable},
		{name: "actor locked", prepare: func(fixture *startFixture) int64 { fixture.players[1].SetAllowTrade(false); return fixture.units[2] }, want: ErrActorNotAllowed},
		{name: "target busy", prepare: func(fixture *startFixture) int64 {
			fixture.service.registry.Start(&traderuntime.Session{First: traderuntime.Participant{PlayerID: 2}, Second: traderuntime.Participant{PlayerID: 99}})
			return fixture.units[2]
		}, want: ErrUnavailable},
		{name: "target locked", prepare: func(fixture *startFixture) int64 { fixture.players[2].SetAllowTrade(false); return fixture.units[2] }, want: ErrTargetNotAllowed},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newStartFixture(t, false, nil)
			_, err := fixture.service.Start(context.Background(), 1, testCase.prepare(fixture), "127.0.0.1:3000")
			if !errors.Is(err, testCase.want) {
				t.Fatalf("got %v want %v", err, testCase.want)
			}
		})
	}
}

// TestStartHonorsControllerBypassAndStatusLifecycle verifies room semantics and projection cleanup.
func TestStartHonorsControllerBypassAndStatusLifecycle(t *testing.T) {
	fixture := newStartFixture(t, true, nil)
	fixture.service.config.Enabled = false
	fixture.room.UpdateCategoryAndTrade(nil, 1)
	session, err := fixture.service.Start(context.Background(), 2, fixture.units[1], "127.0.0.1:3000")
	first, _ := fixture.room.Unit(1)
	second, _ := fixture.room.Unit(2)
	if err != nil || session == nil || !hasTradeStatus(first) || !hasTradeStatus(second) {
		t.Fatalf("session=%#v first=%#v second=%#v err=%v", session, first.Statuses, second.Statuses, err)
	}
	if !fixture.service.Close(2) {
		t.Fatal("close")
	}
	first, _ = fixture.room.Unit(1)
	second, _ = fixture.room.Unit(2)
	if hasTradeStatus(first) || hasTradeStatus(second) {
		t.Fatal("trade status remained after close")
	}
}

// TestStartControllerModeAllowsOnlyOwner verifies the non-bypass controller policy.
func TestStartControllerModeAllowsOnlyOwner(t *testing.T) {
	visitorFixture := newStartFixture(t, false, nil)
	visitorFixture.room.UpdateCategoryAndTrade(nil, 1)
	if _, err := visitorFixture.service.Start(context.Background(), 2, visitorFixture.units[1], "127.0.0.1"); !errors.Is(err, ErrRoomPolicy) {
		t.Fatalf("visitor got %v want %v", err, ErrRoomPolicy)
	}

	ownerFixture := newStartFixture(t, false, nil)
	ownerFixture.room.UpdateCategoryAndTrade(nil, 1)
	if _, err := ownerFixture.service.Start(context.Background(), 1, ownerFixture.units[2], "127.0.0.1"); err != nil {
		t.Fatalf("owner got %v", err)
	}
}

// TestStartThrottlePrecedesBusySessionValidation verifies distributed retry ordering.
func TestStartThrottlePrecedesBusySessionValidation(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.New(redis.Config{Address: server.Addr()})
	defer client.Close()
	fixture := newStartFixture(t, false, client)
	if _, err := fixture.service.Start(context.Background(), 1, fixture.units[2], "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.Start(context.Background(), 1, fixture.units[2], "127.0.0.1"); !errors.Is(err, ErrThrottled) {
		t.Fatalf("got %v", err)
	}
}
