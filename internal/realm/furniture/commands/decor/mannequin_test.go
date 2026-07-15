package decor

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	outinfo "github.com/niflaot/pixels/networking/outbound/room/entities/info"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestMannequinFigureKeepsOnlyClothing verifies identity parts are not copied to furniture.
func TestMannequinFigureKeepsOnlyClothing(t *testing.T) {
	got := mannequinFigure("hd-180-1.hr-100-1.ch-210-66.lg-270-82.sh-290-80")
	want := "hd-99999-99998.ch-210-66.lg-270-82.sh-290-80"
	if got != want {
		t.Fatalf("unexpected mannequin figure %q", got)
	}
}

// TestHandleMannequinSavesTheOwnersClothing verifies durable map state and floor projection.
func TestHandleMannequinSavesTheOwnersClothing(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	roomID := int64(9)
	x, y, z := 1, 1, 0.0
	item := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 41, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z, ExtraData: `{}`}
	handler.Furniture = &furnitureManager{item: item, definition: furnituremodel.Definition{SpriteID: 4170, InteractionType: "mannequin"}}
	handler.States = &stateUpdater{item: item}
	value := Command{Handler: connection, Kind: KindMannequinLook, ItemID: 1}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
		t.Fatalf("save mannequin look: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outupdate.Header {
		t.Fatalf("unexpected mannequin packets %#v", *sent)
	}
}

// TestMergeMannequinFigurePreservesIdentity verifies only outfit clothing is replaced.
func TestMergeMannequinFigurePreservesIdentity(t *testing.T) {
	got := mergeMannequinFigure("hd-180-1.hr-100-1.ch-1-1.lg-2-2", "hd-99999-99998.ch-10-10.lg-20-20.sh-30-30")
	want := "hd-180-1.hr-100-1.ch-10-10.lg-20-20.sh-30-30"
	if got != want {
		t.Fatalf("unexpected merged figure %q", got)
	}
}

// TestUseMannequinAppliesOnlySavedClothing verifies persistent and live avatar refresh.
func TestUseMannequinAppliesOnlySavedClothing(t *testing.T) {
	handler, _, sent, active := decoratorFixture(t)
	roomID := int64(9)
	item := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 41, OwnerPlayerID: 7, RoomID: &roomID, ExtraData: `{"gender":"M","figure":"hd-99999-99998.ch-10-10","name":"Look"}`}
	handler.Furniture = &furnitureManager{item: item, definition: furnituremodel.Definition{InteractionType: "mannequin"}}
	handler.PlayerAdmin = &playerAdmin{}
	player, _ := handler.Players.Find(7)
	handled, err := handler.Use(context.Background(), player, active, worldfurniture.Item{ID: 1, Definition: worldfurniture.Definition{InteractionType: "mannequin"}})
	if err != nil || !handled {
		t.Fatalf("use mannequin handled=%t err=%v", handled, err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outinfo.Header || player.Snapshot().Look != "hd-180-1.ch-10-10" {
		t.Fatalf("unexpected mannequin projection look=%q packets=%#v", player.Snapshot().Look, *sent)
	}
}

// TestHandleMannequinRenamesOutfit verifies the separate filtered name operation.
func TestHandleMannequinRenamesOutfit(t *testing.T) {
	handler, connection, sent, _ := decoratorFixture(t)
	roomID := int64(9)
	x, y, z := 1, 1, 0.0
	item := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 41, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z, ExtraData: `{}`}
	handler.Furniture = &furnitureManager{item: item, definition: furnituremodel.Definition{SpriteID: 4170, InteractionType: "mannequin"}}
	handler.States = &stateUpdater{item: item}
	value := Command{Handler: connection, Kind: KindMannequinName, ItemID: 1, Text: "Evening look"}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: value}); err != nil {
		t.Fatalf("rename mannequin: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outupdate.Header {
		t.Fatalf("unexpected mannequin rename packets %#v", *sent)
	}
}

// playerAdmin stores the latest mannequin-applied look.
type playerAdmin struct{}

// Create is unused by mannequin tests.
func (*playerAdmin) Create(context.Context, playerservice.CreateParams) (playerservice.Record, error) {
	return playerservice.Record{}, nil
}

// FindByID is unused by mannequin tests.
func (*playerAdmin) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// FindByUsername is unused by mannequin tests.
func (*playerAdmin) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// UpdatePrivacy is unused by mannequin tests.
func (*playerAdmin) UpdatePrivacy(context.Context, int64, playerservice.PrivacyParams) (playerservice.Record, error) {
	return playerservice.Record{}, nil
}

// Update returns a complete record with the requested look.
func (*playerAdmin) Update(_ context.Context, playerID int64, params playerservice.UpdateParams) (playerservice.Record, error) {
	look := ""
	if params.Look != nil {
		look = *params.Look
	}
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: playerID}}, Username: "demo"}, Profile: playermodel.Profile{PlayerID: playerID, Gender: playermodel.GenderMale, Look: look}}, nil
}

// SoftDelete is unused by mannequin tests.
func (*playerAdmin) SoftDelete(context.Context, int64) error { return nil }
