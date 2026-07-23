package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
)

// TestGrantPersistenceReadsCollections verifies grant and affected player scans.
func TestGrantPersistenceReadsCollections(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{{"catalog.admin.manage", true}}}}
	repository := New(executor)
	grants, err := repository.ListPlayerNodes(context.Background(), 3)
	if err != nil || len(grants) != 1 || grants[0].Node != permission.Node("catalog.admin.manage") || !grants[0].Allowed {
		t.Fatalf("unexpected grants=%#v err=%v", grants, err)
	}
	executor.rows = &fakeRows{values: [][]any{{"catalog.*", false}}}
	grants, err = repository.ListGroupNodes(context.Background(), 2)
	if err != nil || len(grants) != 1 || grants[0].Allowed {
		t.Fatalf("unexpected group grants=%#v err=%v", grants, err)
	}

	executor.rows = &fakeRows{values: [][]any{{int64(3)}, {int64(5)}}}
	players, err := repository.ListAffectedPlayerIDs(context.Background(), 2)
	if err != nil || len(players) != 2 || players[1] != 5 {
		t.Fatalf("unexpected players=%#v err=%v", players, err)
	}
}

// TestGrantPersistenceMutationsUseExpectedArguments verifies every grant mutation.
func TestGrantPersistenceMutationsUseExpectedArguments(t *testing.T) {
	executor := &fakeExecutor{}
	repository := New(executor)
	operations := []struct {
		name string
		run  func() error
		want int
	}{
		{name: "upsert group node", run: func() error { return repository.UpsertGroupNode(context.Background(), 2, "catalog.*", true) }, want: 3},
		{name: "delete group node", run: func() error { return repository.DeleteGroupNode(context.Background(), 2, "catalog.*") }, want: 2},
		{name: "add membership", run: func() error { return repository.AddPlayerToGroup(context.Background(), 3, 2) }, want: 2},
		{name: "remove membership", run: func() error { return repository.RemovePlayerFromGroup(context.Background(), 3, 2) }, want: 2},
		{name: "upsert player node", run: func() error { return repository.UpsertPlayerNode(context.Background(), 3, "catalog.*", false) }, want: 3},
		{name: "delete player node", run: func() error { return repository.DeletePlayerNode(context.Background(), 3, "catalog.*") }, want: 2},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			if err := operation.run(); err != nil || len(executor.arguments) != operation.want {
				t.Fatalf("unexpected args=%#v err=%v", executor.arguments, err)
			}
		})
	}
}

// TestGrantPersistenceWrapsMutationFailures verifies meaningful mutation errors.
func TestGrantPersistenceWrapsMutationFailures(t *testing.T) {
	failure := errors.New("constraint failed")
	err := New(&fakeExecutor{err: failure}).AddPlayerToGroup(context.Background(), 3, 2)
	if !errors.Is(err, failure) || !strings.Contains(err.Error(), "mutate permission persistence") {
		t.Fatalf("expected wrapped mutation failure, got %v", err)
	}
}

// BenchmarkPermissionGrantMutation measures repository command preparation.
func BenchmarkPermissionGrantMutation(b *testing.B) {
	repository := New(&fakeExecutor{})
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		if err := repository.UpsertGroupNode(ctx, 2, "catalog.admin.manage", true); err != nil {
			b.Fatalf("upsert group node: %v", err)
		}
	}
}
