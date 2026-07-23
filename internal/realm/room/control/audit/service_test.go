package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	auditmodel "github.com/niflaot/pixels/internal/realm/room/control/audit/model"
	bannedevent "github.com/niflaot/pixels/internal/realm/room/control/events/banned"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/control/events/kicked"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/control/events/muted"
	rightsgranted "github.com/niflaot/pixels/internal/realm/room/control/events/rightsgranted"
	rightsrevoked "github.com/niflaot/pixels/internal/realm/room/control/events/rightsrevoked"
	unbannedevent "github.com/niflaot/pixels/internal/realm/room/control/events/unbanned"
	unmutedevent "github.com/niflaot/pixels/internal/realm/room/control/events/unmuted"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// auditStoreForTest captures audit reads and writes.
type auditStoreForTest struct {
	// query stores the latest query.
	query StoreQuery
	// rights stores rights audit rows.
	rights []auditmodel.RightsAudit
	// moderation stores moderation rows.
	moderation []auditmodel.ModerationAction
}

// InsertRights appends one rights row.
func (store *auditStoreForTest) InsertRights(_ context.Context, entry auditmodel.RightsAudit) error {
	store.rights = append(store.rights, entry)
	return nil
}

// InsertModeration appends one moderation row.
func (store *auditStoreForTest) InsertModeration(_ context.Context, entry auditmodel.ModerationAction) error {
	store.moderation = append(store.moderation, entry)
	return nil
}

// RightsHistory returns configured rights rows.
func (store *auditStoreForTest) RightsHistory(_ context.Context, query StoreQuery) ([]auditmodel.RightsAudit, error) {
	store.query = query
	return store.rights, nil
}

// ModerationHistory returns configured moderation rows.
func (store *auditStoreForTest) ModerationHistory(_ context.Context, query StoreQuery) ([]auditmodel.ModerationAction, error) {
	store.query = query
	return store.moderation, nil
}

// TestServiceNormalizesAuditQueries verifies paging and typed action validation.
func TestServiceNormalizesAuditQueries(t *testing.T) {
	store := &auditStoreForTest{}
	service := New(store)
	roomID := int64(9)
	_, err := service.ModerationHistory(context.Background(), Query{RoomID: &roomID, ActionTypes: []moderationmodel.Action{moderationmodel.ActionKick}})
	if err != nil {
		t.Fatalf("moderation history: %v", err)
	}
	if store.query.Limit != defaultLimit || store.query.RoomID == nil || *store.query.RoomID != 9 {
		t.Fatalf("unexpected normalized query %#v", store.query)
	}
	_, err = service.ModerationHistory(context.Background(), Query{ActionTypes: []moderationmodel.Action{"invalid"}})
	if !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("expected invalid query, got %v", err)
	}
	invalidID := int64(0)
	_, err = service.RightsHistory(context.Background(), Query{RoomID: &invalidID})
	if !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("expected invalid room filter, got %v", err)
	}
	_, err = service.RightsHistory(context.Background(), Query{Limit: maxLimit + 10})
	if err != nil || store.query.Limit != maxLimit || len(store.query.ActionTypes) != 0 {
		t.Fatalf("unexpected capped query %#v err=%v", store.query, err)
	}
}

// TestRightsSubscriberPersistsActorAndTimestamp verifies event-to-audit projection.
func TestRightsSubscriberPersistsActorAndTimestamp(t *testing.T) {
	store := &auditStoreForTest{}
	at := time.Unix(1000, 0)
	handler := handleRightsGranted(store)
	err := handler(context.Background(), bus.Event{Name: rightsgranted.Name, At: at, Payload: rightsgranted.Payload{RoomID: 9, PlayerID: 2, ActorID: 1}})
	if err != nil {
		t.Fatalf("handle rights event: %v", err)
	}
	if len(store.rights) != 1 || store.rights[0].ActorID == nil || *store.rights[0].ActorID != 1 || !store.rights[0].CreatedAt.Equal(at) {
		t.Fatalf("unexpected rights audit %#v", store.rights)
	}
}

// TestRegisterSubscriberPersistsEveryEventFamily verifies complete bus wiring.
func TestRegisterSubscriberPersistsEveryEventFamily(t *testing.T) {
	store := &auditStoreForTest{}
	local := bus.New()
	lifecycle := fxtest.NewLifecycle(t)
	if err := RegisterSubscriber(lifecycle, local, store); err != nil {
		t.Fatalf("register subscriber: %v", err)
	}
	expiresAt := time.Unix(2000, 0)
	events := []bus.Event{
		{Name: rightsgranted.Name, Payload: rightsgranted.Payload{RoomID: 9, PlayerID: 2, ActorID: 1}},
		{Name: rightsrevoked.Name, Payload: rightsrevoked.Payload{RoomID: 9, PlayerID: 2, ActorID: 1, Action: rightsrevoked.ActionExplicit}},
		{Name: kickedevent.Name, Payload: kickedevent.Payload{RoomID: 9, TargetPlayerID: 2, ActorID: 1}},
		{Name: mutedevent.Name, Payload: mutedevent.Payload{RoomID: 9, TargetPlayerID: 2, ActorID: 1, DurationSeconds: 60, ExpiresAt: expiresAt}},
		{Name: unmutedevent.Name, Payload: unmutedevent.Payload{RoomID: 9, TargetPlayerID: 2, ActorID: 1}},
		{Name: bannedevent.Name, Payload: bannedevent.Payload{RoomID: 9, TargetPlayerID: 2, ActorID: 1, DurationSeconds: 3600, ExpiresAt: expiresAt}},
		{Name: unbannedevent.Name, Payload: unbannedevent.Payload{RoomID: 9, TargetPlayerID: 2, ActorID: 1}},
	}
	for _, event := range events {
		if err := local.Publish(context.Background(), event); err != nil {
			t.Fatalf("publish %s: %v", event.Name, err)
		}
	}
	if len(store.rights) != 2 || len(store.moderation) != 5 {
		t.Fatalf("unexpected audit counts rights=%d moderation=%d", len(store.rights), len(store.moderation))
	}
	lifecycle.RequireStop()
}

// BenchmarkModerationHistoryQuery measures service query normalization and dispatch.
func BenchmarkModerationHistoryQuery(b *testing.B) {
	store := &auditStoreForTest{moderation: make([]auditmodel.ModerationAction, 50)}
	service := New(store)
	query := Query{RoomID: pointerForBenchmark(9), Limit: 50}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := service.ModerationHistory(ctx, query); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAuditInsert measures one append-only subscriber projection.
func BenchmarkAuditInsert(b *testing.B) {
	store := &auditStoreForTest{rights: make([]auditmodel.RightsAudit, 0, 1)}
	handler := handleRightsGranted(store)
	event := bus.Event{Name: rightsgranted.Name, At: time.Unix(1000, 0), Payload: rightsgranted.Payload{RoomID: 9, PlayerID: 2, ActorID: 1}}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		store.rights = store.rights[:0]
		if err := handler(ctx, event); err != nil {
			b.Fatal(err)
		}
	}
}

// pointerForBenchmark returns a stable int64 pointer.
func pointerForBenchmark(value int64) *int64 {
	return &value
}
