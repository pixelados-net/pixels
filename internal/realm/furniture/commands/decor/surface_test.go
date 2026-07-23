package decor

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outpaint "github.com/niflaot/pixels/networking/outbound/room/paint"
)

// TestHandleSurfaceConsumesAndProjectsAppearance verifies room paint, removal, and inventory refresh ordering.
func TestHandleSurfaceConsumesAndProjectsAppearance(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	handler.Furniture = &furnitureManager{item: inventoryDecoratorItem(1, 21, "101"), definition: furnituremodel.Definition{Name: "floor"}}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, Kind: KindSurface, ItemID: 1}})
	if err != nil {
		t.Fatalf("apply room surface: %v", err)
	}
	want := []uint16{outpaint.Header, outremove.Header, outrefresh.Header}
	if len(*sent) != len(want) {
		t.Fatalf("unexpected packet count %d", len(*sent))
	}
	for index, header := range want {
		if (*sent)[index].Header != header {
			t.Fatalf("packet %d header=%d want=%d", index, (*sent)[index].Header, header)
		}
	}
}
