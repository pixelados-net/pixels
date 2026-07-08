package enter

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	outflooritems "github.com/niflaot/pixels/networking/outbound/room/furniture/flooritems"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestSendFloorItemsSkipsWithoutFurnitureManager verifies the nil-manager guard.
func TestSendFloorItemsSkipsWithoutFurnitureManager(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)

	err := (Handler{}).sendFloorItems(context.Background(), connection, roomForTest(), nil)
	if err != nil {
		t.Fatalf("send floor items: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no packets, got %#v", *sent)
	}
}

// TestSendFloorItemsSkipsWhenRoomHasNoItems verifies the empty-room fast path.
func TestSendFloorItemsSkipsWhenRoomHasNoItems(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	handler := Handler{Furniture: furnitureManagerForTest{}}

	err := handler.sendFloorItems(context.Background(), connection, roomForTest(), nil)
	if err != nil {
		t.Fatalf("send floor items: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no packets, got %#v", *sent)
	}
}

// TestSendFloorItemsSendsPacketWithPlacedItems verifies the populated-room path.
func TestSendFloorItemsSendsPacketWithPlacedItems(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	handler := Handler{Furniture: furnitureManagerForTest{
		definitions: []furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, SpriteID: 39, AllowSit: true}},
		items:       []furnituremodel.Item{placedFurnitureItemForTest(1, 2, 1, 3, 3)},
	}}

	err := handler.sendFloorItems(context.Background(), connection, roomForTest(), nil)
	if err != nil {
		t.Fatalf("send floor items: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outflooritems.Header {
		t.Fatalf("expected one floor items packet, got %#v", *sent)
	}
}

// TestSendFloorItemsPropagatesStoreErrors verifies persistence errors surface.
func TestSendFloorItemsPropagatesStoreErrors(t *testing.T) {
	connection, _ := sessionConnectionForTest(t)
	expected := errors.New("store failed")
	handler := Handler{Furniture: furnitureManagerForTest{
		items: []furnituremodel.Item{placedFurnitureItemForTest(1, 2, 1, 3, 3)},
		err:   expected,
	}}

	err := handler.sendFloorItems(context.Background(), connection, roomForTest(), nil)
	if !errors.Is(err, expected) {
		t.Fatalf("expected store error, got %v", err)
	}
}

// TestHandleJoinsRoomWithFurnitureSendsFloorItemsPacket verifies full command orchestration with furniture.
func TestHandleJoinsRoomWithFurnitureSendsFloorItemsPacket(t *testing.T) {
	player := playerForTest(t)
	connection, sent := sessionConnectionForTest(t)
	handler := Handler{
		Players:  playerRegistryForTest(t, player),
		Bindings: bindingRegistryForTest(t, 7),
		Rooms:    roomManagerForTest{room: roomForTest(), found: true},
		Layouts:  layoutManagerForTest{roomLayout: layoutForTest(), found: true},
		Furniture: furnitureManagerForTest{
			definitions: []furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, SpriteID: 39, AllowSit: true}},
			items:       []furnituremodel.Item{placedFurnitureItemForTest(1, 2, 1, 0, 0)},
		},
		Runtime: newRuntimeForFurnitureTest(t),
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, RoomID: 9},
	})
	if err != nil {
		t.Fatalf("handle command: %v", err)
	}
	if len(*sent) != 7 {
		t.Fatalf("expected entered, model, floor items, and room state packets, got %#v", *sent)
	}

	active, found := handler.Runtime.Find(9)
	if !found {
		t.Fatal("expected active room")
	}
	if items := active.FurnitureItems(); len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("expected loaded furniture snapshot, got %#v", items)
	}
}

// newRuntimeForFurnitureTest creates an active room registry for furniture command tests.
func newRuntimeForFurnitureTest(t *testing.T) *roomlive.Registry {
	t.Helper()

	return roomlive.NewRegistry(nil)
}
