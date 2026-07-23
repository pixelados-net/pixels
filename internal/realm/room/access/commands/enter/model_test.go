package enter

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentrytile "github.com/niflaot/pixels/networking/outbound/room/entrytile"
	outmodel "github.com/niflaot/pixels/networking/outbound/room/model"
	outmodelname "github.com/niflaot/pixels/networking/outbound/room/modelname"
	outrightslevel "github.com/niflaot/pixels/networking/outbound/room/rights/level"
	outrightsowner "github.com/niflaot/pixels/networking/outbound/room/rights/owner"
)

// controlPermissionsForTest resolves configured controller capabilities.
type controlPermissionsForTest struct {
	// allowed stores permitted nodes.
	allowed map[permission.Node]bool
	// err stores an optional resolution failure.
	err error
}

// HasPermission reports one configured permission decision.
func (permissions controlPermissionsForTest) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return permissions.allowed[node], permissions.err
}

// TestSendModelSendsEntryTileBeforeScaledHeightmap verifies room model bootstrapping order.
func TestSendModelSendsEntryTileBeforeScaledHeightmap(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)

	err := SendModel(context.Background(), connection, roomForTest(), layoutForTest())
	if err != nil {
		t.Fatalf("send model: %v", err)
	}
	if len(*sent) != 3 {
		t.Fatalf("expected model packets, got %#v", *sent)
	}
	if (*sent)[0].Header != outmodelname.Header || (*sent)[1].Header != outentrytile.Header || (*sent)[2].Header != outmodel.Header {
		t.Fatalf("unexpected model packet order %#v", *sent)
	}

	entryValues, err := codec.DecodePacketExact((*sent)[1], outentrytile.Definition)
	if err != nil {
		t.Fatalf("decode entry tile packet: %v", err)
	}
	if entryValues[0].Int32 != 0 || entryValues[1].Int32 != 0 || entryValues[2].Int32 != 2 {
		t.Fatalf("unexpected entry tile values %#v", entryValues)
	}

	values, err := codec.DecodePacketExact((*sent)[2], outmodel.Definition)
	if err != nil {
		t.Fatalf("decode model packet: %v", err)
	}
	if !values[0].Boolean {
		t.Fatalf("expected scaled heightmap packet, got %#v", values)
	}
}

// TestSendGeometryOmitsModelName verifies model requests cannot retrigger Nitro's request loop.
func TestSendGeometryOmitsModelName(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	if err := SendGeometry(context.Background(), connection, layoutForTest()); err != nil {
		t.Fatalf("send geometry: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outentrytile.Header || (*sent)[1].Header != outmodel.Header {
		t.Fatalf("expected entry tile and heightmap only, got %#v", *sent)
	}
}

// TestSendRightsProjectsOwnerAndStaffLevels verifies Nitro controller bootstrap values.
func TestSendRightsProjectsOwnerAndStaffLevels(t *testing.T) {
	tests := []struct {
		// name stores the case name.
		name string
		// playerID stores the entering player id.
		playerID int64
		// permissions stores global permission decisions.
		permissions map[permission.Node]bool
		// expectedLevel stores the Nitro controller level.
		expectedLevel int32
		// expectedOwner reports whether ROOM_RIGHTS_OWNER is expected.
		expectedOwner bool
	}{
		{name: "room owner", playerID: 7, expectedLevel: outrightslevel.Owner, expectedOwner: true},
		{name: "global staff", playerID: 8, permissions: map[permission.Node]bool{"rights.any": true, "kick.any": true}, expectedLevel: outrightslevel.Moderator, expectedOwner: true},
		{name: "ordinary visitor", playerID: 8, expectedLevel: outrightslevel.None},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connection, sent := sessionConnectionForTest(t)
			active := activeRoomForRightsTest(t)
			handler := Handler{Control: ControlPolicy{
				Permissions:    controlPermissionsForTest{allowed: test.permissions},
				RightsAnyGrant: "rights.any", ModerationAnyKick: "kick.any",
			}}

			if err := handler.sendRights(context.Background(), connection, roomForTest(), active, test.playerID); err != nil {
				t.Fatalf("send rights: %v", err)
			}
			expectedPackets := 1
			if test.expectedOwner {
				expectedPackets++
				if (*sent)[0].Header != outrightsowner.Header {
					t.Fatalf("expected owner packet, got %#v", *sent)
				}
			}
			if len(*sent) != expectedPackets {
				t.Fatalf("expected %d packets, got %#v", expectedPackets, *sent)
			}
			values, err := codec.DecodePacketExact((*sent)[len(*sent)-1], outrightslevel.Definition)
			if err != nil || values[0].Int32 != test.expectedLevel {
				t.Fatalf("unexpected controller level %#v err=%v", values, err)
			}
			if value, found := flatControlForTest(active, test.playerID); !found || value != strconv.Itoa(int(test.expectedLevel)) {
				t.Fatalf("unexpected avatar flat control value=%q found=%v", value, found)
			}
		})
	}
}

// TestSendRightsProjectsLocalRightsAndPermissionErrors verifies local and failure paths.
func TestSendRightsProjectsLocalRightsAndPermissionErrors(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	active := activeRoomForRightsTest(t)
	active.GrantRights(8)
	if err := (Handler{}).sendRights(context.Background(), connection, roomForTest(), active, 8); err != nil {
		t.Fatalf("send local rights: %v", err)
	}
	values, err := codec.DecodePacketExact((*sent)[0], outrightslevel.Definition)
	if err != nil || values[0].Int32 != outrightslevel.Rights {
		t.Fatalf("unexpected local rights %#v err=%v", values, err)
	}

	permissionErr := errors.New("permission lookup failed")
	handler := Handler{Control: ControlPolicy{Permissions: controlPermissionsForTest{err: permissionErr}, RightsAnyGrant: "rights.any"}}
	if err := handler.sendRights(context.Background(), connection, roomForTest(), active, 8); !errors.Is(err, permissionErr) {
		t.Fatalf("expected permission error, got %v", err)
	}
}

// activeRoomForRightsTest creates a runtime room for rights projection tests.
func activeRoomForRightsTest(t *testing.T) *roomlive.Room {
	t.Helper()
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatalf("create active room: %v", err)
	}
	roomGrid, err := layoutForTest().Grid()
	if err != nil {
		t.Fatalf("parse active room grid: %v", err)
	}
	if err := active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Rules: worldpath.DefaultRules()}); err != nil {
		t.Fatalf("load active room world: %v", err)
	}
	for _, playerID := range []int64{7, 8} {
		if _, err := active.Join(roomlive.Occupant{PlayerID: playerID, Username: "player", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
			t.Fatalf("join active room: %v", err)
		}
	}

	return active
}

// flatControlForTest returns one unit's room controller status.
func flatControlForTest(active *roomlive.Room, playerID int64) (string, bool) {
	for _, roomUnit := range active.Units() {
		if roomUnit.PlayerID != playerID {
			continue
		}
		for _, status := range roomUnit.Statuses {
			if status.Key == worldunit.StatusFlatControl {
				return status.Value, true
			}
		}
	}

	return "", false
}

// TestSendModelReturnsTransportError verifies model packet send failures.
func TestSendModelReturnsTransportError(t *testing.T) {
	connection, sendErr := failingModelConnectionForTest(t)

	err := SendModel(context.Background(), connection, roomForTest(), layoutForTest())
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected send error, got %v", err)
	}
}

// failingModelConnectionForTest creates a connection that fails the second send.
func failingModelConnectionForTest(t *testing.T) (netconn.Context, error) {
	t.Helper()

	sendErr := errors.New("send failed")
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

	sendCount := 0
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       netconn.ID("conn"),
		Kind:     netconn.Kind("websocket"),
		Inbound:  inbound,
		Outbound: outbound,
		Sender: func(context.Context, codec.Packet) error {
			sendCount++
			if sendCount == 2 {
				return sendErr
			}
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

	return captured, sendErr
}
