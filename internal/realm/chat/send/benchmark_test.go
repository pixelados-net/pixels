package send

import (
	"context"
	"strings"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	outflood "github.com/niflaot/pixels/networking/outbound/chat/flood"
	outshout "github.com/niflaot/pixels/networking/outbound/chat/shout"
	outtalk "github.com/niflaot/pixels/networking/outbound/chat/talk"
	outtyping "github.com/niflaot/pixels/networking/outbound/chat/typing"
	outwhisper "github.com/niflaot/pixels/networking/outbound/chat/whisper"
	outmute "github.com/niflaot/pixels/networking/outbound/room/moderation/muted"
	"github.com/niflaot/pixels/pkg/i18n"
)

// TestSanitizeAndGesture verifies protocol-safe text and gesture inference.
func TestSanitizeAndGesture(t *testing.T) {
	if value := sanitize("  hello\nworld  "); value != "hello world" {
		t.Fatalf("unexpected sanitized message %q", value)
	}
	if gesture(":) hello") != 1 || gesture("plain") != 0 {
		t.Fatal("unexpected gesture inference")
	}
}

// TestNativeRejectionFeedback verifies mute, mute-all, flood, and length rejection channels.
func TestNativeRejectionFeedback(t *testing.T) {
	fixture := newFixture(t)
	fixture.active.SetMuted(1, time.Now().Add(time.Minute))
	assertSourceHeader(t, fixture, outmute.Header)
	fixture.active.SetMuted(1, time.Time{})
	fixture.active.SetMuteAll(true)
	assertSourceHeader(t, fixture, outmute.Header)
	fixture.active.SetMuteAll(false)
	fixture.counter.count = 200
	assertSourceHeader(t, fixture, outflood.Header)
	fixture.counter.count = 1
	fixture.request.Message = strings.Repeat("x", 300)
	assertSourceHeader(t, fixture, 3801)
}

// TestShoutWhisperAndTypingDelivery verifies the remaining Nitro chat modes.
func TestShoutWhisperAndTypingDelivery(t *testing.T) {
	fixture := newFixture(t)
	fixture.request.Kind = KindShout
	fixture.request.Message = "hello"
	assertSourceHeader(t, fixture, outshout.Header)
	if len(*fixture.targetPackets) != 1 || (*fixture.targetPackets)[0].Header != outshout.Header {
		t.Fatal("expected target shout")
	}
	*fixture.sourcePackets = nil
	*fixture.targetPackets = nil
	fixture.request.Kind = KindWhisper
	fixture.request.Recipient = "bob"
	assertSourceHeader(t, fixture, outwhisper.Header)
	if len(*fixture.targetPackets) != 1 || (*fixture.targetPackets)[0].Header != outwhisper.Header {
		t.Fatal("expected target whisper")
	}
	if err := fixture.service.Typing(context.Background(), *fixture.context, true); err != nil {
		t.Fatalf("typing: %v", err)
	}
	if (*fixture.targetPackets)[1].Header != outtyping.Header {
		t.Fatal("expected target typing update")
	}
}

// TestWhisperObserverSeesRecipient verifies privileged whisper decoration.
func TestWhisperObserverSeesRecipient(t *testing.T) {
	fixture := newFixture(t)
	observerPackets := make([]codec.Packet, 0, 1)
	registerSession(t, fixture.connections, "observer", &observerPackets, nil)
	addPlayer(t, fixture.players, fixture.bindings, fixture.runtime, fixture.active, 3, "observer", "carol")
	fixture.permissions.allowed["observe"] = true
	fixture.service.nodes.WhisperObserveAny = "observe"
	fixture.service.translations = i18n.NewCatalog(i18n.Config{DefaultLocale: "es"}, map[i18n.Locale]map[i18n.Key]string{
		"es": {"chat.whisper.observer": "Para {recipient}: {message}"},
	})
	fixture.request.Kind = KindWhisper
	fixture.request.Recipient = "bob"
	fixture.request.Message = "secret"
	assertSourceHeader(t, fixture, outwhisper.Header)
	if len(observerPackets) != 1 {
		t.Fatalf("expected one observer packet, got %#v", observerPackets)
	}
	values, err := codec.DecodePacketExact(observerPackets[0], outwhisper.Definition)
	if err != nil || values[1].String != "Para bob: secret" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}

// TestBypassesAndMissingWhisperFeedback verifies capability fast paths and expected feedback.
func TestBypassesAndMissingWhisperFeedback(t *testing.T) {
	fixture := newFixture(t)
	fixture.counter.count = 999
	fixture.permissions.allowed["flood"] = true
	fixture.permissions.allowed["length"] = true
	fixture.permissions.allowed["filter"] = true
	fixture.request.Message = strings.Repeat("clean", 80)
	assertSourceHeader(t, fixture, outtalk.Header)
	fixture.active.SetMuteAll(true)
	fixture.active.GrantRights(1)
	assertSourceHeader(t, fixture, outtalk.Header)
	fixture.active.SetMuteAll(false)
	fixture.request.Kind = KindWhisper
	fixture.request.Recipient = "missing"
	fixture.request.Message = "hello"
	assertSourceHeader(t, fixture, 3801)
	if err := fixture.service.Typing(context.Background(), *fixture.context, false); err != nil {
		t.Fatalf("stop typing: %v", err)
	}
}

// assertSourceHeader executes one source packet and checks the latest response header.
func assertSourceHeader(t *testing.T, fixture fixture, header uint16) {
	t.Helper()
	*fixture.sourcePackets = nil
	if err := fixture.source.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive: %v", err)
	}
	if len(*fixture.sourcePackets) != 1 || (*fixture.sourcePackets)[0].Header != header {
		t.Fatalf("expected header %d, packets=%#v", header, *fixture.sourcePackets)
	}
}

// TestWithinDistance verifies squared-radius boundaries.
func TestWithinDistance(t *testing.T) {
	source := unitAt(2, 2)
	if !withinDistance(source, unitAt(5, 6), 25) {
		t.Fatal("expected boundary point in range")
	}
	if withinDistance(source, unitAt(6, 6), 25) {
		t.Fatal("expected distant point out of range")
	}
}

// BenchmarkDistanceFilter measures a realistic full-room distance scan.
func BenchmarkDistanceFilter(b *testing.B) {
	source := unitAt(10, 10)
	units := make([]roomlive.UnitSnapshot, 100)
	for index := range units {
		units[index] = unitAt(index%20, index/20)
	}
	b.ReportAllocs()
	for range b.N {
		count := 0
		for _, unit := range units {
			if withinDistance(source, unit, 100) {
				count++
			}
		}
		if count == 0 {
			b.Fatal("expected recipients")
		}
	}
}

// unitAt creates one room unit fixture.
func unitAt(x int, y int) roomlive.UnitSnapshot {
	point, _ := grid.NewPoint(x, y)
	return roomlive.UnitSnapshot{Position: worldpath.Position{Point: point}}
}
