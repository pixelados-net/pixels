package event

import (
	"testing"

	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// TestConcreteEventsExposeTypedState verifies cancellable and notification contracts.
func TestConcreteEventsExposeTypedState(t *testing.T) {
	player := sdkplayer.Player{ID: 7, Username: "alice", Online: true}
	chat := NewChatSend(player, 3, "hello")
	if chat.Name() != ChatSendName || chat.Cancelled() || chat.Text != "hello" {
		t.Fatalf("unexpected chat event: %+v", chat)
	}
	chat.SetCancelled(true)
	if !chat.Cancelled() {
		t.Fatal("expected cancellable event state")
	}
	connected := &PlayerConnected{Player: player}
	if connected.Name() != PlayerConnectedName {
		t.Fatalf("unexpected connected name %q", connected.Name())
	}
}
