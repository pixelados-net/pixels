package decor

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestParseToner verifies bounded enabled and RGB-like channel persistence.
func TestParseToner(t *testing.T) {
	enabled, hue, saturation, lightness, valid := parseToner("1:20:30:255")
	if !valid || enabled != 1 || hue != 20 || saturation != 30 || lightness != 255 {
		t.Fatalf("unexpected toner state %d:%d:%d:%d valid=%v", enabled, hue, saturation, lightness, valid)
	}
	for _, value := range []string{"2:1:2:3", "1:-1:2:3", "1:1:2:256", "broken"} {
		if _, _, _, _, ok := parseToner(value); ok {
			t.Fatalf("expected %q to fail", value)
		}
	}
}

// TestUseTonerTogglesExistingState verifies generic furniture clicks reuse the typed projection.
func TestUseTonerTogglesExistingState(t *testing.T) {
	handler, _, sent, active := decoratorFixture(t)
	roomID := int64(9)
	x, y, z := 1, 1, 0.0
	item := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 42, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z, ExtraData: "0:20:30:40"}
	handler.Furniture = &furnitureManager{item: item, definition: furnituremodel.Definition{SpriteID: 4697, InteractionType: "background_toner"}}
	handler.States = &stateUpdater{item: item}
	player, _ := handler.Players.Find(7)
	handled, err := handler.Use(context.Background(), player, active, worldfurniture.Item{ID: 1, Definition: worldfurniture.Definition{InteractionType: "background_toner"}})
	if err != nil || !handled {
		t.Fatalf("toggle toner handled=%t err=%v", handled, err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outupdate.Header {
		t.Fatalf("unexpected toner toggle packets %#v", *sent)
	}
}

// TestHandleTonerPersistsAndProjectsChannels verifies bounded HSL room updates.
func TestHandleTonerPersistsAndProjectsChannels(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	roomID := int64(9)
	x, y, z := 1, 1, 0.0
	item := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 42, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z, ExtraData: "0:0:0:0"}
	handler.Furniture = &furnitureManager{item: item, definition: furnituremodel.Definition{SpriteID: 4697, InteractionType: "background_toner"}}
	handler.States = &stateUpdater{item: item}
	value := Command{Handler: connection, Kind: KindTonerApply, ItemID: 1, First: 20, Second: 30, Third: 40}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
		t.Fatalf("apply toner: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outupdate.Header {
		t.Fatalf("unexpected toner packets %#v", *sent)
	}
}
