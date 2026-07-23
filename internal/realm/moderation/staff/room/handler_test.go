package room

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	roomrealm "github.com/niflaot/pixels/internal/realm/room"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/networking/inbound/moderation/staff/changeroom"
	"github.com/niflaot/pixels/pkg/i18n"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// permissionsForTest stores allowed nodes by player.
type permissionsForTest map[int64]map[permission.Node]bool

// HasPermission resolves one test permission.
func (permissions permissionsForTest) HasPermission(_ context.Context, playerID int64, node permission.Node) (bool, error) {
	return permissions[playerID][node], nil
}

// roomsForTest captures moderator room updates.
type roomsForTest struct {
	// updated stores the submitted update parameters.
	updated roomservice.UpdateParams
	// room stores the persistent room returned to handlers.
	room roommodel.Room
}

// Create satisfies the room configuration manager boundary.
func (*roomsForTest) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID satisfies the room configuration manager boundary.
func (rooms *roomsForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return rooms.room, rooms.room.ID > 0, nil
}

// ListByOwner satisfies the room configuration manager boundary.
func (*roomsForTest) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return nil, nil
}

// ListPopular satisfies the room configuration manager boundary.
func (*roomsForTest) ListPopular(context.Context, int) ([]roommodel.Room, error) { return nil, nil }

// ListHighestScore satisfies the room configuration manager boundary.
func (*roomsForTest) ListHighestScore(context.Context, int) ([]roommodel.Room, error) {
	return nil, nil
}

// Search satisfies the room configuration manager boundary.
func (*roomsForTest) Search(context.Context, string, int) ([]roommodel.Room, error) {
	return nil, nil
}

// ListTags satisfies the room configuration manager boundary.
func (*roomsForTest) ListTags(context.Context, int64) ([]roommodel.Tag, error) { return nil, nil }

// SoftDelete satisfies the room configuration manager boundary.
func (*roomsForTest) SoftDelete(context.Context, int64) error { return nil }

// ListCategories satisfies the room configuration manager boundary.
func (*roomsForTest) ListCategories(context.Context) ([]roommodel.Category, error) { return nil, nil }

// Update captures one room update.
func (rooms *roomsForTest) Update(_ context.Context, _ int64, _ int64, params roomservice.UpdateParams) (roommodel.Room, error) {
	rooms.updated = params
	updated := rooms.room
	if updated.ID == 0 {
		updated.Base.Identity.ID = 9
	}
	if params.Name != nil {
		updated.Name = *params.Name
	}
	if params.DoorMode != nil {
		updated.DoorMode = *params.DoorMode
	}
	updated.Version.Version++
	rooms.room = updated
	return updated, nil
}

// TestUpdateAppliesDoorbellAndLocalizedTitle verifies Nitro room override flags.
func TestUpdateAppliesDoorbellAndLocalizedTitle(t *testing.T) {
	rooms := &roomsForTest{}
	handler := Handler{Context: &moderationruntime.Context{Rooms: rooms}}
	handler.Translations = translationForTest{}
	updated, changed, err := handler.update(roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}}, changeroom.Payload{LockDoor: 1, ChangeTitle: 1})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !changed || updated.DoorMode != roommodel.DoorModeDoorbell || updated.Name != "Sala bajo revisión" {
		t.Fatalf("unexpected updated room: %+v changed=%v", updated, changed)
	}
}

// TestProtectedPreservesActorStaffAndUnkickable verifies mass-kick exclusions.
func TestProtectedPreservesActorStaffAndUnkickable(t *testing.T) {
	permissions := permissionsForTest{
		3: {permission.Node("moderation.tool.access"): true},
		4: {roomrealm.Unkickable: true},
	}
	handler := Handler{Context: &moderationruntime.Context{Permissions: permissions}}
	for _, playerID := range []int64{2, 3, 4} {
		protected, err := handler.protected(playerID, 2)
		if err != nil || !protected {
			t.Fatalf("player %d protected=%v err=%v", playerID, protected, err)
		}
	}
	protected, err := handler.protected(1, 2)
	if err != nil || protected {
		t.Fatalf("room owner protected=%v err=%v", protected, err)
	}
}

// translationForTest resolves the moderator replacement room title.
type translationForTest struct{}

// Default resolves one translation key.
func (translationForTest) Default(key i18n.Key, _ ...i18n.Params) string {
	if key == "moderation.room.inappropriate_name" {
		return "Sala bajo revisión"
	}
	return string(key)
}

// T resolves one explicit locale translation key.
func (translationForTest) T(_ i18n.Locale, key i18n.Key, params ...i18n.Params) string {
	return translationForTest{}.Default(key, params...)
}

// Entries returns the test translation catalog.
func (translationForTest) Entries(i18n.Locale) map[i18n.Key]string { return nil }
