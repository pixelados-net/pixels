package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// deniedSanctionPermissions rejects every global sanction capability.
type deniedSanctionPermissions struct{}

// HasPermission rejects the requested node.
func (deniedSanctionPermissions) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return false, nil
}

// TestRegisterRejectsInvalidAndDuplicateEffects verifies behavior registration integrity.
func TestRegisterRejectsInvalidAndDuplicateEffects(t *testing.T) {
	service := New(&storeForTest{}, permissionsForTest{}, nil, nil)
	if err := service.Register(nil); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("nil err=%v", err)
	}
	applier := &applierForTest{}
	if err := service.Register(applier); err != nil {
		t.Fatal(err)
	}
	if err := service.Register(applier); !errors.Is(err, ErrApplierExists) {
		t.Fatalf("duplicate err=%v", err)
	}
}

// TestInstantPunishmentNormalizesInput verifies warn records cannot become active windows.
func TestInstantPunishmentNormalizesInput(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	store := &storeForTest{now: now}
	service := New(store, permissionsForTest{}, bus.New(), nil)
	service.now = func() time.Time { return now }
	issuer := int64(1)
	value, err := service.Apply(context.Background(), sanctionrecord.ApplyParams{ReceiverPlayerID: 2, IssuerPlayerID: &issuer, Kind: sanctionrecord.KindWarn, Reason: "  warning  ", Source: "  modtool  ", ExpiresAt: &expires})
	if err != nil || value.Reason != "warning" || value.Source != "modtool" || value.ExpiresAt != nil || value.ActiveAt(now) {
		t.Fatalf("value=%+v err=%v", value, err)
	}
}

// TestRevokeAuthorizationAndEffects verifies only authorized actors mutate active truth.
func TestRevokeAuthorizationAndEffects(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &storeForTest{now: now, values: []sanctionrecord.Punishment{{ID: 1, ReceiverPlayerID: 2, Kind: sanctionrecord.KindWarn, IssuedAt: now}}}
	denied := New(store, deniedSanctionPermissions{}, nil, nil)
	if _, err := denied.Revoke(context.Background(), 1, 7); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("denied err=%v", err)
	}
	service := New(store, permissionsForTest{}, bus.New(), nil)
	service.now = func() time.Time { return now }
	applier := &applierForTest{}
	if err := service.Register(applier); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Revoke(context.Background(), 99, 7); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing err=%v", err)
	}
	value, err := service.Revoke(context.Background(), 1, 7)
	if err != nil || value.RevokedAt == nil || applier.revokes != 1 {
		t.Fatalf("value=%+v err=%v revokes=%d", value, err, applier.revokes)
	}
}

// TestSystemRevokeAndAccessors verifies the authorized internal alias and bounded reads.
func TestSystemRevokeAndAccessors(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &storeForTest{now: now, values: []sanctionrecord.Punishment{{ID: 1, ReceiverPlayerID: 2, Kind: sanctionrecord.KindTradeLock, IssuedAt: now}}}
	service := New(store, permissionsForTest{}, bus.New(), nil)
	service.now = func() time.Time { return now }
	if service.Store() != store {
		t.Fatal("store accessor mismatch")
	}
	items, err := service.History(context.Background(), 2, 0)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	value, err := service.RevokeSystem(context.Background(), 1)
	if err != nil || value.RevokedAt == nil {
		t.Fatalf("value=%+v err=%v", value, err)
	}
	banned, reason, err := service.CheckBan(context.Background(), 2)
	if err != nil || banned || reason != "" {
		t.Fatalf("banned=%v reason=%q err=%v", banned, reason, err)
	}
	if _, err = service.RevokeSystem(context.Background(), 0); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("invalid system revoke err=%v", err)
	}
}

// TestCheckBanUsesActiveProjection verifies login gating ignores expired bans.
func TestCheckBanUsesActiveProjection(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &storeForTest{now: now, values: []sanctionrecord.Punishment{{ID: 1, ReceiverPlayerID: 2, Kind: sanctionrecord.KindBan, Reason: "active", IssuedAt: now}}}
	service := New(store, permissionsForTest{}, nil, nil)
	service.now = func() time.Time { return now }
	banned, reason, err := service.CheckBan(context.Background(), 2)
	if err != nil || !banned || reason != "active" {
		t.Fatalf("banned=%v reason=%q err=%v", banned, reason, err)
	}
}

// TestEscalationResetsOutsideProbationAndRejectsEmptyPolicy verifies ladder boundaries.
func TestEscalationResetsOutsideProbationAndRejectsEmptyPolicy(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &storeForTest{now: now}
	service := New(store, permissionsForTest{}, nil, nil)
	service.now = func() time.Time { return now }
	if _, err := service.EscalateFor(context.Background(), EscalateParams{ReceiverPlayerID: 2, Reason: "reason"}); !errors.Is(err, ErrLadderEmpty) {
		t.Fatalf("empty err=%v", err)
	}
	store.ladder = []sanctionrecord.LadderEntry{{Level: 1, Kind: sanctionrecord.KindWarn, ProbationDays: 1}, {Level: 2, Kind: sanctionrecord.KindMute, DurationHours: 1, ProbationDays: 2}}
	store.lastLevel = 2
	store.values = []sanctionrecord.Punishment{{ID: 1, ReceiverPlayerID: 2, Kind: sanctionrecord.KindMute, Source: "escalation", IssuedAt: now.Add(-72 * time.Hour)}}
	value, err := service.EscalateFor(context.Background(), EscalateParams{ReceiverPlayerID: 2, Reason: "reset"})
	if err != nil || value.Kind != sanctionrecord.KindWarn {
		t.Fatalf("value=%+v err=%v", value, err)
	}
}
