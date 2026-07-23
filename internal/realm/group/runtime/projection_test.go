package runtime

import (
	"bytes"
	"context"
	"testing"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
	outobjectbatch "github.com/niflaot/pixels/networking/outbound/room/furniture/objectdata/batch"
)

// TestGroupFurnitureColorsProjectsLinkedObjects verifies supported live color projection stays scoped to linked furniture.
func TestGroupFurnitureColorsProjectsLinkedObjects(t *testing.T) {
	ctx := context.Background()
	metrics := groupobservability.New()
	cache := NewCache()
	cache.SetMetrics(metrics)
	cache.PutGroup(GroupSnapshot{Group: grouprecord.Group{
		ID: 2, BadgeCode: "b001120s004020", ColorAHex: "3C6400", ColorBHex: "E5E5E5",
	}})
	cache.PutFurnitureLinks(2, []int64{39})

	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(ctx, 9) })
	roomGrid, err := grid.Parse("00")
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Furniture: []worldfurniture.Item{{
			ID: 39, ExtraData: "0", Point: grid.MustPoint(1, 0),
			Definition: worldfurniture.Definition{Width: 1, Length: 1, StackHeight: grid.HeightFromInt(1)},
		}},
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
	}); err != nil {
		t.Fatal(err)
	}

	connections := netconn.NewRegistry()
	packets := make([]codec.Packet, 0, 1)
	session := returnSessionForTest(t, "viewer", &packets)
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(ctx, 9, returnedOccupantForTest(7, session.ID())); err != nil {
		t.Fatal(err)
	}
	projector := NewProjector(cache, rooms, connections, nil)
	projector.SetMetrics(metrics)
	if err = projector.GroupFurnitureColors(ctx, 2); err != nil {
		t.Fatal(err)
	}

	expected, err := outobjectbatch.Encode([]int64{39}, []*stuffdata.Data{stuffdata.StringArray([]string{
		"0", "2", "b001120s004020", "3C6400", "E5E5E5",
	})})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 || packets[0].Header != expected.Header || !bytes.Equal(packets[0].Payload, expected.Payload) {
		t.Fatalf("packets=%#v expected=%#v", packets, expected)
	}
}
