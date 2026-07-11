package settings

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	wordfiltermodified "github.com/niflaot/pixels/internal/realm/room/control/events/wordfiltermodified"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// filtersForTest captures filter mutations.
type filtersForTest struct {
	// added stores the last added word.
	added string
	// removed stores the last removed word.
	removed string
}

// Add captures one added word.
func (filters *filtersForTest) Add(_ context.Context, _ int64, _ int64, word string) error {
	filters.added = word
	return nil
}

// Remove captures one removed word.
func (filters *filtersForTest) Remove(_ context.Context, _ int64, _ int64, word string) error {
	filters.removed = word
	return nil
}

// filterEventsForTest captures published events.
type filterEventsForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish captures one event.
func (publisher *filterEventsForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// TestHandleMutatesAndPublishesFilter verifies room filter command behavior.
func TestHandleMutatesAndPublishesFilter(t *testing.T) {
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err = player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	if err = players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	filters, events := &filtersForTest{}, &filterEventsForTest{}
	handler := FilterModifyHandler{Players: players, Bindings: bindings, Filters: filters, Events: events}
	err = handler.Handle(context.Background(), command.Envelope[FilterModifyCommand]{Command: FilterModifyCommand{Handler: netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, RoomID: 9, Add: true, Word: " Spam "}})
	if err != nil {
		t.Fatalf("modify filter: %v", err)
	}
	if filters.added != " Spam " || len(events.events) != 1 || events.events[0].Name != wordfiltermodified.Name {
		t.Fatalf("filter=%#v events=%#v", filters, events.events)
	}
	payload, ok := events.events[0].Payload.(wordfiltermodified.Payload)
	if !ok || payload.Word != "spam" || !payload.Added {
		t.Fatalf("unexpected payload %#v", events.events[0].Payload)
	}
}
