package enter

import (
	"context"
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outheightmap "github.com/niflaot/pixels/networking/outbound/room/heightmap"
)

// TestSendHeightMapSkipsWithoutActiveRoom verifies the nil-room guard.
func TestSendHeightMapSkipsWithoutActiveRoom(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)

	err := (Handler{}).sendHeightMap(context.Background(), connection, nil)
	if err != nil {
		t.Fatalf("send height map: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no packets, got %#v", *sent)
	}
}

// TestSendHeightMapSkipsWithoutLoadedWorld verifies the empty-tiles guard.
func TestSendHeightMapSkipsWithoutLoadedWorld(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 10})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	if err := (Handler{}).sendHeightMap(context.Background(), connection, room); err != nil {
		t.Fatalf("send height map: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no packets, got %#v", *sent)
	}
}

// TestSendHeightMapSendsPacketWithLoadedWorld verifies the populated-room path.
func TestSendHeightMapSendsPacketWithLoadedWorld(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 10})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err := room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Body: worldunit.RotationSouth,
		Head: worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	if err := (Handler{}).sendHeightMap(context.Background(), connection, room); err != nil {
		t.Fatalf("send height map: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outheightmap.Header {
		t.Fatalf("expected one height map packet, got %#v", *sent)
	}
}
