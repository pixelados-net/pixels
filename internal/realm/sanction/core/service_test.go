package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// storeForTest stores enough sanction state for focused engine behavior.
type storeForTest struct {
	values    []sanctionrecord.Punishment
	ladder    []sanctionrecord.LadderEntry
	lastLevel int32
	now       time.Time
}

// Insert records one punishment.
func (store *storeForTest) Insert(_ context.Context, params sanctionrecord.ApplyParams) (sanctionrecord.Punishment, error) {
	value := sanctionrecord.Punishment{ID: int64(len(store.values) + 1), ReceiverPlayerID: params.ReceiverPlayerID, IssuerPlayerID: params.IssuerPlayerID, IssuerKind: params.IssuerKind, Kind: params.Kind, Reason: params.Reason, Source: params.Source, IssuedAt: store.now, ExpiresAt: params.ExpiresAt}
	store.values = append(store.values, value)
	return value, nil
}

// Find returns one punishment.
func (store *storeForTest) Find(_ context.Context, id int64) (sanctionrecord.Punishment, bool, error) {
	for _, value := range store.values {
		if value.ID == id {
			return value, true, nil
		}
	}
	return sanctionrecord.Punishment{}, false, nil
}

// List returns detached punishment history.
func (store *storeForTest) List(context.Context, int64, int32) ([]sanctionrecord.Punishment, error) {
	return append([]sanctionrecord.Punishment(nil), store.values...), nil
}

// Active derives active state for tests.
func (store *storeForTest) Active(_ context.Context, playerID int64, now time.Time) (sanctionrecord.ActiveState, error) {
	var state sanctionrecord.ActiveState
	for index := range store.values {
		value := &store.values[index]
		if value.ReceiverPlayerID != playerID || !value.ActiveAt(now) {
			continue
		}
		switch value.Kind {
		case sanctionrecord.KindBan:
			state.Ban = value
		case sanctionrecord.KindMute:
			state.MutedPermanently = value.ExpiresAt == nil
			state.MuteUntil = value.ExpiresAt
		case sanctionrecord.KindTradeLock:
			state.TradeLockedPermanently = value.ExpiresAt == nil
			state.TradeLockUntil = value.ExpiresAt
		}
	}
	return state, nil
}

// Revoke marks one punishment revoked.
func (store *storeForTest) Revoke(_ context.Context, id int64, actorID *int64, now time.Time) (sanctionrecord.Punishment, bool, error) {
	for index := range store.values {
		if store.values[index].ID == id && store.values[index].RevokedAt == nil {
			store.values[index].RevokedAt, store.values[index].RevokedByPlayerID = &now, actorID
			return store.values[index], true, nil
		}
	}
	return sanctionrecord.Punishment{}, false, nil
}

// LastEscalation returns configured test history.
func (store *storeForTest) LastEscalation(context.Context, int64) (sanctionrecord.Punishment, int32, bool, error) {
	if len(store.values) == 0 {
		return sanctionrecord.Punishment{}, 0, false, nil
	}
	return store.values[len(store.values)-1], store.lastLevel, true, nil
}

// Ladder returns configured escalation policy.
func (store *storeForTest) Ladder(context.Context) ([]sanctionrecord.LadderEntry, error) {
	return store.ladder, nil
}

// ReplaceLadder replaces escalation policy.
func (store *storeForTest) ReplaceLadder(_ context.Context, values []sanctionrecord.LadderEntry) error {
	store.ladder = values
	return nil
}

// QueueAlert accepts pending alerts.
func (*storeForTest) QueueAlert(context.Context, sanctionrecord.Alert) error { return nil }

// PendingAlerts returns no test alerts.
func (*storeForTest) PendingAlerts(context.Context, int64, int32) ([]sanctionrecord.Alert, error) {
	return nil, nil
}

// MarkAlertDelivered accepts test acknowledgements.
func (*storeForTest) MarkAlertDelivered(context.Context, int64, time.Time) error { return nil }

// permissionsForTest permits issuers and optionally protects targets.
type permissionsForTest struct{ immune bool }

// HasPermission resolves focused test nodes.
func (permissions permissionsForTest) HasPermission(_ context.Context, playerID int64, node permission.Node) (bool, error) {
	if node == ImmuneNode && playerID == 2 {
		return permissions.immune, nil
	}
	return true, nil
}

// applierForTest records calls and returns a configured error.
type applierForTest struct {
	applies int
	revokes int
	err     error
}

// Kind identifies test warnings.
func (*applierForTest) Kind() sanctionrecord.Kind { return sanctionrecord.KindWarn }

// Apply records one side effect.
func (applier *applierForTest) Apply(context.Context, sanctionrecord.Punishment) error {
	applier.applies++
	return applier.err
}

// Revoke records one side effect.
func (applier *applierForTest) Revoke(context.Context, sanctionrecord.Punishment) error {
	applier.revokes++
	return applier.err
}

// TestApplyPersistsWhenEffectFails verifies durable truth wins over side effects.
func TestApplyPersistsWhenEffectFails(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &storeForTest{now: now}
	service := New(store, permissionsForTest{}, nil, nil)
	service.now = func() time.Time { return now }
	applier := &applierForTest{err: errors.New("delivery failed")}
	if err := service.Register(applier); err != nil {
		t.Fatal(err)
	}
	issuer := int64(1)
	value, err := service.Apply(context.Background(), sanctionrecord.ApplyParams{ReceiverPlayerID: 2, IssuerPlayerID: &issuer, Kind: sanctionrecord.KindWarn, Reason: "warning", Source: "modtool"})
	if err != nil || value.ID == 0 || len(store.values) != 1 || applier.applies != 1 {
		t.Fatalf("value=%+v err=%v stored=%d applies=%d", value, err, len(store.values), applier.applies)
	}
}

// TestApplyRejectsImmuneTarget verifies authority before persistence.
func TestApplyRejectsImmuneTarget(t *testing.T) {
	store := &storeForTest{now: time.Now()}
	service := New(store, permissionsForTest{immune: true}, nil, nil)
	issuer := int64(1)
	_, err := service.Apply(context.Background(), sanctionrecord.ApplyParams{ReceiverPlayerID: 2, IssuerPlayerID: &issuer, Kind: sanctionrecord.KindMute, Reason: "reason", Source: "modtool"})
	if !errors.Is(err, ErrImmune) || len(store.values) != 0 {
		t.Fatalf("err=%v stored=%d", err, len(store.values))
	}
}

// TestSystemApplyRejectsImmuneTarget verifies automated escalation honors immunity too.
func TestSystemApplyRejectsImmuneTarget(t *testing.T) {
	store := &storeForTest{now: time.Now()}
	service := New(store, permissionsForTest{immune: true}, nil, nil)
	_, err := service.Apply(context.Background(), sanctionrecord.ApplyParams{ReceiverPlayerID: 2, IssuerKind: "system", Kind: sanctionrecord.KindWarn, Reason: "reason", Source: "escalation"})
	if !errors.Is(err, ErrImmune) || len(store.values) != 0 {
		t.Fatalf("err=%v stored=%d", err, len(store.values))
	}
}

// TestApplyRejectsInvalidExpiry verifies finite sanctions cannot overflow or arrive expired.
func TestApplyRejectsInvalidExpiry(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &storeForTest{now: now}
	service := New(store, permissionsForTest{}, nil, nil)
	service.now = func() time.Time { return now }
	issuer := int64(1)
	for _, expiry := range []time.Time{now, now.Add(time.Duration(sanctionrecord.MaxDurationHours+1) * time.Hour)} {
		_, err := service.Apply(context.Background(), sanctionrecord.ApplyParams{ReceiverPlayerID: 2, IssuerPlayerID: &issuer, Kind: sanctionrecord.KindMute, Reason: "reason", Source: "modtool", ExpiresAt: &expiry})
		if !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("expiry=%v err=%v", expiry, err)
		}
	}
}

// TestActiveStateIsTimestampDerived verifies expired sanctions need no cleanup worker.
func TestActiveStateIsTimestampDerived(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	expired := now.Add(-time.Second)
	store := &storeForTest{now: now, values: []sanctionrecord.Punishment{{ID: 1, ReceiverPlayerID: 2, Kind: sanctionrecord.KindMute, IssuedAt: now.Add(-time.Hour), ExpiresAt: &expired}}}
	service := New(store, permissionsForTest{}, nil, nil)
	service.now = func() time.Time { return now }
	state, err := service.Active(context.Background(), 2)
	if err != nil || state.MutedPermanently || state.MuteUntil != nil {
		t.Fatalf("state=%+v err=%v", state, err)
	}
}

// TestEscalateWithinProbationAdvancesLevel verifies timestamp-only escalation.
func TestEscalateWithinProbationAdvancesLevel(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	store := &storeForTest{now: now, lastLevel: 1, ladder: []sanctionrecord.LadderEntry{{Level: 1, Kind: sanctionrecord.KindWarn, ProbationDays: 2}, {Level: 2, Kind: sanctionrecord.KindMute, DurationHours: 4, ProbationDays: 7}}, values: []sanctionrecord.Punishment{{ID: 1, ReceiverPlayerID: 2, Kind: sanctionrecord.KindWarn, Source: "escalation", IssuedAt: now.Add(-time.Hour)}}}
	service := New(store, permissionsForTest{}, nil, nil)
	service.now = func() time.Time { return now }
	value, err := service.EscalateFor(context.Background(), EscalateParams{ReceiverPlayerID: 2, Reason: "repeat"})
	if err != nil || value.Kind != sanctionrecord.KindMute || value.ExpiresAt == nil || value.ExpiresAt.Sub(now) != 4*time.Hour {
		t.Fatalf("value=%+v err=%v", value, err)
	}
}

// BenchmarkActiveSanctionLookup measures the indexed-login service boundary without transport allocations.
func BenchmarkActiveSanctionLookup(b *testing.B) {
	now := time.Now()
	expires := now.Add(time.Hour)
	store := &storeForTest{now: now, values: []sanctionrecord.Punishment{{ID: 1, ReceiverPlayerID: 2, Kind: sanctionrecord.KindMute, IssuedAt: now, ExpiresAt: &expires}}}
	service := New(store, permissionsForTest{}, nil, nil)
	service.now = func() time.Time { return now }
	b.ReportAllocs()
	for range b.N {
		if _, err := service.Active(context.Background(), 2); err != nil {
			b.Fatal(err)
		}
	}
}
