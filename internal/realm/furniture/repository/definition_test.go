package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// TestFindDefinitionByIDScansReturnedRecord verifies definition lookup scanning.
func TestFindDefinitionByIDScansReturnedRecord(t *testing.T) {
	definition, found, err := New(&fakeExecutor{row: fakeRow{values: definitionValuesForTest()}}).FindDefinitionByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("find definition: %v", err)
	}
	if !found || definition.ID != 2 || definition.Name != "chair_plasto" || definition.Kind != furnituremodel.KindFloor {
		t.Fatalf("unexpected definition=%#v found=%v", definition, found)
	}
	if definition.StackHeight != 1.0 || !definition.AllowSit || definition.AllowStack {
		t.Fatalf("unexpected definition flags %#v", definition)
	}
}

// TestFindDefinitionByIDReportsMissing verifies missing definition lookup.
func TestFindDefinitionByIDReportsMissing(t *testing.T) {
	_, found, err := New(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).FindDefinitionByID(context.Background(), 2)
	if err != nil || found {
		t.Fatalf("expected missing definition, found=%v err=%v", found, err)
	}
}

// TestDefinitionQueriesPropagateErrors verifies definition query failures.
func TestDefinitionQueriesPropagateErrors(t *testing.T) {
	expected := errors.New("definition query failed")
	_, _, err := New(&fakeExecutor{row: fakeRow{err: expected}}).FindDefinitionByID(context.Background(), 2)
	if !errors.Is(err, expected) {
		t.Fatalf("expected find error, got %v", err)
	}
	_, err = New(&fakeExecutor{rows: &fakeRows{err: expected}}).ListDefinitions(context.Background())
	if !errors.Is(err, expected) {
		t.Fatalf("expected list error, got %v", err)
	}
}

// TestListDefinitionsScansRows verifies definition list scanning.
func TestListDefinitionsScansRows(t *testing.T) {
	definitions, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{definitionValuesForTest()}}}).ListDefinitions(context.Background())
	if err != nil || len(definitions) != 1 || definitions[0].Name != "chair_plasto" {
		t.Fatalf("unexpected definitions=%#v err=%v", definitions, err)
	}
}
