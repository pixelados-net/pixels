package broadcast

import (
	"context"
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outeffect "github.com/niflaot/pixels/networking/outbound/room/entities/effect"
)

// TestSendRoomActionsSynchronizesExistingEffect verifies late entrants see prior team effects.
func TestSendRoomActionsSynchronizesExistingEffect(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "viewer")
	connection, found := connections.Get("websocket", "viewer")
	if !found {
		t.Fatal("viewer connection missing")
	}
	room := loadedRoomForSpawnTest(t)
	occupant := roomlive.Occupant{PlayerID: 7, Username: "existing", ConnectionID: "existing", ConnectionKind: "websocket"}
	if _, err := room.Join(occupant); err != nil {
		t.Fatalf("join existing player: %v", err)
	}
	if _, found = room.SetUnitEffect(7, 35); !found {
		t.Fatal("set team effect")
	}
	if err := SendRoomActions(context.Background(), connection, room); err != nil {
		t.Fatalf("send room actions: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outeffect.Header {
		t.Fatalf("packets=%+v", *sent)
	}
	values, err := codec.DecodePacketExact((*sent)[0], outeffect.Definition)
	if err != nil || values[1].Int32 != 35 {
		t.Fatalf("effect values=%+v err=%v", values, err)
	}
}
