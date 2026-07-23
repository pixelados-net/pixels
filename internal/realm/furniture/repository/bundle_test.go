package repository

import (
	"context"
	"strings"
	"testing"
)

// TestCloneRoomItemsUsesOneResettingStatement verifies efficient safe cloning.
func TestCloneRoomItemsUsesOneResettingStatement(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: []any{7}}}
	count, err := New(executor).CloneRoomItems(context.Background(), 100, 44, 7)
	if err != nil || count != 7 {
		t.Fatalf("count=%d error=%v", count, err)
	}
	checks := []string{"insert into furniture_items", "select definition_id, $3, $2", "extra_data, null, false, false", "gift_sender_player_id", "metadata"}
	for _, check := range checks {
		if !strings.Contains(executor.query, check) {
			t.Fatalf("clone query does not contain %q: %s", check, executor.query)
		}
	}
	if len(executor.arguments) != 3 {
		t.Fatalf("arguments=%#v", executor.arguments)
	}
}

// TestListRoomBundleProductsUsesDatabaseGrouping verifies previews avoid per-item loading.
func TestListRoomBundleProductsUsesDatabaseGrouping(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{{int64(3), int32(2)}, {int64(26), int32(1)}}}}
	products, err := New(executor).ListRoomBundleProducts(context.Background(), 100)
	if err != nil || len(products) != 2 || products[0].Quantity != 2 {
		t.Fatalf("products=%#v error=%v", products, err)
	}
	if !strings.Contains(executor.query, "group by definition_id") {
		t.Fatalf("query=%s", executor.query)
	}
}
