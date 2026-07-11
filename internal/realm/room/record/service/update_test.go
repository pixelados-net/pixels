package service

import (
	"context"
	"errors"
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// profanityForTest reports configured prohibited content.
type profanityForTest string

// UpdateRoom updates a room for tests.
func (store *fakeStore) UpdateRoom(_ context.Context, params UpdateRecordParams, tags []string) (roommodel.Room, bool, error) {
	if params.ExpectedVersion != store.room.Version.Version {
		return roommodel.Room{}, false, nil
	}
	store.room = params.Room
	store.room.Version.Version++
	store.tags = append([]string(nil), tags...)

	return store.room, store.found, nil
}

// Contains reports whether text equals the prohibited fixture.
func (profanity profanityForTest) Contains(_ context.Context, text string) (bool, error) {
	return text == string(profanity), nil
}

// TestUpdatePersistsNormalizedTagsAndPasswordHash verifies secure settings mutation.
func TestUpdatePersistsNormalizedTagsAndPasswordHash(t *testing.T) {
	store := newFakeStore()
	store.room.MaxUsers = 25
	name := "Updated Room"
	password := "1234"
	doorMode := roommodel.DoorModePassword
	tags := []string{" Social ", "social", "Build"}
	updated, err := New(store, fakeLayouts{}).Update(context.Background(), store.room.ID, store.room.Version.Version, UpdateParams{Name: &name, Password: &password, DoorMode: &doorMode, Tags: &tags})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != name || updated.PasswordHash == nil || *updated.PasswordHash == password {
		t.Fatalf("unexpected update %#v", updated)
	}
	if len(store.tags) != 2 || store.tags[0] != "social" {
		t.Fatalf("unexpected tags %#v", store.tags)
	}
}

// TestUpdateRejectsConflictPasswordAndProhibitedText verifies settings failure boundaries.
func TestUpdateRejectsConflictPasswordAndProhibitedText(t *testing.T) {
	store := newFakeStore()
	store.room.MaxUsers = 25
	doorMode := roommodel.DoorModePassword
	if _, err := New(store, fakeLayouts{}).Update(context.Background(), store.room.ID, 99, UpdateParams{}); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if _, err := New(store, fakeLayouts{}).Update(context.Background(), store.room.ID, store.room.Version.Version, UpdateParams{DoorMode: &doorMode}); !errors.Is(err, ErrPasswordRequired) {
		t.Fatalf("expected password required, got %v", err)
	}
	name := "blocked"
	if _, err := New(store, fakeLayouts{}).WithProfanity(profanityForTest("blocked")).Update(context.Background(), store.room.ID, store.room.Version.Version, UpdateParams{Name: &name}); !errors.Is(err, ErrProhibitedName) {
		t.Fatalf("expected prohibited text, got %v", err)
	}
}

// TestApplyAndValidateUpdateCoversEditableFields verifies complete partial-field projection.
func TestApplyAndValidateUpdateCoversEditableFields(t *testing.T) {
	name, description := "Complete Room", "Updated description"
	categoryID := int64(4)
	category := &categoryID
	maxUsers, wall, floor := 50, 1, -1
	door, trade := roommodel.DoorModeOpen, roommodel.TradeModeAllowed
	allowWalk, allowPets, allowEat, hideWalls := true, false, true, true
	chatMode, chatWeight, chatSpeed, chatDistance, chatProtection := int16(1), int16(2), int16(1), int16(40), int16(2)
	moderation := roommodel.ModerationPolicyOwnerAndRights
	room := roommodel.Room{Name: "Old Room", MaxUsers: 25}
	applyUpdate(&room, UpdateParams{Name: &name, Description: &description, CategoryID: &category, MaxUsers: &maxUsers,
		DoorMode: &door, TradeMode: &trade, AllowWalkthrough: &allowWalk, AllowPets: &allowPets,
		AllowPetsEat: &allowEat, HideWalls: &hideWalls, WallThickness: &wall, FloorThickness: &floor,
		ChatMode: &chatMode, ChatWeight: &chatWeight, ChatSpeed: &chatSpeed, ChatDistance: &chatDistance,
		ChatProtection: &chatProtection, ModerationMute: &moderation, ModerationKick: &moderation, ModerationBan: &moderation})
	if room.Name != name || room.CategoryID == nil || room.MaxUsers != 50 || room.TradeMode != trade || !room.AllowWalkthrough || room.AllowPets || !room.AllowPetsEat || !room.HideWalls || room.ChatDistance != 40 || room.ModerationBan != moderation {
		t.Fatalf("unexpected merged room %#v", room)
	}
}

// TestValidateUpdateRejectsMalformedSettings verifies every settings value family.
func TestValidateUpdateRejectsMalformedSettings(t *testing.T) {
	valid := roommodel.Room{Name: "Valid Room", MaxUsers: 25, ChatDistance: 50}
	tags := []string{"official:featured"}
	if err := validateUpdate(valid, UpdateParams{Tags: &tags}, tags); !errors.Is(err, ErrReservedTag) {
		t.Fatalf("expected reserved tag, got %v", err)
	}
	invalid := []struct {
		room     roommodel.Room
		expected error
	}{
		{roommodel.Room{Name: "x", MaxUsers: 25}, ErrInvalidRoomName},
		{func() roommodel.Room {
			room := valid
			room.Description = string(make([]byte, MaxRoomDescriptionLength+1))
			return room
		}(), ErrInvalidDescription},
		{func() roommodel.Room { room := valid; room.MaxUsers = 0; return room }(), ErrInvalidMaxUsers},
		{func() roommodel.Room { room := valid; room.DoorMode = 9; return room }(), ErrInvalidDoorMode},
		{func() roommodel.Room { room := valid; room.TradeMode = 9; return room }(), ErrInvalidTradeMode},
		{func() roommodel.Room { room := valid; room.WallThickness = 2; return room }(), ErrInvalidRoomID},
		{func() roommodel.Room { room := valid; room.ChatDistance = 101; return room }(), ErrInvalidChatSettings},
		{func() roommodel.Room { room := valid; room.ModerationMute = 9; return room }(), ErrInvalidModerationSettings},
	}
	for _, testCase := range invalid {
		if err := validateUpdate(testCase.room, UpdateParams{}, nil); !errors.Is(err, testCase.expected) {
			t.Fatalf("expected %v, got %v", testCase.expected, err)
		}
	}
}

// TestValidateCategoryHonorsVisibilityAndStaffCapability verifies category selection policy.
func TestValidateCategoryHonorsVisibilityAndStaffCapability(t *testing.T) {
	categoryID := int64(4)
	store := newFakeStore()
	store.categories = []roommodel.Category{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: categoryID}}, Visible: true, StaffOnly: true}}
	service := New(store, fakeLayouts{})
	if err := service.validateCategory(context.Background(), &categoryID, false); !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("expected staff category denial, got %v", err)
	}
	if err := service.validateCategory(context.Background(), &categoryID, true); err != nil {
		t.Fatalf("allow staff category: %v", err)
	}
}

// TestCreateValidatesCategoryAndContent verifies creation uses shared policies.
func TestCreateValidatesCategoryAndContent(t *testing.T) {
	categoryID := int64(4)
	store := newFakeStore()
	store.categories = []roommodel.Category{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: categoryID}}, Visible: true}}
	params := validCreateForTest()
	params.CategoryID = &categoryID
	if _, err := New(store, fakeLayouts{found: true, enabled: true}).Create(context.Background(), params); err != nil {
		t.Fatalf("create valid category: %v", err)
	}
	missingID := int64(5)
	params.CategoryID = &missingID
	if _, err := New(store, fakeLayouts{found: true, enabled: true}).Create(context.Background(), params); !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("missing category error=%v", err)
	}
	params.CategoryID = &categoryID
	params.Name = "blocked"
	if _, err := New(store, fakeLayouts{found: true, enabled: true}).WithProfanity(profanityForTest("blocked")).Create(context.Background(), params); !errors.Is(err, ErrProhibitedName) {
		t.Fatalf("prohibited name error=%v", err)
	}
}

// BenchmarkUpdateValidation measures in-memory settings merge and validation.
func BenchmarkUpdateValidation(b *testing.B) {
	service := New(newFakeStore(), fakeLayouts{})
	room := newFakeStore().room
	room.MaxUsers = 25
	name := "Benchmark Room"
	params := UpdateParams{Name: &name}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		if _, _, err := service.mergeUpdate(ctx, room, params); err != nil {
			b.Fatal(err)
		}
	}
}
