package command

import (
	"context"
	"testing"

	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// accessForTest stores player sender outcomes.
type accessForTest struct {
	// allowed stores the permission decision.
	allowed bool
	// message stores the latest reply.
	message string
}

// Message records one reply.
func (access *accessForTest) Message(_ int64, message string) error {
	access.message = message
	return nil
}

// HasPermission returns the fixture permission decision.
func (access *accessForTest) HasPermission(int64, string) (bool, error) {
	return access.allowed, nil
}

// TestPlayerAndConsoleSenders verifies identity, permissions, and reply channels.
func TestPlayerAndConsoleSenders(t *testing.T) {
	access := &accessForTest{allowed: true}
	player := NewPlayerSender(sdkplayer.Player{ID: 7, Username: "alice"}, access)
	if player.Name() != "alice" || player.Kind() != SenderKindPlayer || !player.HasPermission("plugin.test.use") {
		t.Fatal("expected allowed player sender")
	}
	if err := player.Reply(context.Background(), "hello"); err != nil || access.message != "hello" {
		t.Fatalf("expected player reply, message=%q err=%v", access.message, err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := player.Reply(cancelled, "late"); err == nil || access.message != "hello" {
		t.Fatalf("expected cancelled reply suppression, message=%q err=%v", access.message, err)
	}
	replied := false
	console := ConsoleSender{ReplyFunc: func(context.Context, string) error { replied = true; return nil }}
	if console.Name() != "console" || console.Kind() != SenderKindConsole || !console.HasPermission("anything") {
		t.Fatal("expected trusted console sender")
	}
	if err := console.Reply(context.Background(), "ok"); err != nil || !replied {
		t.Fatalf("expected console reply, replied=%v err=%v", replied, err)
	}
	if err := (ConsoleSender{}).Reply(context.Background(), "ignored"); err != nil {
		t.Fatalf("expected optional console reply, got %v", err)
	}
}
