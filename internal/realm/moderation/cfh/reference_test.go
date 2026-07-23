package cfh

import (
	"testing"
	"time"

	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	connection "github.com/niflaot/pixels/networking/connection"
)

// TestPositiveReferenceRejectsProtocolSentinels verifies optional player identifiers.
func TestPositiveReferenceRejectsProtocolSentinels(t *testing.T) {
	if positiveReference(0) != nil || positiveReference(-1) != nil {
		t.Fatal("non-positive references must be absent")
	}
	if value := positiveReference(3); value == nil || *value != 3 {
		t.Fatalf("value=%v", value)
	}
}

// TestReportErrorUsesNitroResultCodes verifies throttling and generic failures remain distinct.
func TestReportErrorUsesNitroResultCodes(t *testing.T) {
	handler := Handler{Context: &moderationruntime.Context{}}
	if code, message := handler.reportError(moderationcore.ErrThrottled); code != 1 || message != "moderation.report.throttled" {
		t.Fatalf("throttled code=%d message=%q", code, message)
	}
	if code, message := handler.reportError(moderationcore.ErrDisabled); code != 3 || message != "moderation.report.disabled" {
		t.Fatalf("disabled code=%d message=%q", code, message)
	}
}

// TestReportRoomReferenceFallsBackToLivePresence verifies Nitro's zero room identifier.
func TestReportRoomReferenceFallsBackToLivePresence(t *testing.T) {
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer(connection.ID("reporter"), "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 4, Username: "Carol"}, peer)
	if err != nil {
		t.Fatal(err)
	}
	if err = player.EnterRoom(7); err != nil {
		t.Fatal(err)
	}
	if err = players.Add(player); err != nil {
		t.Fatal(err)
	}
	if value := reportRoomReference(players, 4, 0); value == nil || *value != 7 {
		t.Fatalf("fallback=%v", value)
	}
	if value := reportRoomReference(players, 4, 9); value == nil || *value != 9 {
		t.Fatalf("wire=%v", value)
	}
}

// TestReportRoomReferenceAllowsNoRoom verifies reports outside room presence.
func TestReportRoomReferenceAllowsNoRoom(t *testing.T) {
	if reportRoomReference(nil, 4, 0) != nil {
		t.Fatal("missing runtime room must remain absent")
	}
}
