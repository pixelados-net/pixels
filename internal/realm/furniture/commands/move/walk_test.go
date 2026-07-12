package move

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestHandleMovingInteractionPublishesSyntheticWalkedOff verifies moved footprints release units.
func TestHandleMovingInteractionPublishesSyntheticWalkedOff(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, connections, "conn")
	active, _ := handler.Runtime.Find(9)
	previous := worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(1, 0), ExtraData: "0",
		Definition: worldfurniture.Definition{
			InteractionType: "default", InteractionModesCount: 2,
			Width: 1, Length: 1, AllowWalk: true,
		},
	}
	if _, err := active.ReloadFurniture(previous.ID, &previous); err != nil {
		t.Fatalf("load interaction: %v", err)
	}
	if _, err := active.TeleportUnit(7, previous.Point, worldunit.RotationNorth, false); err != nil {
		t.Fatalf("position unit: %v", err)
	}
	definition := chairDefinitionForTest()
	definition.InteractionType = "default"
	definition.InteractionModesCount = 2
	definition.AllowWalk = true
	handler.Furniture = &fakeManager{
		definition: definition, definitionFound: true,
		item: placedItemForTest(), itemFound: true,
	}
	publisher := &fakePublisher{}
	handler.Events = publisher
	if err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	}); err != nil {
		t.Fatalf("move interaction: %v", err)
	}
	found := false
	for _, event := range publisher.events {
		if event.Name == furniturewalkedoff.Name {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected synthetic walked-off event, got %#v", publisher.events)
	}
}

// TestPublishWalkedOffIgnoresRoller verifies roller movement owns its own unit transitions.
func TestPublishWalkedOffIgnoresRoller(t *testing.T) {
	handler, _ := handlerForTest(t)
	publisher := &fakePublisher{}
	handler.Events = publisher
	previous := worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(0, 0),
		Definition: worldfurniture.Definition{InteractionType: "roller", Width: 1, Length: 1},
	}
	if err := handler.publishWalkedOff(context.Background(), nil, previous, true, previous); err != nil {
		t.Fatalf("ignore roller: %v", err)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("expected no roller events, got %#v", publisher.events)
	}
}
