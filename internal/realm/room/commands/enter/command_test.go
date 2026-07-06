package enter

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
)

// TestCommandName verifies the stable command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// TestLoadRoomPropagatesStoreErrors verifies persistence errors.
func TestLoadRoomPropagatesStoreErrors(t *testing.T) {
	storeErr := errors.New("store failed")
	handler := Handler{Rooms: roomManagerForTest{err: storeErr}}
	_, _, err := handler.loadRoom(context.Background(), 9)
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected store error, got %v", err)
	}

	handler = Handler{
		Rooms:   roomManagerForTest{room: roomForTest(), found: true},
		Layouts: layoutManagerForTest{err: storeErr},
	}
	_, _, err = handler.loadRoom(context.Background(), 9)
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected layout error, got %v", err)
	}
}

// TestJoinAllowsMissingEventBus verifies publish is optional.
func TestJoinAllowsMissingEventBus(t *testing.T) {
	handler := Handler{Runtime: roomlive.NewRegistry(nil)}
	err := handler.join(context.Background(), playerForTest(t), connectionForTest(), roomForTest(), layoutForTest())
	if err != nil {
		t.Fatalf("join without events: %v", err)
	}
}

// TestCommandEnvelopeValid verifies command envelope naming.
func TestCommandEnvelopeValid(t *testing.T) {
	envelope := command.Envelope[Command]{Command: Command{RoomID: 9}}
	if !envelope.Valid() {
		t.Fatal("expected valid command envelope")
	}
}
