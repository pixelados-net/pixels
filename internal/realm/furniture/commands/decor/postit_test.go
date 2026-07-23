package decor

import (
	"bytes"
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	outplaced "github.com/niflaot/pixels/networking/outbound/inventory/furniture/postitplaced"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outitemdata "github.com/niflaot/pixels/networking/outbound/room/furniture/itemdata"
	outopen "github.com/niflaot/pixels/networking/outbound/room/furniture/postitopen"
	outwalladd "github.com/niflaot/pixels/networking/outbound/room/furniture/walladd"
	outwallupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/wallupdate"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestHandlePostItPlacesAndOpensEditor verifies inventory placement feedback.
func TestHandlePostItPlacesAndOpensEditor(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	handler.Furniture = &furnitureManager{item: inventoryDecoratorItem(1, 20, ""), definition: furnituremodel.Definition{InteractionType: "postit"}}
	value := Command{Handler: connection, Kind: KindPostItPlace, ItemID: 1, WallPosition: ":w=3,2 l=1,1 r"}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
		t.Fatalf("place post-it: %v", err)
	}
	if len(*sent) != 4 || (*sent)[0].Header != outremove.Header || (*sent)[1].Header != outwalladd.Header || (*sent)[2].Header != outplaced.Header || (*sent)[3].Header != outopen.Header {
		t.Fatalf("unexpected placement packets %#v", *sent)
	}
}

// TestHandlePostItConsecutivePlacementsRenderWithoutSave verifies a replaced editor cannot hide the previous note.
func TestHandlePostItConsecutivePlacementsRenderWithoutSave(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	manager := &furnitureManager{definition: furnituremodel.Definition{SpriteID: 1, InteractionType: "postit"}}
	handler.Furniture = manager
	for _, placement := range []struct {
		id       int64
		position string
	}{{id: 1, position: ":w=3,2 l=1,1 r"}, {id: 2, position: ":w=3,2 l=2,2 r"}} {
		manager.item = inventoryDecoratorItem(placement.id, 20, "0")
		value := Command{Handler: connection, Kind: KindPostItPlace, ItemID: placement.id, WallPosition: placement.position}
		if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
			t.Fatalf("place post-it %d: %v", placement.id, err)
		}
	}
	wallAdds := 0
	for _, packet := range *sent {
		if packet.Header != outwalladd.Header {
			continue
		}
		wallAdds++
		if !bytes.Contains(packet.Payload, []byte(roomdecor.DefaultPostItColor)) {
			t.Fatalf("wall add lacks renderable state: %q", packet.Payload)
		}
	}
	if wallAdds != 2 {
		t.Fatalf("expected both notes to render before either save, got %d", wallAdds)
	}
}

// TestHandlePostItRejectsGuestWithoutStickyPole verifies denied placement has client feedback.
func TestHandlePostItRejectsGuestWithoutStickyPole(t *testing.T) {
	handler, connection, sent, _ := decoratorFixtureForOwner(t, 8)
	handler.Furniture = &furnitureManager{item: inventoryDecoratorItem(1, 20, ""), definition: furnituremodel.Definition{InteractionType: "postit"}}
	value := Command{Handler: connection, Kind: KindPostItPlace, ItemID: 1, WallPosition: ":w=3,2 l=1,1 r"}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
		t.Fatalf("reject guest post-it: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outbubble.Header {
		t.Fatalf("expected no-rights bubble, got %#v", *sent)
	}
}

// TestHandlePostItAllowsGuestWithStickyPole verifies the classic guest-posting exception.
func TestHandlePostItAllowsGuestWithStickyPole(t *testing.T) {
	handler, connection, sent, active := decoratorFixtureForOwner(t, 8)
	handler.Furniture = &furnitureManager{item: inventoryDecoratorItem(1, 20, ""), definition: furnituremodel.Definition{InteractionType: "postit"}}
	pole := worldfurniture.Item{ID: 40, Point: grid.MustPoint(1, 0), Definition: worldfurniture.Definition{InteractionType: "sticky_pole"}}
	if _, err := active.ReloadFurniture(pole.ID, &pole); err != nil {
		t.Fatalf("place sticky pole: %v", err)
	}
	value := Command{Handler: connection, Kind: KindPostItPlace, ItemID: 1, WallPosition: ":w=3,2 l=1,1 r"}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
		t.Fatalf("place guest post-it: %v", err)
	}
	if len(*sent) != 4 || (*sent)[0].Header != outremove.Header || (*sent)[1].Header != outwalladd.Header || (*sent)[2].Header != outplaced.Header || (*sent)[3].Header != outopen.Header {
		t.Fatalf("unexpected guest placement packets %#v", *sent)
	}
}

// TestHandlePostItReadsSavesAndEditsData verifies every durable note operation.
func TestHandlePostItReadsSavesAndEditsData(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	roomID := int64(9)
	wallPosition := ":w=3,2 l=1,1 r"
	item := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 20, OwnerPlayerID: 7, RoomID: &roomID, WallPosition: &wallPosition, ExtraData: "FFFF33 old"}
	handler.Furniture = &furnitureManager{item: item, definition: furnituremodel.Definition{SpriteID: 1, InteractionType: "postit"}}
	handler.States = &stateUpdater{item: item}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, Kind: KindPostItGet, ItemID: 1}}); err != nil {
		t.Fatalf("read post-it: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outitemdata.Header {
		t.Fatalf("unexpected read packets %#v", *sent)
	}
	*sent = (*sent)[:0]
	save := Command{Handler: connection, Kind: KindPostItSave, ItemID: 1, Color: "9CCEFF", Text: "hello"}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: save}); err != nil {
		t.Fatalf("save post-it: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outwallupdate.Header {
		t.Fatalf("unexpected save packets %#v", *sent)
	}
	if !bytes.Contains((*sent)[0].Payload, []byte("9CCEFF")) || bytes.Contains((*sent)[0].Payload, []byte("hello")) {
		t.Fatalf("post-it wall update must contain color without text: %q", (*sent)[0].Payload)
	}
	*sent = (*sent)[:0]
	edit := Command{Handler: connection, Kind: KindPostItSet, ItemID: 1, Color: "FF9CFF", Text: "changed"}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: edit}); err != nil {
		t.Fatalf("edit post-it: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outwallupdate.Header {
		t.Fatalf("unexpected edit packets %#v", *sent)
	}
}

// TestPostItHelpersValidateColorDataAndStickyPole verifies legacy defaults and guest placement detection.
func TestPostItHelpersValidateColorDataAndStickyPole(t *testing.T) {
	if postItData("") != roomdecor.DefaultPostItColor || postItData("9CCEFF note") != "9CCEFF note" {
		t.Fatal("unexpected post-it data normalization")
	}
	if !validPostItColor("9CCEFF") || validPostItColor("000000") {
		t.Fatal("unexpected post-it color validation")
	}
	_, _, _, active := decoratorFixture(t)
	if hasStickyPole(active) {
		t.Fatal("empty room unexpectedly has a sticky pole")
	}
}
