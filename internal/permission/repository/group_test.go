package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestGroupPersistenceReadsAndCreates verifies group projections and inserts.
func TestGroupPersistenceReadsAndCreates(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{groupValues()}}}
	repository := New(executor)
	groups, err := repository.ListGroupsByPlayer(context.Background(), 4)
	if err != nil || len(groups) != 1 || groups[0].Name != "admin" || groups[0].ParentGroupID == nil {
		t.Fatalf("unexpected groups=%#v err=%v", groups, err)
	}
	if !strings.Contains(executor.query, "g.created_at") {
		t.Fatalf("expected qualified joined projection, got %s", executor.query)
	}
	executor.rows = &fakeRows{values: [][]any{groupValues()}}
	groups, err = repository.ListGroups(context.Background())
	if err != nil || len(groups) != 1 {
		t.Fatalf("unexpected complete groups=%#v err=%v", groups, err)
	}

	executor.row = fakeRow{values: groupValues()}
	created, err := repository.CreateGroup(context.Background(), permissionmodel.Group{Name: "admin", Weight: 100})
	if err != nil || created.ID != 2 || len(executor.arguments) != 7 {
		t.Fatalf("unexpected created=%#v args=%#v err=%v", created, executor.arguments, err)
	}
	executor.row = fakeRow{values: groupValues()}
	foundGroup, found, err := repository.FindGroupByName(context.Background(), "admin")
	if err != nil || !found || foundGroup.ID != 2 {
		t.Fatalf("unexpected found group=%#v found=%v err=%v", foundGroup, found, err)
	}
}

// TestGroupPersistenceHandlesOptionalAndOptimisticResults verifies row misses.
func TestGroupPersistenceHandlesOptionalAndOptimisticResults(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}
	repository := New(executor)
	_, found, err := repository.FindGroupByID(context.Background(), 99)
	if err != nil || found {
		t.Fatalf("expected missing group, found=%v err=%v", found, err)
	}

	group := permissionmodel.Group{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}, Version: sharedmodel.Version{Version: 3}}}
	_, changed, err := repository.UpdateGroup(context.Background(), group)
	if err != nil || changed {
		t.Fatalf("expected optimistic miss, changed=%v err=%v", changed, err)
	}
}

// TestGroupPersistenceWrapsFailures verifies actionable repository errors.
func TestGroupPersistenceWrapsFailures(t *testing.T) {
	failure := errors.New("database unavailable")
	executor := &fakeExecutor{err: failure}
	_, err := New(executor).ListGroups(context.Background())
	if !errors.Is(err, failure) || !strings.Contains(err.Error(), "query permission groups") {
		t.Fatalf("expected wrapped query failure, got %v", err)
	}
}

// BenchmarkScanPermissionGroup measures durable group row projection.
func BenchmarkScanPermissionGroup(b *testing.B) {
	row := fakeRow{values: groupValues()}
	b.ReportAllocs()
	for b.Loop() {
		group, err := scanGroup(row)
		if err != nil || group.ID != 2 {
			b.Fatalf("unexpected group=%#v err=%v", group, err)
		}
	}
}
