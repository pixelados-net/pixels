package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// teleportRooms stores one room fixture.
type teleportRooms struct {
	// room stores the returned room.
	room roommodel.Room
}

// Create creates no room in teleport tests.
func (rooms teleportRooms) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID returns the room fixture.
func (rooms teleportRooms) FindByID(_ context.Context, id int64) (roommodel.Room, bool, error) {
	return rooms.room, rooms.room.ID == id, nil
}

// ListByOwner returns no room fixtures.
func (teleportRooms) ListByOwner(context.Context, int64) ([]roommodel.Room, error) { return nil, nil }

// ListPopular returns no room fixtures.
func (teleportRooms) ListPopular(context.Context, int) ([]roommodel.Room, error) { return nil, nil }

// ListHighestScore returns no room fixtures.
func (teleportRooms) ListHighestScore(context.Context, int) ([]roommodel.Room, error) {
	return nil, nil
}

// Search returns no room fixtures.
func (teleportRooms) Search(context.Context, string, int) ([]roommodel.Room, error) { return nil, nil }

// ListTags returns no room tags.
func (teleportRooms) ListTags(context.Context, int64) ([]roommodel.Tag, error) { return nil, nil }

// SoftDelete performs no deletion.
func (teleportRooms) SoftDelete(context.Context, int64) error { return nil }

// ListCategories returns no room categories.
func (teleportRooms) ListCategories(context.Context) ([]roommodel.Category, error) { return nil, nil }

// TestTeleportHandlerForwardsAndGrantsOneBypass verifies admin teleport behavior.
func TestTeleportHandlerForwardsAndGrantsOneBypass(t *testing.T) {
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7, DoorMode: roommodel.DoorModePassword}
	entry := roomentry.New(roomentry.Config{TrustedTTL: time.Minute}, nil, nil, nil, roomentry.Nodes{})
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("player", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 8, Username: "Guest"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	connections := netconn.NewRegistry()
	sent := registerTeleportConnection(t, connections)
	app := fiber.New()
	Register(app, teleportRooms{room: room}, roomlive.NewRegistry(nil), connections, nil, players, entry, Dependencies{})
	request := httptest.NewRequest(http.MethodPost, "/api/admin/rooms/players/8/teleport", strings.NewReader(`{"targetRoomId":9,"bypass":true}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if response.StatusCode != http.StatusOK || len(*sent) != 1 || (*sent)[0].Header != outforward.Header {
		t.Fatalf("unexpected response=%d packets=%#v", response.StatusCode, *sent)
	}
	if _, err := entry.Authorize(context.Background(), roomentry.Request{Room: room, PlayerID: 8}); err != nil {
		t.Fatalf("consume trusted entry: %v", err)
	}
	if _, err := entry.Authorize(context.Background(), roomentry.Request{Room: room, PlayerID: 8}); !errors.Is(err, roomentry.ErrWrongPassword) {
		t.Fatalf("expected one-time bypass, got %v", err)
	}
}

// registerTeleportConnection registers one captured player connection.
func registerTeleportConnection(t *testing.T, connections *netconn.Registry) *[]codec.Packet {
	t.Helper()
	sent := make([]codec.Packet, 0, 1)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "player", Kind: "websocket", Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := connections.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	return &sent
}
