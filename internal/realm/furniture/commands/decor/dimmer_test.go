package decor

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	outsettings "github.com/niflaot/pixels/networking/outbound/room/dimmer/settings"
	outwallupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/wallupdate"
)

// TestHandleDimmerReadsAndSavesPresets verifies settings and wall state projection.
func TestHandleDimmerReadsAndSavesPresets(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	wallPosition := ":w=3,2 l=1,1 r"
	roomID := int64(9)
	handler.Furniture = &furnitureManager{item: furnituremodel.Item{RoomID: &roomID, WallPosition: &wallPosition, DefinitionID: 9, OwnerPlayerID: 7}, definition: furnituremodel.Definition{SpriteID: 4027, InteractionType: "dimmer"}}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, Kind: KindDimmerSettings}}); err != nil {
		t.Fatalf("load dimmer settings: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outsettings.Header {
		t.Fatalf("unexpected settings packets %#v", *sent)
	}
	*sent = (*sent)[:0]
	value := Command{Handler: connection, Kind: KindDimmerSave, PresetID: 1, Color: "#000000", First: 255, Apply: true}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
		t.Fatalf("save dimmer preset: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outwallupdate.Header || (*sent)[1].Header != outsettings.Header {
		t.Fatalf("unexpected save packets %#v", *sent)
	}
	*sent = (*sent)[:0]
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, Kind: KindDimmerToggle}}); err != nil {
		t.Fatalf("toggle dimmer: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outwallupdate.Header || (*sent)[1].Header != outsettings.Header {
		t.Fatalf("unexpected toggle packets %#v", *sent)
	}
}
