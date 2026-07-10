package pickup

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestBubbleErrorKeyMapsKnownErrors verifies every mapped soft error resolves to its bubble key.
func TestBubbleErrorKeyMapsKnownErrors(t *testing.T) {
	cases := []struct {
		err      error
		wantKey  string
		wantSoft bool
	}{
		{furnitureservice.ErrNotItemOwner, "session.bubble.furniture.no_rights", true},
		{furnitureservice.ErrItemNotFound, "session.bubble.furniture.item_not_found", true},
		{furnitureservice.ErrItemNotPlaced, "session.bubble.furniture.item_not_found", true},
		{furnitureservice.ErrInvalidItemID, "session.bubble.furniture.invalid_move", true},
		{furnitureservice.ErrInvalidPlayerID, "session.bubble.furniture.invalid_move", true},
		{errors.New("unmapped"), "", false},
	}

	for _, testCase := range cases {
		key, soft := bubbleErrorKey(testCase.err)
		if key != testCase.wantKey || soft != testCase.wantSoft {
			t.Fatalf("bubbleErrorKey(%v) = (%q, %v), want (%q, %v)", testCase.err, key, soft, testCase.wantKey, testCase.wantSoft)
		}
	}
}

// fakePublisher records published events for tests.
type fakePublisher struct {
	events []bus.Event
}

// Publish records a published event for tests.
func (publisher *fakePublisher) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)

	return nil
}

// TestHandleSkipsBroadcastWhenConnectionsNil verifies a nil connection registry is tolerated.
func TestHandleSkipsBroadcastWhenConnectionsNil(t *testing.T) {
	handler, standaloneConnections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, standaloneConnections, "conn")
	handler.Connections = nil
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("expected no error with nil connections, got %v", err)
	}
}

// TestHandleSkipsMissingRuntimeRoom verifies stale player presence remains a harmless no-op.
func TestHandleSkipsMissingRuntimeRoom(t *testing.T) {
	handler, _ := handlerForTest(t)
	handler.Runtime = roomlive.NewRegistry(nil)
	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), ItemID: 1},
	})
	if err != nil {
		t.Fatalf("expected missing runtime room to be ignored, got %v", err)
	}
}

// TestBroadcastHeightMapSkipsMissingDefinition verifies stale definition data is tolerated.
func TestBroadcastHeightMapSkipsMissingDefinition(t *testing.T) {
	handler, _ := handlerForTest(t)
	handler.Furniture = &fakeManager{}
	room, _ := handler.Runtime.Find(9)
	if err := handler.broadcastHeightMapUpdate(context.Background(), room, placedPickedItemForTest()); err != nil {
		t.Fatalf("expected missing definition to be ignored, got %v", err)
	}
}

// TestCommandName verifies the stable command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}
