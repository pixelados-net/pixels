package settings

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incamerafollow "github.com/niflaot/pixels/networking/inbound/user/settings/camerafollow"
	inhomeroom "github.com/niflaot/pixels/networking/inbound/user/settings/homeroom"
	inoldchat "github.com/niflaot/pixels/networking/inbound/user/settings/oldchat"
	involume "github.com/niflaot/pixels/networking/inbound/user/settings/volume"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outhomeroom "github.com/niflaot/pixels/networking/outbound/user/settings/homeroom"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// settingsRooms returns one deterministic home-room target.
type settingsRooms struct {
	room  roommodel.Room
	found bool
}

// FindByID returns the configured room.
func (rooms settingsRooms) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return rooms.room, rooms.found, nil
}

// settingsRights returns one invisible-room visibility decision.
type settingsRights bool

// HasRights returns the configured visibility decision.
func (rights settingsRights) HasRights(context.Context, int64, int64) (bool, error) {
	return bool(rights), nil
}

// TestSettingsHandlersProjectClientMutations verifies volume, old-chat, and home-room adapters.
func TestSettingsHandlersProjectClientMutations(t *testing.T) {
	store := &memoryStore{}
	service := New(store)
	writer := NewWriter(service, nil, Config{FlushInterval: time.Second, PendingLimit: 2})
	handler, connection, packets, player := settingsFixture(t, service)
	handler.Writer = writer
	volumePacket, _ := codec.NewPacket(involume.Header, involume.Definition, codec.Int32(10), codec.Int32(20), codec.Int32(30))
	if err := handler.volume(connection, volumePacket); err != nil || player.Snapshot().VolumeSystem != 10 {
		t.Fatalf("volume snapshot=%#v err=%v", player.Snapshot(), err)
	}
	oldPacket, _ := codec.NewPacket(inoldchat.Header, inoldchat.Definition, codec.Bool(true))
	if err := handler.oldChat(connection, oldPacket); err != nil || !player.Snapshot().OldChat {
		t.Fatalf("old chat snapshot=%#v err=%v", player.Snapshot(), err)
	}
	cameraPacket, _ := codec.NewPacket(incamerafollow.Header, incamerafollow.Definition, codec.Bool(true))
	if err := handler.cameraFollow(connection, cameraPacket); err != nil || !player.Snapshot().CameraFollowBlocked {
		t.Fatalf("camera snapshot=%#v err=%v", player.Snapshot(), err)
	}
	writer.flush(context.Background())
	roomPacket, _ := codec.NewPacket(inhomeroom.Header, inhomeroom.Definition, codec.Int32(9))
	if err := handler.homeRoom(connection, roomPacket); err != nil || lastSettingsHeader(*packets) != outhomeroom.Header || player.Snapshot().HomeRoomID == nil {
		t.Fatalf("home snapshot=%#v packets=%#v err=%v", player.Snapshot(), *packets, err)
	}
}

// TestSettingsHandlersFallbackAndValidation verifies synchronous persistence and feedback.
func TestSettingsHandlersFallbackAndValidation(t *testing.T) {
	service := New(&memoryStore{})
	handler, connection, packets, player := settingsFixture(t, service)
	volumePacket, _ := codec.NewPacket(involume.Header, involume.Definition, codec.Int32(1), codec.Int32(2), codec.Int32(3))
	if err := handler.volume(connection, volumePacket); err != nil || player.Snapshot().VolumeTrax != 3 {
		t.Fatalf("volume snapshot=%#v err=%v", player.Snapshot(), err)
	}
	oldPacket, _ := codec.NewPacket(inoldchat.Header, inoldchat.Definition, codec.Bool(true))
	if err := handler.oldChat(connection, oldPacket); err != nil || !player.Snapshot().OldChat {
		t.Fatalf("old chat snapshot=%#v err=%v", player.Snapshot(), err)
	}
	invalid, _ := codec.NewPacket(involume.Header, involume.Definition, codec.Int32(-1), codec.Int32(0), codec.Int32(0))
	if err := handler.volume(connection, invalid); err != nil || lastSettingsHeader(*packets) != outalert.Header {
		t.Fatalf("invalid volume packets=%#v err=%v", *packets, err)
	}
	handler.Rooms = settingsRooms{}
	homePacket, _ := codec.NewPacket(inhomeroom.Header, inhomeroom.Definition, codec.Int32(99))
	if err := handler.homeRoom(connection, homePacket); err != nil || lastSettingsHeader(*packets) != outalert.Header {
		t.Fatalf("invalid home packets=%#v err=%v", *packets, err)
	}
}

// TestSettingsRegistrationAndVisibility verifies registration and invisible-room policy.
func TestSettingsRegistrationAndVisibility(t *testing.T) {
	RegisterHandlers(nil, Handler{})
	registry := netconn.NewHandlerRegistry()
	RegisterHandlers(registry, Handler{})
	if registry.Len() != 4 {
		t.Fatalf("handlers=%d", registry.Len())
	}
	invisible := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 8, DoorMode: roommodel.DoorModeInvisible}
	handler := Handler{Rooms: settingsRooms{room: invisible, found: true}}
	if err := handler.validateHomeRoom(7, 9); err != ErrInvalidHomeRoom {
		t.Fatalf("expected hidden room rejection, got %v", err)
	}
	handler.Rights = settingsRights(true)
	if err := handler.validateHomeRoom(7, 9); err != nil {
		t.Fatalf("rights home room: %v", err)
	}
	if _, _, found := (Handler{}).player(netconn.Context{}); found {
		t.Fatal("expected missing player")
	}
}

// settingsFixture creates one authenticated packet-capturing settings handler.
func settingsFixture(t *testing.T, service *Service) (Handler, netconn.Context, *[]codec.Packet, *playerlive.Player) {
	t.Helper()
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(ctx netconn.Context, _ codec.Packet) error { connection = ctx; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 3)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "settings", Kind: "websocket", Inbound: inbound, Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})
	peer, _ := playerlive.NewSessionPeer("settings", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	players := playerlive.NewRegistry()
	_ = players.Add(player)
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "settings", ConnectionKind: "websocket"})
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7, DoorMode: roommodel.DoorModeOpen}
	return Handler{Service: service, Bindings: bindings, Players: players, Rooms: settingsRooms{room: room, found: true}}, connection, &packets, player
}

// lastSettingsHeader returns the latest captured packet header.
func lastSettingsHeader(packets []codec.Packet) uint16 {
	if len(packets) == 0 {
		return 0
	}
	return packets[len(packets)-1].Header
}
