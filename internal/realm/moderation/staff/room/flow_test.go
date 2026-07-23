package room

import (
	"context"
	"errors"
	"testing"
	"time"

	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inchange "github.com/niflaot/pixels/networking/inbound/moderation/staff/changeroom"
	inalert "github.com/niflaot/pixels/networking/inbound/moderation/staff/roomalert"
	outaction "github.com/niflaot/pixels/networking/outbound/moderation/staff/actionresult"
	outupdated "github.com/niflaot/pixels/networking/outbound/room/settings/updated"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outdesktop "github.com/niflaot/pixels/networking/outbound/session/desktop"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// eventsForTest captures room settings events.
type eventsForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish captures one event.
func (events *eventsForTest) Publish(_ context.Context, event bus.Event) error {
	events.events = append(events.events, event)
	return nil
}

// flowFixture stores a complete live moderator room scenario.
type flowFixture struct {
	// handler executes tested packet behavior.
	handler Handler
	// source identifies the moderator connection.
	source netconn.Context
	// active stores the live room.
	active *roomlive.Room
	// rooms stores durable room state.
	rooms *roomsForTest
	// players stores live players.
	players *playerlive.Registry
	// permissions stores mutable test grants.
	permissions permissionsForTest
	// events captures settings changes.
	events *eventsForTest
	// sent stores packets by player id.
	sent map[int64]*[]codec.Packet
}

// TestAlertBroadcastsCurrentRoomWithoutActionResult verifies the fixed Nitro flow.
func TestAlertBroadcastsCurrentRoomWithoutActionResult(t *testing.T) {
	fixture := newFlowFixture(t)
	packet := alertPacket(t, alertAction, "  Causitas  ")
	if err := fixture.handler.alert(fixture.source, packet); err != nil {
		t.Fatalf("room alert: %v", err)
	}
	for _, playerID := range []int64{7, 8} {
		packets := *fixture.sent[playerID]
		if !hasHeader(packets, outalert.Header) || hasHeader(packets, outaction.Header) {
			t.Fatalf("player %d packets=%#v", playerID, packets)
		}
	}
}

// TestChangeAppliesEveryRoomOverride verifies persistence, refresh, and safe mass kick.
func TestChangeAppliesEveryRoomOverride(t *testing.T) {
	fixture := newFlowFixture(t)
	packet, err := codec.NewPacket(inchange.Header, inchange.Definition, codec.Int32(9), codec.Int32(1), codec.Int32(1), codec.Int32(1))
	if err != nil {
		t.Fatalf("encode change: %v", err)
	}
	if err = fixture.handler.change(fixture.source, packet); err != nil {
		t.Fatalf("change room: %v", err)
	}
	if fixture.rooms.room.DoorMode != roommodel.DoorModeDoorbell || fixture.rooms.room.Name != "Sala bajo revisión" {
		t.Fatalf("unexpected room: %+v", fixture.rooms.room)
	}
	if fixture.active.Occupancy().Count != 1 || len(fixture.events.events) != 2 || fixture.events.events[0].Name != "room.settings_updated" {
		t.Fatalf("occupancy=%+v events=%+v", fixture.active.Occupancy(), fixture.events.events)
	}
	if !hasHeader(*fixture.sent[7], outupdated.Header) {
		t.Fatalf("moderator packets=%#v", *fixture.sent[7])
	}
	if !hasHeader(*fixture.sent[8], outdesktop.Header) || !hasHeader(*fixture.sent[8], outalert.Header) {
		t.Fatalf("guest packets=%#v", *fixture.sent[8])
	}
}

// TestAlertRejectsInvalidRuntimeStates verifies packet and presence failures.
func TestAlertRejectsInvalidRuntimeStates(t *testing.T) {
	t.Run("header", func(t *testing.T) {
		fixture := newFlowFixture(t)
		err := fixture.handler.alert(fixture.source, codec.Packet{Header: inalert.Header + 1})
		if !errors.Is(err, codec.ErrUnexpectedHeader) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("action", func(t *testing.T) {
		fixture := newFlowFixture(t)
		if err := fixture.handler.alert(fixture.source, alertPacket(t, 2, "x")); !errors.Is(err, errInvalidAction) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("permission", func(t *testing.T) {
		fixture := newFlowFixture(t)
		fixture.permissions[7][moderationpolicy.RoomOverride] = false
		if err := fixture.handler.alert(fixture.source, alertPacket(t, alertAction, "x")); !errors.Is(err, errUnauthorized) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("offline", func(t *testing.T) {
		fixture := newFlowFixture(t)
		fixture.players.Remove(7)
		if err := fixture.handler.alert(fixture.source, alertPacket(t, alertAction, "x")); !errors.Is(err, errActorOffline) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("outside", func(t *testing.T) {
		fixture := newFlowFixture(t)
		actor, _ := fixture.players.Find(7)
		actor.LeaveRoom()
		if err := fixture.handler.alert(fixture.source, alertPacket(t, alertAction, "x")); !errors.Is(err, errActorOutsideRoom) {
			t.Fatalf("err=%v", err)
		}
	})
}

// TestChangeHandlesNoopAndMissingRoom verifies independent settings packet behavior.
func TestChangeHandlesNoopAndMissingRoom(t *testing.T) {
	fixture := newFlowFixture(t)
	packet, _ := codec.NewPacket(inchange.Header, inchange.Definition, codec.Int32(9), codec.Int32(0), codec.Int32(0), codec.Int32(0))
	if err := fixture.handler.change(fixture.source, packet); err != nil {
		t.Fatalf("noop change: %v", err)
	}
	if len(fixture.events.events) != 0 {
		t.Fatalf("unexpected events: %+v", fixture.events.events)
	}
	fixture.rooms.room = roommodel.Room{}
	if err := fixture.handler.change(fixture.source, packet); !errors.Is(err, roomservice.ErrRoomNotFound) {
		t.Fatalf("missing room err=%v", err)
	}
}

// newFlowFixture creates two live occupants and all room moderation dependencies.
func newFlowFixture(t *testing.T) *flowFixture {
	t.Helper()
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	connections := netconn.NewRegistry()
	liveRooms := roomlive.NewRegistry(nil)
	active, err := liveRooms.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 8, MaxUsers: 10})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	t.Cleanup(func() { _, _, _ = liveRooms.Close(context.Background(), 9) })
	sent := map[int64]*[]codec.Packet{}
	for _, identity := range []struct {
		id   int64
		name string
	}{{7, "Moderator"}, {8, "Guest"}} {
		connectionID := netconn.ID(identity.name)
		peer, _ := playerlive.NewSessionPeer(connectionID, "websocket", time.Now())
		player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: identity.id, Username: identity.name}, peer)
		_ = player.EnterRoom(9)
		_ = players.Add(player)
		_ = bindings.Add(binding.Binding{PlayerID: identity.id, ConnectionID: connectionID, ConnectionKind: "websocket"})
		packets := []codec.Packet{}
		sent[identity.id] = &packets
		outbound := netconn.NewHandlerRegistry()
		outbound.SetFallback(func(_ netconn.Context, packet codec.Packet) error {
			packets = append(packets, packet)
			return nil
		}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
		session, _ := netconn.NewSession(netconn.SessionConfig{ID: connectionID, Kind: "websocket", Outbound: outbound, Sender: func(context.Context, codec.Packet) error { return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
		_ = connections.Register(session)
		_, _ = liveRooms.Join(context.Background(), 9, roomlive.Occupant{PlayerID: identity.id, Username: identity.name, ConnectionID: connectionID, ConnectionKind: "websocket"})
	}
	permissions := permissionsForTest{7: {moderationpolicy.RoomOverride: true}}
	rooms := &roomsForTest{room: roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}, Version: sharedmodel.Version{Version: 2}}, OwnerPlayerID: 8, Name: "Original", DoorMode: roommodel.DoorModeOpen}}
	events := &eventsForTest{}
	runtime := &moderationruntime.Context{Moderation: moderationcore.New(moderationconfig.Config{Enabled: true}, nil, nil, nil, nil, nil, nil), Players: players, Rooms: rooms, RoomsLive: liveRooms, Bindings: bindings, Connections: connections, Permissions: permissions, Translations: translationForTest{}, Events: events}
	return &flowFixture{handler: Handler{Context: runtime}, source: netconn.Context{ConnectionID: "Moderator", ConnectionKind: "websocket"}, active: active, rooms: rooms, players: players, permissions: permissions, events: events, sent: sent}
}

// alertPacket creates one Nitro moderator action packet.
func alertPacket(t *testing.T, action int32, message string) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inalert.Header, inalert.Definition, codec.Int32(action), codec.String(message), codec.String(""))
	if err != nil {
		t.Fatalf("encode alert: %v", err)
	}
	return packet
}

// hasHeader reports whether one packet list contains a header.
func hasHeader(packets []codec.Packet, header uint16) bool {
	for _, packet := range packets {
		if packet.Header == header {
			return true
		}
	}
	return false
}
