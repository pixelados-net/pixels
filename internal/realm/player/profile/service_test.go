package profile

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	playerfigure "github.com/niflaot/pixels/internal/realm/player/figure"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
)

// profileStore records focused profile persistence calls.
type profileStore struct {
	// tags stores the last normalized replacement.
	tags []string
	// result stores the next respect result.
	result RespectResult
}

// profileAdmin applies profile fields to one deterministic record.
type profileAdmin struct {
	playerservice.AdminManager
	record playerservice.Record
	// err stores one injected mutation failure.
	err error
}

// profilePermissions returns one respect quota policy.
type profilePermissions bool

// HasPermission returns the configured policy decision.
func (allowed profilePermissions) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return bool(allowed), nil
}

// profileClothing returns one clothing unlock snapshot.
type profileClothing struct {
	sets []int32
}

// Clothing returns the configured clothing unlock snapshot.
func (clothing profileClothing) Clothing(context.Context, int64) (playerwardrobe.ClothingSnapshot, error) {
	return playerwardrobe.ClothingSnapshot{FigureSetIDs: clothing.sets}, nil
}

// Update applies figure, gender, and motto fields used by profile behavior.
func (admin *profileAdmin) Update(_ context.Context, _ int64, params playerservice.UpdateParams) (playerservice.Record, error) {
	if admin.err != nil {
		return playerservice.Record{}, admin.err
	}
	if params.Look != nil {
		admin.record.Profile.Look = *params.Look
	}
	if params.Gender != nil {
		admin.record.Profile.Gender = *params.Gender
	}
	if params.Motto != nil {
		admin.record.Profile.Motto = *params.Motto
	}
	return admin.record, nil
}

// Tags returns stored tags.
func (store *profileStore) Tags(context.Context, int64) ([]string, error) { return store.tags, nil }

// ReplaceTags stores normalized tags.
func (store *profileStore) ReplaceTags(_ context.Context, _ int64, tags []string) error {
	store.tags = append([]string(nil), tags...)
	return nil
}

// RespectState returns fixed allowance state.
func (*profileStore) RespectState(context.Context, int64, time.Time, int, int) (RespectState, error) {
	return RespectState{Received: 2, UserRemaining: 1, PetRemaining: 3}, nil
}

// GrantRespect returns the configured serialized result.
func (store *profileStore) GrantRespect(context.Context, int64, int64, time.Time, int, bool) (RespectResult, error) {
	return store.result, nil
}

// TestReplaceTagsNormalizesAndRejectsDuplicates verifies stable bounded tag replacement.
func TestReplaceTagsNormalizesAndRejectsDuplicates(t *testing.T) {
	store := &profileStore{}
	service := New(store, nil)
	if err := service.ReplaceTags(context.Background(), 1, []string{" Builder ", "Trade"}); err != nil {
		t.Fatalf("replace tags: %v", err)
	}
	if len(store.tags) != 2 || store.tags[0] != "Builder" {
		t.Fatalf("unexpected tags %#v", store.tags)
	}
	if err := service.ReplaceTags(context.Background(), 1, []string{"Trade", "trade"}); !errors.Is(err, ErrInvalidTags) {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
}

// TestGrantRespectValidatesAndMapsExhaustion verifies identity and quota policy.
func TestGrantRespectValidatesAndMapsExhaustion(t *testing.T) {
	store := &profileStore{}
	service := New(store, nil)
	if _, err := service.GrantRespect(context.Background(), 1, 1); !errors.Is(err, ErrRespectNotAllowed) {
		t.Fatalf("expected self respect rejection, got %v", err)
	}
	if _, err := service.GrantRespect(context.Background(), 1, 2); !errors.Is(err, ErrRespectExhausted) {
		t.Fatalf("expected exhaustion, got %v", err)
	}
	store.result = RespectResult{Applied: true, TotalReceived: 9, Remaining: 2}
	result, err := service.GrantRespect(context.Background(), 1, 2)
	if err != nil || result.TotalReceived != 9 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

// TestGrantRespectMapsDuplicate verifies a repeated target has a distinct outcome.
func TestGrantRespectMapsDuplicate(t *testing.T) {
	service := New(&profileStore{result: RespectResult{Duplicate: true, Remaining: 2}}, nil)
	_, err := service.GrantRespect(context.Background(), 1, 2)
	if !errors.Is(err, ErrRespectAlreadyGranted) {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
}

// TestProfileMutationsAndState verifies figure, motto, tag reads, and quota state.
func TestProfileMutationsAndState(t *testing.T) {
	store := &profileStore{tags: []string{"builder"}}
	admin := &profileAdmin{}
	service := New(store, admin)
	if _, err := service.UpdateFigure(context.Background(), 1, "X", "hd-180-1"); !errors.Is(err, ErrInvalidFigure) {
		t.Fatalf("expected invalid gender, got %v", err)
	}
	record, err := service.UpdateFigure(context.Background(), 1, "m", "hd-180-1")
	if err != nil || record.Profile.Look != "hd-180-1" {
		t.Fatalf("record=%#v err=%v", record, err)
	}
	if _, err = service.UpdateMotto(context.Background(), 1, "Pixels"); err != nil {
		t.Fatalf("motto: %v", err)
	}
	if _, err = service.UpdateMotto(context.Background(), 1, string(make([]rune, DefaultConfig().MottoMaximumRunes+1))); !errors.Is(err, ErrInvalidMotto) {
		t.Fatalf("expected invalid motto, got %v", err)
	}
	tags, err := service.Tags(context.Background(), 1)
	if err != nil || len(tags) != 1 || tags[0] != "builder" {
		t.Fatalf("tags=%#v err=%v", tags, err)
	}
	state, err := service.RespectState(context.Background(), 1)
	if err != nil || state.Received != 2 || state.UserRemaining != 1 {
		t.Fatalf("state=%#v err=%v", state, err)
	}
}

// TestProfileConfigFallsBackAndLoads verifies configuration boundaries.
func TestProfileConfigFallsBackAndLoads(t *testing.T) {
	service := newService(&profileStore{}, nil, nil, nil, Config{})
	if service.config.MottoMaximumRunes != DefaultConfig().MottoMaximumRunes {
		t.Fatalf("config=%#v", service.config)
	}
	t.Setenv("PIXELS_PLAYER_MOTTO_MAX_RUNES", "20")
	config, err := LoadConfig()
	if err != nil || config.MottoMaximumRunes != 20 {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}

// TestConfiguredFigureEntitlementAndUnlimitedRespect verifies injected policy boundaries.
func TestConfiguredFigureEntitlementAndUnlimitedRespect(t *testing.T) {
	path := filepath.Join(t.TempDir(), "figuredata.xml")
	data := []byte(`<figuredata><colors><palette id="1"><color id="1" club="0" selectable="1"/></palette></colors><sets><settype type="ha" paletteid="1"><set id="400" gender="U" club="0" selectable="1" sellable="1"/></settype></sets></figuredata>`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := playerfigure.NewCatalog(playerfigure.Config{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	store := &profileStore{}
	service := NewConfigured(store, &profileAdmin{}, profilePermissions(true), nil, catalog, profileClothing{sets: []int32{400}}, nil, DefaultConfig())
	if _, err = service.UpdateFigure(context.Background(), 1, "M", "ha-400-1"); err != nil {
		t.Fatalf("entitled figure: %v", err)
	}
	service.unlocks = profileClothing{}
	if _, err = service.UpdateFigure(context.Background(), 1, "M", "ha-400-1"); !errors.Is(err, ErrInvalidFigure) {
		t.Fatalf("expected entitlement rejection, got %v", err)
	}
	state, err := service.RespectState(context.Background(), 1)
	if err != nil || state.UserRemaining != int32(DefaultConfig().DailyRespectLimit) {
		t.Fatalf("state=%#v err=%v", state, err)
	}
}

// playerAdminStub documents the complete player administration boundary used by profile.
type playerAdminStub struct{ playerservice.AdminManager }
