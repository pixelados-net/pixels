package enter

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// playerForTest creates a live player.
func playerForTest(t *testing.T) *playerlive.Player {
	t.Helper()

	peer, err := playerlive.NewSessionPeer(netconn.ID("conn"), netconn.Kind("websocket"), time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	return player
}

// playerRegistryForTest creates a registry with one player.
func playerRegistryForTest(t *testing.T, player *playerlive.Player) *playerlive.Registry {
	t.Helper()

	registry := playerlive.NewRegistry()
	if err := registry.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return registry
}

// bindingRegistryForTest creates a connection binding registry.
func bindingRegistryForTest(t *testing.T, playerID int64) *binding.Registry {
	t.Helper()

	registry := binding.NewRegistry()
	err := registry.Add(binding.Binding{PlayerID: playerID, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")})
	if err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return registry
}

// occupantForTest creates an active room occupant.
func occupantForTest(playerID int64) roomlive.Occupant {
	return roomlive.Occupant{
		PlayerID:       playerID,
		Username:       "demo",
		ConnectionID:   netconn.ID("conn"),
		ConnectionKind: netconn.Kind("websocket"),
	}
}

// connectionForTest creates a handler connection context.
func connectionForTest() netconn.Context {
	return netconn.Context{ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
}

// sessionConnectionForTest creates a connection context backed by a session.
func sessionConnectionForTest(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()

	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	inbound := netconn.NewHandlerRegistry()
	var captured netconn.Context
	if err := inbound.Register(1, func(context netconn.Context, _ codec.Packet) error {
		captured = context
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register inbound: %v", err)
	}

	sent := make([]codec.Packet, 0, 2)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       netconn.ID("conn"),
		Kind:     netconn.Kind("websocket"),
		Inbound:  inbound,
		Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive context packet: %v", err)
	}

	return captured, &sent
}

// roomForTest creates a persistent room.
func roomForTest() roommodel.Room {
	return roommodel.Room{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		OwnerPlayerID: 7,
		Name:          "Room",
		ModelName:     "model_a",
		MaxUsers:      25,
	}
}

// layoutForTest creates a persistent layout.
func layoutForTest() layout.Layout {
	return layout.Layout{Name: "model_a", TileSize: 1, Heightmap: "0", DoorX: 0, DoorY: 0, DoorDirection: 2, Enabled: true}
}

// publisherForTest captures published events.
type publisherForTest struct {
	// events stores published events.
	events []bus.Event
	// onPublish observes publication timing when configured.
	onPublish func(bus.Event)
}

// Publish captures an event.
func (publisher *publisherForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	if publisher.onPublish != nil {
		publisher.onPublish(event)
	}

	return nil
}

// roomManagerForTest stubs room persistence.
type roomManagerForTest struct {
	// room stores the returned room.
	room roommodel.Room

	// found reports whether the room exists.
	found bool

	// err stores the returned error.
	err error
}

// Create creates a room record.
func (manager roomManagerForTest) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID finds a room by id.
func (manager roomManagerForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return manager.room, manager.found, manager.err
}

// ListByOwner lists owned rooms.
func (manager roomManagerForTest) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return nil, nil
}

// ListPopular lists popular rooms.
func (manager roomManagerForTest) ListPopular(context.Context, int) ([]roommodel.Room, error) {
	return nil, nil
}

// ListHighestScore lists highest score rooms.
func (manager roomManagerForTest) ListHighestScore(context.Context, int) ([]roommodel.Room, error) {
	return nil, nil
}

// Search searches rooms.
func (manager roomManagerForTest) Search(context.Context, string, int) ([]roommodel.Room, error) {
	return nil, nil
}

// ListTags lists room tags.
func (manager roomManagerForTest) ListTags(context.Context, int64) ([]roommodel.Tag, error) {
	return nil, nil
}

// SoftDelete soft deletes a room.
func (manager roomManagerForTest) SoftDelete(context.Context, int64) error {
	return nil
}

// ListCategories lists room categories.
func (manager roomManagerForTest) ListCategories(context.Context) ([]roommodel.Category, error) {
	return nil, nil
}
