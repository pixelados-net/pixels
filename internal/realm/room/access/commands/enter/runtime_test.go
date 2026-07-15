package enter

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/permission"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomleft "github.com/niflaot/pixels/internal/realm/room/access/events/left"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// doorbellPermissions resolves one global responder node.
type doorbellPermissions struct {
	// playerID identifies the allowed player.
	playerID int64
	// node identifies the allowed node.
	node permission.Node
}

// HasPermission reports whether a player holds the responder fixture node.
func (permissions doorbellPermissions) HasPermission(_ context.Context, playerID int64, node permission.Node) (bool, error) {
	return playerID == permissions.playerID && node == permissions.node, nil
}

// TestRoomSnapshotMapsRuntimeFields verifies persistent room to runtime mapping.
func TestRoomSnapshotMapsRuntimeFields(t *testing.T) {
	categoryID := int64(3)
	snapshot := roomSnapshot(roommodel.Room{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		OwnerPlayerID: 7,
		CategoryID:    &categoryID,
		MaxUsers:      25,
	})

	if snapshot.ID != 9 || snapshot.OwnerPlayerID != 7 || snapshot.CategoryID == nil || *snapshot.CategoryID != 3 || snapshot.MaxUsers != 25 {
		t.Fatalf("unexpected snapshot %#v", snapshot)
	}
}

// TestHandleJoinsRoomAndSendsEntryPackets verifies full command orchestration.
func TestHandleJoinsRoomAndSendsEntryPackets(t *testing.T) {
	player := playerForTest(t)
	connection, sent := sessionConnectionForTest(t)
	players := playerRegistryForTest(t, player)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("bind player: %v", err)
	}
	publisher := &publisherForTest{}
	publishedAfterPackets := -1
	publisher.onPublish = func(bus.Event) {
		publishedAfterPackets = len(*sent)
	}
	handler := Handler{
		Players:  players,
		Bindings: bindings,
		Rooms:    roomManagerForTest{room: roomForTest(), found: true},
		Layouts:  layoutManagerForTest{roomLayout: layoutForTest(), found: true},
		Runtime:  roomlive.NewRegistry(nil),
		Events:   publisher,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, RoomID: 9},
	})
	if err != nil {
		t.Fatalf("handle command: %v", err)
	}
	if roomID, found := player.CurrentRoom(); !found || roomID != 9 {
		t.Fatalf("expected current room 9, got %d found=%v", roomID, found)
	}
	if active, found := handler.Runtime.Find(9); !found || !active.WorldLoaded() {
		t.Fatalf("expected active loaded room")
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != roomentered.Name {
		t.Fatalf("unexpected events %#v", publisher.events)
	}
	if len(*sent) != 10 {
		t.Fatalf("expected entered, model, and height map packets, got %#v", *sent)
	}
	if publishedAfterPackets != len(*sent) {
		t.Fatalf("expected room entry publication after bootstrap packets, published_at=%d packets=%d", publishedAfterPackets, len(*sent))
	}
}

// TestHandleReturnsLoadError verifies command load failures.
func TestHandleReturnsLoadError(t *testing.T) {
	player := playerForTest(t)
	connection, _ := sessionConnectionForTest(t)
	handler := Handler{
		Players:  playerRegistryForTest(t, player),
		Bindings: bindingRegistryForTest(t, 7),
		Rooms:    roomManagerForTest{},
		Layouts:  layoutManagerForTest{roomLayout: layoutForTest(), found: true},
		Runtime:  roomlive.NewRegistry(nil),
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, RoomID: 9},
	})
	if !errors.Is(err, roomservice.ErrRoomNotFound) {
		t.Fatalf("expected room not found, got %v", err)
	}
}

// TestHandleSendsRoomFullEntryError verifies room-full packet mapping.
func TestHandleSendsRoomFullEntryError(t *testing.T) {
	player := playerForTest(t)
	connection, sent := sessionConnectionForTest(t)
	runtime := roomlive.NewRegistry(nil)
	if _, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 1}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, occupantForTest(8)); err != nil {
		t.Fatalf("fill room: %v", err)
	}
	fullRoom := roomForTest()
	fullRoom.MaxUsers = 1
	handler := Handler{
		Players:  playerRegistryForTest(t, player),
		Bindings: bindingRegistryForTest(t, 7),
		Rooms:    roomManagerForTest{room: fullRoom, found: true},
		Layouts:  layoutManagerForTest{roomLayout: layoutForTest(), found: true},
		Runtime:  runtime,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, RoomID: 9},
	})
	if err != nil {
		t.Fatalf("handle room full: %v", err)
	}
	if len(*sent) != 1 {
		t.Fatalf("expected entry error packet, got %#v", *sent)
	}
}

// TestJoinLeavesPreviousRoomAndLoadsWorld verifies runtime join orchestration.
func TestJoinLeavesPreviousRoomAndLoadsWorld(t *testing.T) {
	player := playerForTest(t)
	runtime := roomlive.NewRegistry(nil)
	publisher := &publisherForTest{}
	handler := Handler{Runtime: runtime, Events: publisher}
	if _, err := runtime.Activate(roomlive.Snapshot{ID: 3, MaxUsers: 5}); err != nil {
		t.Fatalf("activate previous room: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 3, occupantForTest(7)); err != nil {
		t.Fatalf("join previous room: %v", err)
	}
	if err := player.EnterRoom(3); err != nil {
		t.Fatalf("enter previous room: %v", err)
	}

	_, err := handler.join(context.Background(), player, connectionForTest(), roomForTest(), layoutForTest())
	if err != nil {
		t.Fatalf("join target room: %v", err)
	}

	previous, found := runtime.Find(3)
	if !found || previous.Occupancy().Count != 0 {
		t.Fatalf("expected previous room empty")
	}
	active, found := runtime.Find(9)
	if !found || !active.WorldLoaded() || len(active.Units()) != 1 {
		t.Fatalf("expected loaded target room")
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != roomleft.Name {
		t.Fatalf("unexpected events %#v", publisher.events)
	}
}

// TestLoadRoomFindsRoomAndLayout verifies persistent room lookup.
func TestLoadRoomFindsRoomAndLayout(t *testing.T) {
	handler := Handler{
		Rooms:   roomManagerForTest{room: roomForTest(), found: true},
		Layouts: layoutManagerForTest{roomLayout: layoutForTest(), found: true},
	}

	room, roomLayout, err := handler.loadRoom(context.Background(), 9)
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	if room.ID != 9 || roomLayout.Name != "model_a" {
		t.Fatalf("unexpected room=%#v layout=%#v", room, roomLayout)
	}
}

// TestLoadRoomRejectsMissingRecords verifies room and layout misses.
func TestLoadRoomRejectsMissingRecords(t *testing.T) {
	handler := Handler{Rooms: roomManagerForTest{}}
	_, _, err := handler.loadRoom(context.Background(), 9)
	if !errors.Is(err, roomservice.ErrRoomNotFound) {
		t.Fatalf("expected missing room, got %v", err)
	}

	handler = Handler{
		Rooms:   roomManagerForTest{room: roomForTest(), found: true},
		Layouts: layoutManagerForTest{},
	}
	_, _, err = handler.loadRoom(context.Background(), 9)
	if !errors.Is(err, roomservice.ErrLayoutNotAvailable) {
		t.Fatalf("expected missing layout, got %v", err)
	}
}

// TestDoorbellApproversIncludesGlobalPermission verifies non-owner responder selection.
func TestDoorbellApproversIncludesGlobalPermission(t *testing.T) {
	const node permission.Node = "room.doorbell.answer.any"
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatalf("create active room: %v", err)
	}
	if _, err := active.Join(occupantForTest(2)); err != nil {
		t.Fatalf("join moderator: %v", err)
	}
	entryService := roomentry.New(roomentry.Config{}, nil, doorbellPermissions{playerID: 2, node: node}, nil, roomentry.Nodes{AnswerAnyDoorbell: node})
	approvers, err := (Handler{Entry: entryService}).doorbellApprovers(context.Background(), active, roomForTest())
	if err != nil || len(approvers) != 1 || approvers[0].PlayerID != 2 {
		t.Fatalf("unexpected approvers=%#v err=%v", approvers, err)
	}
}

// TestSendEntryErrorMapsRoomFull verifies protocol entry error handling.
func TestSendEntryErrorMapsRoomFull(t *testing.T) {
	handler := Handler{}
	err := handler.sendEntryError(context.Background(), netconn.Context{}, errors.New("other"))
	if err == nil || err.Error() != "other" {
		t.Fatalf("expected original error, got %v", err)
	}
	err = handler.sendEntryError(context.Background(), netconn.Context{}, roomlive.ErrRoomFull)
	if !errors.Is(err, netconn.ErrInvalidConnection) {
		t.Fatalf("expected invalid connection, got %v", err)
	}
}
