package broadcast

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionchanged "github.com/niflaot/pixels/internal/permission/events/changed"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outperks "github.com/niflaot/pixels/networking/outbound/session/perks"
	outpermissions "github.com/niflaot/pixels/networking/outbound/session/permissions"
	"github.com/niflaot/pixels/pkg/bus"
)

var (
	// testPerkNode stores a client-visible permission fixture.
	testPerkNode = permission.RegisterNode("permission.broadcast.test", "TEST_BROADCAST")
)

// fakePermissions supplies permission projection fixtures.
type fakePermissions struct {
	permissionservice.Manager
	// allowed stores node decisions.
	allowed map[permission.Node]bool
	// group stores the primary group fixture.
	group permissionmodel.Group
	// affected stores affected player fixtures.
	affected []int64
	// err stores an injected failure.
	err error
}

// HasPermission returns one fixture node decision.
func (permissions *fakePermissions) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return permissions.allowed[node], permissions.err
}

// PrimaryGroup returns the fixture primary group.
func (permissions *fakePermissions) PrimaryGroup(context.Context, int64) (permissionmodel.Group, bool, error) {
	return permissions.group, permissions.group.Weight != 0, permissions.err
}

// AffectedPlayerIDs returns fixture affected players.
func (permissions *fakePermissions) AffectedPlayerIDs(context.Context, int64) ([]int64, error) {
	return permissions.affected, permissions.err
}

// fakeConnection records projected packets.
type fakeConnection struct {
	netconn.Connection
	// packets stores sent protocol packets.
	packets []codec.Packet
}

// ID returns the fixture connection id.
func (connection *fakeConnection) ID() netconn.ID { return "connection-1" }

// Kind returns the fixture connection kind.
func (connection *fakeConnection) Kind() netconn.Kind { return "websocket" }

// Send records one projected packet.
func (connection *fakeConnection) Send(_ context.Context, packet codec.Packet) error {
	connection.packets = append(connection.packets, packet)
	return nil
}

// TestProjectorBuildsPermissionAndPerkPackets verifies Nitro permission state.
func TestProjectorBuildsPermissionAndPerkPackets(t *testing.T) {
	permissions := &fakePermissions{allowed: map[permission.Node]bool{testPerkNode: true}, group: permissionmodel.Group{Weight: 80}}
	players := playerlive.NewRegistry()
	expiresAt := time.Now().Add(time.Hour)
	peer, _ := playerlive.NewSessionPeer("club-connection", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 3, Username: "club", Club: playermodel.Club{Level: playermodel.ClubLevelVIP, ExpiresAt: &expiresAt}}, peer)
	_ = players.Add(player)
	packets, err := NewProjector(permissions, players).Packets(context.Background(), 3)
	if err != nil || len(packets) != 2 {
		t.Fatalf("unexpected packets=%#v err=%v", packets, err)
	}
	if packets[0].Header != outpermissions.Header || packets[1].Header != outperks.Header {
		t.Fatalf("unexpected headers %d and %d", packets[0].Header, packets[1].Header)
	}
	values, err := codec.DecodePacketExact(packets[0], outpermissions.Definition)
	if err != nil || values[0].Int32 != int32(playermodel.ClubLevelVIP) || values[1].Int32 != 80 || values[2].Boolean {
		t.Fatalf("unexpected permission values=%#v err=%v", values, err)
	}
	values, rest, err := codec.DecodePayload(nil, outperks.Definition, packets[1].Payload)
	if err != nil || values[0].Int32 < 1 || len(rest) == 0 {
		t.Fatalf("unexpected perk payload values=%#v rest=%d err=%v", values, len(rest), err)
	}
}

// TestProjectorPropagatesPermissionFailures verifies projection errors.
func TestProjectorPropagatesPermissionFailures(t *testing.T) {
	failure := errors.New("permission lookup failed")
	_, err := NewProjector(&fakePermissions{err: failure}, nil).Packets(context.Background(), 3)
	if !errors.Is(err, failure) {
		t.Fatalf("expected projection failure, got %v", err)
	}
}

// TestBroadcasterProjectsAffectedLivePlayersOnce verifies deduplicated delivery.
func TestBroadcasterProjectsAffectedLivePlayersOnce(t *testing.T) {
	permissions := &fakePermissions{allowed: map[permission.Node]bool{}, affected: []int64{3, 3}}
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("connection-1", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 3, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("register player: %v", err)
	}
	connections := netconn.NewRegistry()
	connection := &fakeConnection{}
	if err := connections.Register(connection); err != nil {
		t.Fatalf("register connection: %v", err)
	}
	groupID := int64(2)
	event := bus.Event{Name: permissionchanged.Name, Payload: permissionchanged.Payload{PlayerID: 3, GroupID: &groupID}}
	if err := New(NewProjector(permissions, players), permissions, players, connections).Handle(context.Background(), event); err != nil {
		t.Fatalf("broadcast permission state: %v", err)
	}
	if len(connection.packets) != 2 {
		t.Fatalf("expected one two-packet projection, got %d packets", len(connection.packets))
	}
}

// TestBroadcasterIgnoresForeignEvents verifies payload type filtering.
func TestBroadcasterIgnoresForeignEvents(t *testing.T) {
	players := playerlive.NewRegistry()
	broadcaster := New(NewProjector(&fakePermissions{}, players), &fakePermissions{}, players, netconn.NewRegistry())
	if err := broadcaster.Handle(context.Background(), bus.Event{Payload: "foreign"}); err != nil {
		t.Fatalf("expected foreign event to be ignored, got %v", err)
	}
}

// BenchmarkProjectorPackets measures live permission protocol projection.
func BenchmarkProjectorPackets(b *testing.B) {
	permissions := &fakePermissions{allowed: map[permission.Node]bool{testPerkNode: true}, group: permissionmodel.Group{Weight: 100}}
	projector := NewProjector(permissions, nil)
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		packets, err := projector.Packets(ctx, 3)
		if err != nil || len(packets) != 2 {
			b.Fatalf("unexpected packets=%#v err=%v", packets, err)
		}
	}
}

// BenchmarkBroadcasterHandle measures one live player permission refresh.
func BenchmarkBroadcasterHandle(b *testing.B) {
	permissions := &fakePermissions{allowed: map[permission.Node]bool{testPerkNode: true}}
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("connection-1", "websocket", time.Now())
	if err != nil {
		b.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 3, Username: "demo"}, peer)
	if err != nil {
		b.Fatalf("create player: %v", err)
	}
	if err := players.Add(player); err != nil {
		b.Fatalf("register player: %v", err)
	}
	connections := netconn.NewRegistry()
	connection := &fakeConnection{packets: make([]codec.Packet, 0, 2)}
	if err := connections.Register(connection); err != nil {
		b.Fatalf("register connection: %v", err)
	}
	broadcaster := New(NewProjector(permissions, players), permissions, players, connections)
	event := bus.Event{Name: permissionchanged.Name, Payload: permissionchanged.Payload{PlayerID: 3}}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		connection.packets = connection.packets[:0]
		if err := broadcaster.Handle(ctx, event); err != nil || len(connection.packets) != 2 {
			b.Fatalf("unexpected packets=%d err=%v", len(connection.packets), err)
		}
	}
}
