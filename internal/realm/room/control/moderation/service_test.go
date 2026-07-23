package moderation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// moderationStoreForTest stores active sanctions in memory.
type moderationStoreForTest struct {
	// mutes stores mute expiry by player.
	mutes map[int64]time.Time
	// bans stores ban expiry by player.
	bans map[int64]time.Time
}

// WithinTransaction runs work directly.
func (store *moderationStoreForTest) WithinTransaction(ctx context.Context, work TransactionWork) error {
	return work(ctx)
}

// Mute creates a mute.
func (store *moderationStoreForTest) Mute(_ context.Context, _ int64, playerID int64, endsAt time.Time) error {
	store.mutes[playerID] = endsAt
	return nil
}

// Unmute removes an active mute.
func (store *moderationStoreForTest) Unmute(_ context.Context, _ int64, playerID int64, _ time.Time) (bool, error) {
	_, found := store.mutes[playerID]
	delete(store.mutes, playerID)
	return found, nil
}

// Ban creates a ban.
func (store *moderationStoreForTest) Ban(_ context.Context, _ int64, playerID int64, endsAt time.Time) error {
	store.bans[playerID] = endsAt
	return nil
}

// Unban removes an active ban.
func (store *moderationStoreForTest) Unban(_ context.Context, _ int64, playerID int64, _ time.Time) (bool, error) {
	_, found := store.bans[playerID]
	delete(store.bans, playerID)
	return found, nil
}

// IsMuted reports whether a mute is current.
func (store *moderationStoreForTest) IsMuted(_ context.Context, _ int64, playerID int64, now time.Time) (bool, error) {
	return store.mutes[playerID].After(now), nil
}

// IsBanned reports whether a ban is current.
func (store *moderationStoreForTest) IsBanned(_ context.Context, _ int64, playerID int64, now time.Time) (bool, error) {
	return store.bans[playerID].After(now), nil
}

// ListMutes lists active mutes.
func (*moderationStoreForTest) ListMutes(context.Context, int64, time.Time) ([]moderationmodel.Sanction, error) {
	return nil, nil
}

// ListBans lists active bans.
func (*moderationStoreForTest) ListBans(context.Context, int64, time.Time) ([]moderationmodel.Sanction, error) {
	return nil, nil
}

// moderationRoomForTest stores room moderation policy.
type moderationRoomForTest struct {
	// room stores the room result.
	room roommodel.Room
}

// FindByID finds the configured room.
func (finder moderationRoomForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return finder.room, true, nil
}

// moderationRightsForTest stores rights decisions.
type moderationRightsForTest bool

// HasRights returns the configured decision.
func (allowed moderationRightsForTest) HasRights(context.Context, int64, int64) (bool, error) {
	return bool(allowed), nil
}

// moderationPermissionsForTest stores player node decisions.
type moderationPermissionsForTest map[int64]map[permission.Node]bool

// HasPermission returns one configured decision.
func (permissions moderationPermissionsForTest) HasPermission(_ context.Context, playerID int64, node permission.Node) (bool, error) {
	return permissions[playerID][node], nil
}

// moderationServiceForTest creates a deterministic moderation service.
func moderationServiceForTest(store *moderationStoreForTest, room roommodel.Room, rights RightsChecker, permissions moderationPermissionsForTest, nodes Nodes) *Service {
	service := New(Config{MinMuteMinutes: 1, MaxMuteMinutes: 60}, store, moderationRoomForTest{room: room}, rights, permissions, nil, nodes)
	service.now = func() time.Time { return time.Unix(1000, 0) }
	return service
}

// TestServiceModeratesWithOwnerPolicy verifies mute, ban, and expiry checks.
func TestServiceModeratesWithOwnerPolicy(t *testing.T) {
	store := &moderationStoreForTest{mutes: make(map[int64]time.Time), bans: make(map[int64]time.Time)}
	nodes := Nodes{OwnMute: "own.mute", OwnBan: "own.ban"}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	permissions := moderationPermissionsForTest{1: {nodes.OwnMute: true, nodes.OwnBan: true}}
	service := moderationServiceForTest(store, room, nil, permissions, nodes)

	if err := service.Mute(context.Background(), 9, 1, 2, 5); err != nil {
		t.Fatalf("mute: %v", err)
	}
	if muted, err := service.IsMuted(context.Background(), 9, 2); err != nil || !muted {
		t.Fatalf("expected active mute, muted=%v err=%v", muted, err)
	}
	if err := service.Ban(context.Background(), 9, 1, 2, moderationmodel.BanDurationHour); err != nil {
		t.Fatalf("ban: %v", err)
	}
	if banned, err := service.IsBanned(context.Background(), 9, 2); err != nil || !banned {
		t.Fatalf("expected active ban, banned=%v err=%v", banned, err)
	}
}

// TestServiceAllowsRightsPolicyAndRejectsProtectedTarget verifies combined authorization.
func TestServiceAllowsRightsPolicyAndRejectsProtectedTarget(t *testing.T) {
	store := &moderationStoreForTest{mutes: make(map[int64]time.Time), bans: make(map[int64]time.Time)}
	nodes := Nodes{OwnKick: "own.kick", Unkickable: "unkickable"}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1, ModerationKick: roommodel.ModerationPolicyOwnerAndRights}
	permissions := moderationPermissionsForTest{3: {nodes.OwnKick: true}}
	service := moderationServiceForTest(store, room, moderationRightsForTest(true), permissions, nodes)
	if err := service.Kick(context.Background(), 9, 3, 2); err != nil {
		t.Fatalf("rights holder kick: %v", err)
	}
	permissions[2] = map[permission.Node]bool{nodes.Unkickable: true}
	if err := service.Kick(context.Background(), 9, 3, 2); !errors.Is(err, ErrTargetProtected) {
		t.Fatalf("expected target protected, got %v", err)
	}
}

// TestSystemModerationPreservesTargetImmunity verifies WIRED cannot bypass owner or staff protection.
func TestSystemModerationPreservesTargetImmunity(t *testing.T) {
	store := &moderationStoreForTest{mutes: make(map[int64]time.Time), bans: make(map[int64]time.Time)}
	nodes := Nodes{Unkickable: "unkickable"}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	permissions := moderationPermissionsForTest{3: {nodes.Unkickable: true}}
	service := moderationServiceForTest(store, room, nil, permissions, nodes)
	if err := service.SystemKick(context.Background(), 9, 1); !errors.Is(err, ErrTargetOwner) {
		t.Fatalf("owner system kick error=%v", err)
	}
	if err := service.SystemMute(context.Background(), 9, 3, 5); !errors.Is(err, ErrTargetProtected) {
		t.Fatalf("protected system mute error=%v", err)
	}
	if err := service.SystemKick(context.Background(), 9, 2); err != nil {
		t.Fatalf("ordinary system kick error=%v", err)
	}
}

// TestServiceRejectsInvalidDurationsAndTargets verifies validation boundaries.
func TestServiceRejectsInvalidDurationsAndTargets(t *testing.T) {
	store := &moderationStoreForTest{mutes: make(map[int64]time.Time), bans: make(map[int64]time.Time)}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	service := moderationServiceForTest(store, room, nil, nil, Nodes{})
	if err := service.Mute(context.Background(), 9, 1, 2, 0); !errors.Is(err, ErrInvalidMuteDuration) {
		t.Fatalf("expected invalid mute duration, got %v", err)
	}
	if err := service.Ban(context.Background(), 9, 1, 2, "invalid"); !errors.Is(err, ErrInvalidBanDuration) {
		t.Fatalf("expected invalid ban duration, got %v", err)
	}
	if err := service.Kick(context.Background(), 9, 1, 1); !errors.Is(err, ErrSelfTarget) {
		t.Fatalf("expected self target, got %v", err)
	}
}

// TestServiceEndsSanctionsAndReadsLists verifies reversal and read paths.
func TestServiceEndsSanctionsAndReadsLists(t *testing.T) {
	store := &moderationStoreForTest{mutes: map[int64]time.Time{2: time.Unix(2000, 0)}, bans: map[int64]time.Time{2: time.Unix(2000, 0)}}
	nodes := Nodes{OwnMute: "own.mute", OwnBan: "own.ban"}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	permissions := moderationPermissionsForTest{1: {nodes.OwnMute: true, nodes.OwnBan: true}}
	service := moderationServiceForTest(store, room, nil, permissions, nodes)
	if err := service.Unmute(context.Background(), 9, 1, 2); err != nil {
		t.Fatalf("unmute: %v", err)
	}
	if err := service.Unban(context.Background(), 9, 1, 2); err != nil {
		t.Fatalf("unban: %v", err)
	}
	if mutes, err := service.ListMutes(context.Background(), 9); err != nil || len(mutes) != 0 {
		t.Fatalf("list mutes %#v err=%v", mutes, err)
	}
	if bans, err := service.ListBans(context.Background(), 9); err != nil || len(bans) != 0 {
		t.Fatalf("list bans %#v err=%v", bans, err)
	}
}

// TestServiceAllowsGlobalStaffAndRejectsOwnerTarget verifies staff bypass limits.
func TestServiceAllowsGlobalStaffAndRejectsOwnerTarget(t *testing.T) {
	store := &moderationStoreForTest{mutes: make(map[int64]time.Time), bans: make(map[int64]time.Time)}
	nodes := Nodes{AnyKick: "any.kick"}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	service := moderationServiceForTest(store, room, nil, moderationPermissionsForTest{3: {nodes.AnyKick: true}}, nodes)
	if err := service.Kick(context.Background(), 9, 3, 2); err != nil {
		t.Fatalf("staff kick: %v", err)
	}
	if err := service.Kick(context.Background(), 9, 3, 1); !errors.Is(err, ErrTargetOwner) {
		t.Fatalf("expected owner target, got %v", err)
	}
}

// TestConfigLoadsAndNormalizesMuteLimits verifies environment and fallback behavior.
func TestConfigLoadsAndNormalizesMuteLimits(t *testing.T) {
	t.Setenv("PIXELS_ROOM_MODERATION_MIN_MUTE_MINUTES", "3")
	t.Setenv("PIXELS_ROOM_MODERATION_MAX_MUTE_MINUTES", "90")
	config, err := LoadConfig()
	if err != nil || config.MinMuteMinutes != 3 || config.MaxMuteMinutes != 90 {
		t.Fatalf("unexpected config %#v err=%v", config, err)
	}
	normalized := (Config{MinMuteMinutes: -1, MaxMuteMinutes: -1}).Normalize()
	if normalized.MinMuteMinutes != defaultMinMuteMinutes || normalized.MaxMuteMinutes != defaultMaxMuteMinutes {
		t.Fatalf("unexpected normalized config %#v", normalized)
	}
	service := moderationServiceForTest(&moderationStoreForTest{}, roommodel.Room{}, nil, nil, Nodes{})
	if _, err := service.ListBans(context.Background(), 0); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("expected invalid list identity, got %v", err)
	}
	if banned, err := service.IsBanned(context.Background(), 0, 2); err != nil || banned {
		t.Fatalf("expected invalid ban lookup miss, banned=%v err=%v", banned, err)
	}
}

// BenchmarkIsBanned measures active sanction resolution above repository I/O.
func BenchmarkIsBanned(b *testing.B) {
	store := &moderationStoreForTest{mutes: make(map[int64]time.Time), bans: map[int64]time.Time{2: time.Unix(2000, 0)}}
	service := moderationServiceForTest(store, roommodel.Room{}, nil, nil, Nodes{})
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		banned, err := service.IsBanned(ctx, 9, 2)
		if err != nil || !banned {
			b.Fatal("ban lookup failed")
		}
	}
}

// BenchmarkAuthorizeModerationAction measures owner-policy authorization overhead.
func BenchmarkAuthorizeModerationAction(b *testing.B) {
	store := &moderationStoreForTest{mutes: make(map[int64]time.Time), bans: make(map[int64]time.Time)}
	nodes := Nodes{OwnKick: "own.kick"}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	service := moderationServiceForTest(store, room, nil, moderationPermissionsForTest{1: {nodes.OwnKick: true}}, nodes)
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := service.authorize(ctx, 9, 1, 2, moderationmodel.ActionKick); err != nil {
			b.Fatal(err)
		}
	}
}
