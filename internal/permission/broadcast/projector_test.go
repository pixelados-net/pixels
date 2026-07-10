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
	packets, err := NewProjector(permissions).Packets(context.Background(), 3)
	if err != nil || len(packets) != 2 {
		t.Fatalf("unexpected packets=%#v err=%v", packets, err)
	}
	if packets[0].Header != outpermissions.Header || packets[1].Header != outperks.Header {
		t.Fatalf("unexpected headers %d and %d", packets[0].Header, packets[1].Header)
	}
	values, err := codec.DecodePacketExact(packets[0], outpermissions.Definition)
	if err != nil || values[1].Int32 != 80 || values[2].Boolean {
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
	_, err := NewProjector(&fakePermissions{err: failure}).Packets(context.Background(), 3)
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
	if err := New(NewProjector(permissions), permissions, players, connections).Handle(context.Background(), event); err != nil {
		t.Fatalf("broadcast permission state: %v", err)
	}
	if len(connection.packets) != 2 {
		t.Fatalf("expected one two-packet projection, got %d packets", len(connection.packets))
	}
}

// TestBroadcasterIgnoresForeignEvents verifies payload type filtering.
func TestBroadcasterIgnoresForeignEvents(t *testing.T) {
	broadcaster := New(NewProjector(&fakePermissions{}), &fakePermissions{}, playerlive.NewRegistry(), netconn.NewRegistry())
	if err := broadcaster.Handle(context.Background(), bus.Event{Payload: "foreign"}); err != nil {
		t.Fatalf("expected foreign event to be ignored, got %v", err)
	}
}
