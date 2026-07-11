package save

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/permission"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/settings"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outerror "github.com/niflaot/pixels/networking/outbound/room/settings/error"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// managerForTest stores room settings fixtures.
type managerForTest struct {
	// room stores the current room.
	room roommodel.Room
	// updated stores the committed room.
	updated roommodel.Room
	// updateErr stores an optional mutation failure.
	updateErr error
	// updates counts committed mutations.
	updates int
}

// FindByID returns the room fixture.
func (manager *managerForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return manager.room, manager.room.ID > 0, nil
}

// Update records and returns one room mutation.
func (manager *managerForTest) Update(context.Context, int64, int64, roomservice.UpdateParams) (roommodel.Room, error) {
	manager.updates++
	return manager.updated, manager.updateErr
}

// permissionsForBehaviorTest stores permission fixtures.
type permissionsForBehaviorTest map[permission.Node]bool

// HasPermission returns one permission fixture.
func (permissions permissionsForBehaviorTest) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return permissions[node], nil
}

// publisherForTest captures settings events.
type publisherForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish captures one event.
func (publisher *publisherForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// TestHandleCommitsBroadcastsAndRejectsHC verifies settings orchestration and entitlement enforcement.
func TestHandleCommitsBroadcastsAndRejectsHC(t *testing.T) {
	handler, input, manager, sent, events := behaviorFixture(t)
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); err != nil {
		t.Fatalf("save room settings: %v", err)
	}
	if manager.updates != 1 || len(*sent) != 4 || len(events.events) != 1 {
		t.Fatalf("updates=%d packets=%d events=%d", manager.updates, len(*sent), len(events.events))
	}

	input.HideWalls = true
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); !errors.Is(err, roomsettings.ErrClubRequired) {
		t.Fatalf("expected club rejection, got %v", err)
	}
	if manager.updates != 1 {
		t.Fatalf("club rejection committed update count=%d", manager.updates)
	}
	manager.room = roommodel.Room{}
	input.HideWalls = false
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); err != nil {
		t.Fatalf("send missing room result: %v", err)
	}
	values, err := codec.DecodePacketExact((*sent)[len(*sent)-1], outerror.Definition)
	if err != nil || values[1].Int32 != outerror.CodeRoomNotFound {
		t.Fatalf("unexpected missing room result values=%#v err=%v", values, err)
	}
}

// TestSendErrorMapsRoomSettingFailures verifies protocol error classification.
func TestSendErrorMapsRoomSettingFailures(t *testing.T) {
	connection, sent := connectionForBehaviorTest(t, netconn.NewRegistry())
	tests := []struct {
		// cause stores the domain failure.
		cause error
		// code stores the protocol result.
		code int32
	}{
		{roomservice.ErrRoomNotFound, outerror.CodeRoomNotFound}, {roomsettings.ErrAccessDenied, outerror.CodeNotOwner},
		{roomservice.ErrInvalidDoorMode, outerror.CodeInvalidDoorMode}, {roomservice.ErrInvalidMaxUsers, outerror.CodeInvalidUserLimit},
		{roomservice.ErrInvalidCategory, outerror.CodeInvalidCategory}, {roomservice.ErrPasswordRequired, outerror.CodeInvalidPassword},
		{roomservice.ErrInvalidDescription, outerror.CodeInvalidDescription}, {roomservice.ErrProhibitedName, outerror.CodeUnacceptableName},
		{roomservice.ErrProhibitedDescription, outerror.CodeUnacceptableDescription}, {roomservice.ErrProhibitedTag, outerror.CodeInvalidTag},
		{roomservice.ErrReservedTag, outerror.CodeReservedTag}, {roomservice.ErrInvalidTag, outerror.CodeInvalidTag},
		{errors.New("unknown settings failure"), outerror.CodeInvalidName},
	}
	for _, test := range tests {
		if err := (Handler{}).sendError(context.Background(), connection, 9, test.cause); err != nil {
			t.Fatalf("send error %v: %v", test.cause, err)
		}
		values, err := codec.DecodePacketExact((*sent)[len(*sent)-1], outerror.Definition)
		if err != nil || values[1].Int32 != test.code {
			t.Fatalf("cause=%v values=%#v err=%v", test.cause, values, err)
		}
	}
}

// behaviorFixture creates a complete active-room settings fixture.
func behaviorFixture(t *testing.T) (Handler, Command, *managerForTest, *[]codec.Packet, *publisherForTest) {
	t.Helper()
	connections := netconn.NewRegistry()
	connection, sent := connectionForBehaviorTest(t, connections)
	players := playerlive.NewRegistry()
	peer, _ := playerlive.NewSessionPeer(connection.ConnectionID, connection.ConnectionKind, time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "owner"}, peer)
	_ = player.EnterRoom(9)
	_ = players.Add(player)
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: connection.ConnectionID, ConnectionKind: connection.ConnectionKind})
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}, Version: sharedmodel.Version{Version: 1}}, OwnerPlayerID: 7, Name: "Room", MaxUsers: 25}
	updated := room
	updated.Name = "Updated"
	updated.Version.Version = 2
	manager := &managerForTest{room: room, updated: updated}
	runtime := roomlive.NewRegistry(nil)
	active, _ := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	_, _ = active.Join(roomlive.Occupant{PlayerID: 7, Username: "owner", ConnectionID: connection.ConnectionID, ConnectionKind: connection.ConnectionKind})
	events := &publisherForTest{}
	authorizer := roomsettings.New(permissionsForBehaviorTest{"allow": true}, roomsettings.Nodes{OwnManage: "allow", OwnPolicyManage: "allow"})
	handler := Handler{Players: players, Bindings: bindings, Rooms: manager, Authorize: authorizer, Runtime: runtime, Connections: connections, Events: events}
	input := Command{Handler: connection, RoomID: 9, Name: "Updated", MaxUsers: 25}

	return handler, input, manager, sent, events
}

// connectionForBehaviorTest creates a registered packet-capturing connection context.
func connectionForBehaviorTest(t *testing.T, connections *netconn.Registry) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var captured netconn.Context
	_ = inbound.Register(1, func(context netconn.Context, _ codec.Packet) error { captured = context; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 16)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "settings", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = connections.Register(session)
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})

	return captured, &sent
}
