package broadcast

import (
	"context"
	"errors"
	"testing"

	bannedevent "github.com/niflaot/pixels/internal/realm/room/events/banned"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/events/kicked"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/events/muted"
	unbannedevent "github.com/niflaot/pixels/internal/realm/room/events/unbanned"
	unmutedevent "github.com/niflaot/pixels/internal/realm/room/events/unmuted"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outdesktop "github.com/niflaot/pixels/networking/outbound/session/desktop"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestKickedNoticeUsesGlobalLocalizedAlert verifies post-desktop compatibility and localization.
func TestKickedNoticeUsesGlobalLocalizedAlert(t *testing.T) {
	translations := i18n.NewCatalog(i18n.Config{DefaultLocale: "es"}, map[i18n.Locale]map[i18n.Key]string{
		"es": {"room.moderation.kicked": "Fuiste expulsado."},
	})
	packet, err := KickedNotice(translations)
	if err != nil {
		t.Fatalf("encode kicked notice: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, outalert.Definition)
	if err != nil {
		t.Fatalf("decode kicked notice: %v", err)
	}
	if packet.Header != outalert.Header || len(values) != 1 || values[0].String != "Fuiste expulsado." {
		t.Fatalf("unexpected kicked notice packet=%#v values=%#v", packet, values)
	}
}

// TestBroadcasterKickUsesStandardRoomLeave verifies runtime removal and occupancy cleanup.
func TestBroadcasterKickUsesStandardRoomLeave(t *testing.T) {
	runtime, active := occupiedRoomForTest(t)
	broadcaster := New(nil, nil, runtime, nil, nil, nil)
	if err := broadcaster.Kick(context.Background(), 9, 2); err != nil {
		t.Fatalf("kick projection: %v", err)
	}
	if active.Occupancy().Count != 0 {
		t.Fatalf("expected empty room, got %#v", active.Occupancy())
	}
}

// TestBroadcasterKickWalksReachableTargetToDoor verifies soft kicks defer removal while A-star has a path.
func TestBroadcasterKickWalksReachableTargetToDoor(t *testing.T) {
	runtime, active := occupiedRoomForTest(t)
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse room grid: %v", err)
	}
	if err := active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatalf("load room world: %v", err)
	}
	if _, err := active.MoveTo(2, grid.MustPoint(2, 0)); err != nil {
		t.Fatalf("move target away from door: %v", err)
	}
	active.Tick()
	active.Tick()
	active.Tick()

	connections, sent := moderationConnectionForTest(t)
	if err := New(nil, nil, runtime, connections, nil, nil).Kick(context.Background(), 9, 2); err != nil {
		t.Fatalf("kick projection: %v", err)
	}
	if active.Occupancy().Count != 1 {
		t.Fatalf("expected walking occupant, got %#v", active.Occupancy())
	}
	if _, err := active.MoveTo(2, grid.MustPoint(1, 0)); !errors.Is(err, roomlive.ErrUnitExiting) {
		t.Fatalf("expected forced exit path, got %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected kicked notice after walking, got %#v", *sent)
	}
}

// TestBroadcasterBanUsesStandardRoomLeave verifies ban removal behavior.
func TestBroadcasterBanUsesStandardRoomLeave(t *testing.T) {
	runtime, active := occupiedRoomForTest(t)
	broadcaster := New(nil, nil, runtime, nil, nil, nil)
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
			if err := test.remove(New(nil, nil, runtime, connections, nil, nil)); err != nil {
				t.Fatalf("forced removal: %v", err)
			}
			if active.Occupancy().Count != 0 {
				t.Fatalf("expected empty room, got %#v", active.Occupancy())
			}
			if len(*sent) != 2 {
				t.Fatalf("expected two forced-removal packets, got %#v", *sent)
			}
			if test.name == "kick" && ((*sent)[0].Header != outdesktop.Header || (*sent)[1].Header != outalert.Header) {
				t.Fatalf("expected desktop followed by kicked notice, got %#v", *sent)
			}
			if test.name == "ban" && (*sent)[1].Header != outdesktop.Header {
				t.Fatalf("expected ban notice followed by desktop, got %#v", *sent)
			}
		})
	}
}

// TestDeferredKickRunsWithoutTransaction verifies standalone event projection.
func TestDeferredKickRunsWithoutTransaction(t *testing.T) {
	runtime, active := occupiedRoomForTest(t)
	handler := handleKick(New(nil, nil, runtime, nil, nil, nil), zap.NewNop())
	err := handler(context.Background(), bus.Event{Name: kickedevent.Name, Payload: kickedevent.Payload{RoomID: 9, TargetPlayerID: 2, ActorID: 1}})
	if err != nil || active.Occupancy().Count != 0 {
		t.Fatalf("deferred kick occupancy=%d err=%v", active.Occupancy().Count, err)
	}
}

// TestBroadcasterSendsMuteAndUnbanPackets verifies target and room delivery.
func TestBroadcasterSendsMuteAndUnbanPackets(t *testing.T) {
	runtime, _ := occupiedRoomForTest(t)
	connections, sent := moderationConnectionForTest(t)
	broadcaster := New(nil, nil, runtime, connections, nil, nil)
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
	broadcaster := New(nil, nil, runtime, nil, nil, nil)
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
