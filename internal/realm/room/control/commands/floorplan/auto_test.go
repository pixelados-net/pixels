package floorplan

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// furnitureForTest stores mutable persistent room furniture.
type furnitureForTest struct {
	furnitureservice.Manager
	// items stores currently placed furniture.
	items []furnituremodel.Item
	// definitions stores furniture metadata.
	definitions []furnituremodel.Definition
	// picked stores returned item ids.
	picked []int64
}

// ListRoomItems lists placed items.
func (furniture *furnitureForTest) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return append([]furnituremodel.Item(nil), furniture.items...), nil
}

// ListDefinitions lists furniture definitions.
func (furniture *furnitureForTest) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return furniture.definitions, nil
}

// Pickup returns one placed item to inventory.
func (furniture *furnitureForTest) Pickup(_ context.Context, params furnitureservice.PickupParams) (furnituremodel.Item, error) {
	for index, item := range furniture.items {
		if item.ID != params.ItemID {
			continue
		}
		furniture.items = append(furniture.items[:index], furniture.items[index+1:]...)
		furniture.picked = append(furniture.picked, item.ID)
		item.RoomID, item.X, item.Y, item.Z = nil, nil, nil, nil

		return item, nil
	}

	return furnituremodel.Item{}, furnitureservice.ErrItemNotFound
}

// TestSaveHandleAutoPicksBlockingFurniture verifies internal auto-pickup ownership and refresh.
func TestSaveHandleAutoPicksBlockingFurniture(t *testing.T) {
	players, bindings, player := floorplanActorForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7, OwnerName: "demo", ModelName: "model_a", MaxUsers: 5}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 5})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	initial, _ := grid.Parse("00", grid.WithDoor(0, 0))
	worldItem := worldfurniture.Item{ID: 5, Point: grid.MustPoint(1, 0), Definition: worldfurniture.Definition{Width: 1, Length: 1}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: initial, Furniture: []worldfurniture.Item{worldItem}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err = active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	roomID, x, y, z := int64(9), 1, 0, float64(0)
	definition := furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 3}}, Width: 1, Length: 1, Kind: furnituremodel.KindFloor}
	furniture := &furnitureForTest{
		items:       []furnituremodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 5}}, DefinitionID: 3, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z}},
		definitions: []furnituremodel.Definition{definition},
	}
	connections := netconn.NewRegistry()
	sent := registerFloorplanConnectionForTest(t, connections)
	authorizer := domain.NewAuthorizer(permissionsForTest{"own": true}, nil, domain.Nodes{OwnEdit: "own", AnyEdit: "any"})
	handler := SaveHandler{
		Players: players, Bindings: bindings, Rooms: roomsForTest{room: room}, Layouts: &layoutsForTest{},
		Furniture: furniture, Runtime: runtime, Connections: connections, Authorize: authorizer,
		Config: domain.Config{RejectZeroEffectiveHeight: true},
	}
	input := SaveCommand{Handler: netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, Params: domain.SaveParams{
		Heightmap: "01", DoorX: 0, DoorY: 0, DoorDirection: 2, WallHeight: -1, AutoPickup: true,
	}}
	if err = handler.Handle(context.Background(), command.Envelope[SaveCommand]{Command: input}); err != nil {
		t.Fatalf("save with auto-pickup: %v", err)
	}
	if len(furniture.picked) != 1 || furniture.picked[0] != 5 || len(*sent) != 3 {
		t.Fatalf("picked=%#v packets=%d", furniture.picked, len(*sent))
	}
	if (*sent)[0].Header != outforward.Header || (*sent)[1].Header != outunseen.Header || (*sent)[2].Header != outrefresh.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
}
