package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// TestEligibleRoomsAgainstPostgres verifies creator rooms use the current room schema and binding state.
func TestEligibleRoomsAgainstPostgres(t *testing.T) {
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
	rooms, err := repository.EligibleRooms(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	creator, found := eligibleRoom(rooms, 130)
	if !found {
		t.Fatalf("unexpected creator room %#v found=%t", creator, found)
	}
	if _, found = eligibleRoom(rooms, 100); found {
		t.Fatal("bundle template must not be eligible")
	}
	if _, found = eligibleRoom(rooms, 131); found {
		t.Fatal("bound group room must not be eligible")
	}
	if err = repository.LockEligibleRoom(context.Background(), 130, 1); err != nil {
		t.Fatalf("lock creator room: %v", err)
	}
	if err = repository.LockEligibleRoom(context.Background(), 131, 1); !errors.Is(err, grouprecord.ErrConflict) {
		t.Fatalf("expected bound room conflict, got %v", err)
	}
}

// TestReturnHQFurnitureAgainstPostgres verifies the joined update has no ambiguous columns.
func TestReturnHQFurnitureAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("PIXELS_GROUP_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PIXELS_GROUP_TEST_DATABASE_URL is not configured")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	rows, err := tx.Query(context.Background(), returnHQFurnitureSQL, int64(2), int64(3), 1001)
	if err != nil {
		t.Fatalf("return headquarters furniture: %v", err)
	}
	rows.Close()
}

// TestPlayerGroupsWithoutFavoriteAgainstPostgres verifies absent preferences project as false.
func TestPlayerGroupsWithoutFavoriteAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("PIXELS_GROUP_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PIXELS_GROUP_TEST_DATABASE_URL is not configured")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	groups, err := New(pool).PlayerGroups(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) == 0 {
		t.Fatal("expected Bob's seeded memberships")
	}
	for _, group := range groups {
		if group.Favorite {
			t.Fatalf("unexpected favorite group: %#v", group)
		}
	}
}

// TestPopularGroupsAgainstPostgres verifies ranked group headquarters exclude ordinary rooms.
func TestPopularGroupsAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("PIXELS_GROUP_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PIXELS_GROUP_TEST_DATABASE_URL is not configured")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	groups, err := New(pool).PopularGroups(context.Background(), 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) == 0 {
		t.Fatal("expected seeded social groups")
	}
	for index, group := range groups {
		if !group.Active() || group.HomeRoomID <= 0 {
			t.Fatalf("unexpected group projection: %#v", group)
		}
		if index > 0 && groups[index-1].MemberCount < group.MemberCount {
			t.Fatalf("groups are not ranked by members: %#v", groups)
		}
	}
}

// eligibleRoom finds one room in a creator option projection.
func eligibleRoom(rooms []grouprecord.EligibleRoom, roomID int64) (grouprecord.EligibleRoom, bool) {
	for _, room := range rooms {
		if room.ID == roomID {
			return room, true
		}
	}

	return grouprecord.EligibleRoom{}, false
}
