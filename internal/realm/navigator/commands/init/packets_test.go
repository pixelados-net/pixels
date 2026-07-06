package init

import (
	"testing"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	outcollapsed "github.com/niflaot/pixels/networking/outbound/navigator/collapsed"
	outevents "github.com/niflaot/pixels/networking/outbound/navigator/eventcategories"
	outfavorites "github.com/niflaot/pixels/networking/outbound/navigator/favorites"
	outflatcats "github.com/niflaot/pixels/networking/outbound/navigator/flatcats"
	outlifted "github.com/niflaot/pixels/networking/outbound/navigator/lifted"
	outmetadata "github.com/niflaot/pixels/networking/outbound/navigator/metadata"
	outsaved "github.com/niflaot/pixels/networking/outbound/navigator/savedsearches"
	outsettings "github.com/niflaot/pixels/networking/outbound/navigator/settings"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestInitialPacketsBuildsBootstrapOrder verifies navigator init packet order.
func TestInitialPacketsBuildsBootstrapOrder(t *testing.T) {
	packets, err := initialPackets(
		[]navmodel.SavedSearch{{Identity: sharedmodel.Identity{ID: 3}, Code: "hotel_view", Filter: "demo"}},
		navmodel.Preference{WindowX: 1, WindowY: 2, WindowWidth: 3, WindowHeight: 4},
		[]navmodel.LiftedRoom{{RoomID: 8, AreaID: 1}},
		[]int64{8},
		[]navmodel.CategoryPreference{{Code: "popular", Collapsed: true}},
		[]roommodel.Category{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, Caption: "Social", Visible: true}},
	)
	if err != nil {
		t.Fatalf("build packets: %v", err)
	}

	headers := []uint16{
		outmetadata.Header,
		outlifted.Header,
		outsaved.Header,
		outsettings.Header,
		outcollapsed.Header,
		outevents.Header,
		outfavorites.Header,
		outflatcats.Header,
	}
	for index, header := range headers {
		if packets[index].Header != header {
			t.Fatalf("packet %d expected header %d got %d", index, header, packets[index].Header)
		}
	}
}
