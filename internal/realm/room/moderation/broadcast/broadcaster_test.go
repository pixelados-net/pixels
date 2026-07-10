package broadcast

import (
	"context"
	"testing"

	bannedevent "github.com/niflaot/pixels/internal/realm/room/events/banned"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/events/kicked"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/events/muted"
	unbannedevent "github.com/niflaot/pixels/internal/realm/room/events/unbanned"
	unmutedevent "github.com/niflaot/pixels/internal/realm/room/events/unmuted"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdesktop "github.com/niflaot/pixels/networking/outbound/session/desktop"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestBroadcasterKickUsesStandardRoomLeave verifies runtime removal and occupancy cleanup.
func TestBroadcasterKickUsesStandardRoomLeave(t *testing.T) {
	runtime, active := occupiedRoomForTest(t)
	broadcaster := New(nil, nil, runtime, nil, nil)
	if err := broadcaster.Kick(context.Background(), 9, 2); err != nil {
		t.Fatalf("kick projection: %v", err)
	}
	if active.Occupancy().Count != 0 {
		t.Fatalf("expected empty room, got %#v", active.Occupancy())
	}
}

// TestBroadcasterBanUsesStandardRoomLeave verifies ban removal behavior.
func TestBroadcasterBanUsesStandardRoomLeave(t *testing.T) {
	runtime, active := occupiedRoomForTest(t)
	broadcaster := New(nil, nil, runtime, nil, nil)
	if err := broadcaster.Ban(context.Background(), 9, 2); err != nil {
		t.Fatalf("ban projection: %v", err)
	}
	if active.Occupancy().Count != 0 {
		t.Fatalf("expected empty room, got %#v", active.Occupancy())
	}
}

// TestBroadcasterForcedRemovalReturnsTargetToDesktop verifies complete kick and ban teardown.
func TestBroadcasterForcedRemovalReturnsTargetToDesktop(t *testing.T) {
	tests := []struct {
		// name stores the case name.
		name string
		// remove executes the forced removal.
		remove func(*Broadcaster) error
	}{
		{name: "kick", remove: func(broadcaster *Broadcaster) error { return broadcaster.Kick(context.Background(), 9, 2) }},
		{name: "ban", remove: func(broadcaster *Broadcaster) error { return broadcaster.Ban(context.Background(), 9, 2) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime, active := occupiedRoomForTest(t)
			connections, sent := moderationConnectionForTest(t)
			if err := test.remove(New(nil, nil, runtime, connections, nil)); err != nil {
				t.Fatalf("forced removal: %v", err)
			}
			if active.Occupancy().Count != 0 {
				t.Fatalf("expected empty room, got %#v", active.Occupancy())
			}
			if len(*sent) != 2 || (*sent)[1].Header != outdesktop.Header {
				t.Fatalf("expected notice followed by desktop view, got %#v", *sent)
			}
		})
	}
}

// TestDeferredKickRunsWithoutTransaction verifies standalone event projection.
func TestDeferredKickRunsWithoutTransaction(t *testing.T) {
	runtime, active := occupiedRoomForTest(t)
	handler := handleKick(New(nil, nil, runtime, nil, nil), zap.NewNop())
	err := handler(context.Background(), bus.Event{Name: kickedevent.Name, Payload: kickedevent.Payload{RoomID: 9, TargetPlayerID: 2, ActorID: 1}})
	if err != nil || active.Occupancy().Count != 0 {
		t.Fatalf("deferred kick occupancy=%d err=%v", active.Occupancy().Count, err)
	}
}

// TestBroadcasterSendsMuteAndUnbanPackets verifies target and room delivery.
func TestBroadcasterSendsMuteAndUnbanPackets(t *testing.T) {
	runtime, _ := occupiedRoomForTest(t)
	connections, sent := moderationConnectionForTest(t)
	broadcaster := New(nil, nil, runtime, connections, nil)
	if err := broadcaster.Mute(context.Background(), 9, 2, 300); err != nil {
		t.Fatalf("mute projection: %v", err)
	}
	if err := broadcaster.Unban(context.Background(), 9, 2); err != nil {
		t.Fatalf("unban projection: %v", err)
	}
	if len(*sent) != 2 {
		t.Fatalf("expected mute and unban packets, got %#v", *sent)
	}
}

// TestRegisterProjectsEveryModerationEvent verifies subscriber wiring and cleanup.
func TestRegisterProjectsEveryModerationEvent(t *testing.T) {
	runtime, _ := occupiedRoomForTest(t)
	local := bus.New()
	lifecycle := fxtest.NewLifecycle(t)
	broadcaster := New(nil, nil, runtime, nil, nil)
	if err := Register(lifecycle, local, broadcaster, zap.NewNop()); err != nil {
		t.Fatalf("register: %v", err)
	}
	events := []bus.Event{
		{Name: mutedevent.Name, Payload: mutedevent.Payload{RoomID: 9, TargetPlayerID: 2, DurationSeconds: 60}},
		{Name: unmutedevent.Name, Payload: unmutedevent.Payload{RoomID: 9, TargetPlayerID: 2}},
		{Name: unbannedevent.Name, Payload: unbannedevent.Payload{RoomID: 9, TargetPlayerID: 2}},
		{Name: kickedevent.Name, Payload: kickedevent.Payload{RoomID: 9, TargetPlayerID: 2}},
		{Name: bannedevent.Name, Payload: bannedevent.Payload{RoomID: 9, TargetPlayerID: 2}},
	}
	for _, event := range events {
		if err := local.Publish(context.Background(), event); err != nil {
			t.Fatalf("publish %s: %v", event.Name, err)
		}
	}
	lifecycle.RequireStop()
}

// occupiedRoomForTest creates one active occupied room.
func occupiedRoomForTest(t *testing.T) (*roomlive.Registry, *roomlive.Room) {
	t.Helper()
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	_, err = runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 2, Username: "Alice", ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")})
	if err != nil {
		t.Fatalf("join room: %v", err)
	}

	return runtime, active
}

// moderationConnectionForTest creates one packet-capturing connection.
func moderationConnectionForTest(t *testing.T) (*netconn.Registry, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 3)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "conn", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	registry := netconn.NewRegistry()
	if err := registry.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	return registry, &sent
}
