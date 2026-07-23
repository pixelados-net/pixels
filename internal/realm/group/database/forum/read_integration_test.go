package forum

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestForumReadsAgainstPostgres verifies authorization SQL and placeholder typing.
func TestForumReadsAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("PIXELS_GROUP_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PIXELS_GROUP_TEST_DATABASE_URL is not configured")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repository := New(pool)
	items, total, err := repository.ForumSummaries(context.Background(), 1, 0, 0, 50, false, time.Now().Add(-7*24*time.Hour))
	if err != nil || total < 1 || len(items) < 1 || items[0].Group.ID != 5 {
		t.Fatalf("items=%#v total=%d err=%v", items, total, err)
	}
	items, total, err = repository.ForumSummaries(context.Background(), 1, 2, 0, 50, false, time.Now().Add(-7*24*time.Hour))
	if err != nil || total < 1 || len(items) < 1 {
		t.Fatalf("member items=%#v total=%d err=%v", items, total, err)
	}
	if _, err = repository.UnreadCount(context.Background(), 1, false); err != nil {
		t.Fatal(err)
	}
}
