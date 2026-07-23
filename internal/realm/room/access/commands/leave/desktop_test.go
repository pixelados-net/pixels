package leave

import (
	"context"
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdesktop "github.com/niflaot/pixels/networking/outbound/session/desktop"
	outerror "github.com/niflaot/pixels/networking/outbound/session/error"
)

// TestToDesktopUsesStandardLeaveAndSendsHotelView verifies complete door-exit teardown.
func TestToDesktopUsesStandardLeaveAndSendsHotelView(t *testing.T) {
	player := playerForTest(t)
	players := playerRegistryForTest(t, player)
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "conn")
	runtime := roomlive.NewRegistry(nil)
	active := activeRoomForTest(t, runtime)

	err := (Handler{Players: players, Runtime: runtime, Connections: connections}).ToDesktop(context.Background(), 7)
	if err != nil {
		t.Fatalf("leave to desktop: %v", err)
	}
	if active.Occupancy().Count != 0 {
		t.Fatalf("expected empty room, got %#v", active.Occupancy())
	}
	if len(*sent) != 1 || (*sent)[0].Header != outdesktop.Header {
		t.Fatalf("expected desktop packet, got %#v", *sent)
	}
}

// TestToDesktopThenSendsNoticeAfterHotelView verifies post-exit notice ordering.
func TestToDesktopThenSendsNoticeAfterHotelView(t *testing.T) {
	player := playerForTest(t)
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "conn")
	runtime := roomlive.NewRegistry(nil)
	activeRoomForTest(t, runtime)
	notice, err := outerror.Encode(4008)
	if err != nil {
		t.Fatalf("encode notice: %v", err)
	}

	err = (Handler{Players: playerRegistryForTest(t, player), Runtime: runtime, Connections: connections}).ToDesktopThen(context.Background(), 7, notice)
	if err != nil {
		t.Fatalf("leave then notify: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outdesktop.Header || (*sent)[1].Header != outerror.Header {
		t.Fatalf("expected desktop then notice, got %#v", *sent)
	}
}
