package navigator

import (
	"context"
	"testing"
	"time"

	navviewer "github.com/niflaot/pixels/internal/realm/navigator/viewer/live"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/searchresult"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestRoomCountBroadcasterRefreshesAffectedViewer verifies visible room refreshes.
func TestRoomCountBroadcasterRefreshesAffectedViewer(t *testing.T) {
	players := playerlive.NewRegistry()
	connections := netconn.NewRegistry()
	runtime := roomlive.NewRegistry(nil)
	room := roomForCountTest(7)
	connection := &countConnection{id: "ws-1", kind: "websocket", done: make(chan struct{})}
	player := countPlayer(t, 1, connection.id, connection.kind)
	player.OpenNavigator().SetSearch(navviewer.LastSearch{Code: "hotel_view"}, []int64{7})
	mustCount(t, players.Add(player))
	mustCount(t, connections.Register(connection))
	mustActivateCountRoom(t, runtime, room.ID)
	mustJoinCountRoom(t, runtime, room.ID, player.ID(), connection)

	broadcaster := NewRoomCountBroadcaster(players, connections, nil, countRooms{rooms: []roommodel.Room{room}}, runtime)
	broadcaster.pending[room.ID] = struct{}{}
	broadcaster.Flush()

	if len(connection.sent) != 1 || connection.sent[0].Header != outsearch.Header {
		t.Fatalf("unexpected sent packets %#v", connection.sent)
	}
}

// TestRoomCountBroadcasterSkipsUnaffectedViewer verifies room filtering.
func TestRoomCountBroadcasterSkipsUnaffectedViewer(t *testing.T) {
	players := playerlive.NewRegistry()
	connections := netconn.NewRegistry()
	runtime := roomlive.NewRegistry(nil)
	connection := &countConnection{id: "ws-1", kind: "websocket", done: make(chan struct{})}
	player := countPlayer(t, 1, connection.id, connection.kind)
	player.OpenNavigator().SetSearch(navviewer.LastSearch{Code: "hotel_view"}, []int64{9})
	mustCount(t, players.Add(player))
	mustCount(t, connections.Register(connection))

	broadcaster := NewRoomCountBroadcaster(players, connections, nil, countRooms{rooms: []roommodel.Room{roomForCountTest(7)}}, runtime)
	broadcaster.pending[7] = struct{}{}
	broadcaster.Flush()

	if len(connection.sent) != 0 {
		t.Fatalf("unexpected sent packets %#v", connection.sent)
	}
}

// countRooms provides room search data for broadcaster tests.
type countRooms struct {
	// rooms stores navigator results.
	rooms []roommodel.Room
}

// Create creates a room for broadcaster tests.
func (rooms countRooms) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID finds a room for broadcaster tests.
func (rooms countRooms) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return roommodel.Room{}, false, nil
}

// ListByOwner lists owned rooms for broadcaster tests.
func (rooms countRooms) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return nil, nil
}

// ListPopular lists popular rooms for broadcaster tests.
func (rooms countRooms) ListPopular(context.Context, int) ([]roommodel.Room, error) {
	return rooms.rooms, nil
}

// ListHighestScore lists highest score rooms for broadcaster tests.
func (rooms countRooms) ListHighestScore(context.Context, int) ([]roommodel.Room, error) {
	return rooms.rooms, nil
}

// Search searches rooms for broadcaster tests.
func (rooms countRooms) Search(context.Context, string, int) ([]roommodel.Room, error) {
	return rooms.rooms, nil
}

// ListTags lists room tags for broadcaster tests.
func (rooms countRooms) ListTags(context.Context, int64) ([]roommodel.Tag, error) {
	return nil, nil
}

// SoftDelete deletes rooms for broadcaster tests.
func (rooms countRooms) SoftDelete(context.Context, int64) error {
	return nil
}

// ListCategories lists categories for broadcaster tests.
func (rooms countRooms) ListCategories(context.Context) ([]roommodel.Category, error) {
	return nil, nil
}

// countConnection records sent packets for broadcaster tests.
type countConnection struct {
	// id identifies the connection.
	id netconn.ID
	// kind identifies the connection transport.
	kind netconn.Kind
	// sent stores outbound packets.
	sent []codec.Packet
	// done closes when disconnected.
	done chan struct{}
}

// ID returns the connection id.
func (connection *countConnection) ID() netconn.ID {
	return connection.id
}

// Kind returns the connection kind.
func (connection *countConnection) Kind() netconn.Kind {
	return connection.kind
}

// StartedAt returns the connection start time.
func (connection *countConnection) StartedAt() time.Time {
	return time.Now()
}

// AuthenticatedAt returns the connection authentication time.
func (connection *countConnection) AuthenticatedAt() (time.Time, bool) {
	return time.Now(), true
}

// Authenticate marks the connection authenticated.
func (connection *countConnection) Authenticate(time.Time) error {
	return nil
}

// State returns the connection state.
func (connection *countConnection) State() netconn.State {
	return netconn.StateConnected
}

// Receive handles an inbound packet.
func (connection *countConnection) Receive(context.Context, codec.Packet) error {
	return nil
}

// Send records an outbound packet.
func (connection *countConnection) Send(_ context.Context, packet codec.Packet) error {
	connection.sent = append(connection.sent, packet)

	return nil
}

// Disconnect closes the connection.
func (connection *countConnection) Disconnect(context.Context, netconn.Reason) error {
	close(connection.done)

	return nil
}

// Done returns the connection disposal channel.
func (connection *countConnection) Done() <-chan struct{} {
	return connection.done
}

// countPlayer creates a live player for broadcaster tests.
func countPlayer(t *testing.T, id int64, connectionID netconn.ID, kind netconn.Kind) *playerlive.Player {
	t.Helper()

	peer, err := playerlive.NewSessionPeer(connectionID, kind, time.Now())
	mustCount(t, err)
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: id, Username: "demo"}, peer)
	mustCount(t, err)

	return player
}

// roomForCountTest creates a room model for broadcaster tests.
func roomForCountTest(id int64) roommodel.Room {
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}, Name: "Demo", OwnerName: "demo", MaxUsers: 25}
}

// mustActivateCountRoom activates a runtime room for broadcaster tests.
func mustActivateCountRoom(t *testing.T, runtime *roomlive.Registry, roomID int64) {
	t.Helper()

	_, err := runtime.Activate(roomlive.Snapshot{ID: roomID, MaxUsers: 25})
	mustCount(t, err)
}

// mustJoinCountRoom joins a runtime room for broadcaster tests.
func mustJoinCountRoom(t *testing.T, runtime *roomlive.Registry, roomID int64, playerID int64, connection *countConnection) {
	t.Helper()

	_, err := runtime.Join(context.Background(), roomID, roomlive.Occupant{PlayerID: playerID, ConnectionID: connection.ID(), ConnectionKind: connection.Kind()})
	mustCount(t, err)
}

// mustCount fails a broadcaster test when err is not nil.
func mustCount(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
