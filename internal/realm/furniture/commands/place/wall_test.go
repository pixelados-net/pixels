package place

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outwalladd "github.com/niflaot/pixels/networking/outbound/room/furniture/walladd"
)

// TestHandlePlacesGenericWallFurniture verifies inventory, persistence, and room projection.
func TestHandlePlacesGenericWallFurniture(t *testing.T) {
	handler, connections := handlerForTest(t)
	actor, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{
		item: inventoryItemForTest(), itemFound: true,
		definition: furnituremodel.Definition{Kind: furnituremodel.KindWall, SpriteID: 9, InteractionType: "dimmer"}, definitionFound: true,
	}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Handler: actor, ItemID: 1, WallPosition: ":w=3,2 l=1,1 r",
	}})
	if err != nil {
		t.Fatalf("place wall item: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outremove.Header || (*sent)[1].Header != outwalladd.Header {
		t.Fatalf("unexpected wall placement packets %#v", *sent)
	}
}

// TestHandleRejectsMalformedWallPosition verifies invalid wall input never mutates inventory.
func TestHandleRejectsMalformedWallPosition(t *testing.T) {
	handler, connections := handlerForTest(t)
	actor, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{
		item: inventoryItemForTest(), itemFound: true,
		definition: furnituremodel.Definition{Kind: furnituremodel.KindWall, SpriteID: 9}, definitionFound: true,
	}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: actor, ItemID: 1, WallPosition: "invalid"}}); err != nil {
		t.Fatalf("reject malformed wall placement: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("unexpected packets %#v", *sent)
	}
}
