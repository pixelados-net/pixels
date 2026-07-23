package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// TestCreatePlayerScansRecord verifies player creation scans returned fields.
func TestCreatePlayerScansRecord(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	executor := &fakeExecutor{row: fakeRow{values: []any{
		int64(7),
		"ian",
		now,
		now,
		pgtype.Timestamptz{},
		int64(1),
		pgtype.Timestamptz{Time: now, Valid: true},
		pgtype.Timestamptz{},
		pgtype.Timestamptz{Time: now, Valid: true},
		int16(playermodel.ClubLevelVIP),
		pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
	}}}

	player, err := New(executor).CreatePlayer(context.Background(), CreatePlayerParams{Username: "ian"})
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	if player.ID != 7 {
		t.Fatalf("expected player id 7, got %d", player.ID)
	}

	if player.LastLoginAt == nil {
		t.Fatal("expected last login time")
	}
	if player.Club.Level != playermodel.ClubLevelVIP || player.Club.ExpiresAt == nil {
		t.Fatalf("expected active club record, got %#v", player.Club)
	}

	if !strings.Contains(executor.query, "insert into players") {
		t.Fatalf("expected create query, got %q", executor.query)
	}
}

// TestFindPlayerByIDNotFound verifies missing players are reported.
func TestFindPlayerByIDNotFound(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}

	_, found, err := New(executor).FindPlayerByID(context.Background(), 8)
	if err != nil {
		t.Fatalf("find player: %v", err)
	}

	if found {
		t.Fatal("expected missing player")
	}
}

// TestFindPlayerByUsernameWrapsScanError verifies scan failures are wrapped.
func TestFindPlayerByUsernameWrapsScanError(t *testing.T) {
	expected := errors.New("scan failed")
	executor := &fakeExecutor{row: fakeRow{err: expected}}

	_, _, err := New(executor).FindPlayerByUsername(context.Background(), "ian")
	if !errors.Is(err, expected) {
		t.Fatalf("expected scan error, got %v", err)
	}
}

// TestCreateProfileRejectsInvalidGender verifies profile gender validation.
func TestCreateProfileRejectsInvalidGender(t *testing.T) {
	_, err := New(&fakeExecutor{}).CreateProfile(context.Background(), CreateProfileParams{Gender: playermodel.Gender("X")})
	if !errors.Is(err, ErrInvalidGender) {
		t.Fatalf("expected invalid gender, got %v", err)
	}
}

// TestCreateProfileScansRecord verifies profile creation scans returned fields.
func TestCreateProfileScansRecord(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	homeRoomID := int64(99)
	executor := &fakeExecutor{row: fakeRow{values: []any{
		int64(7),
		"hd-180-1",
		"M",
		"hello",
		pgtype.Int8{Int64: homeRoomID, Valid: true},
		true,
		int32(3),
		false,
		false,
		false,
		now,
		now,
		int64(2),
	}}}

	profile, err := New(executor).CreateProfile(context.Background(), CreateProfileParams{
		PlayerID:        7,
		Look:            "hd-180-1",
		Gender:          playermodel.GenderMale,
		Motto:           "hello",
		HomeRoomID:      &homeRoomID,
		AllowNameChange: true,
	})
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}

	if profile.HomeRoomID == nil || *profile.HomeRoomID != homeRoomID {
		t.Fatal("expected home room id")
	}
}

// TestFindProfileByPlayerIDNotFound verifies missing profiles are reported.
func TestFindProfileByPlayerIDNotFound(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}

	_, found, err := New(executor).FindProfileByPlayerID(context.Background(), 8)
	if err != nil {
		t.Fatalf("find profile: %v", err)
	}

	if found {
		t.Fatal("expected missing profile")
	}
}

// TestUpdatePlayerNotFound verifies optimistic identity conflicts are reported.
func TestUpdatePlayerNotFound(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}
	_, matched, err := New(executor).UpdatePlayer(context.Background(), UpdatePlayerParams{PlayerID: 7, Username: "ian", ExpectedVersion: 1})
	if err != nil || matched {
		t.Fatalf("expected unmatched update, matched=%t err=%v", matched, err)
	}
}

// TestUpdateProfileRejectsInvalidGender verifies administrative profile validation.
func TestUpdateProfileRejectsInvalidGender(t *testing.T) {
	_, _, err := New(&fakeExecutor{}).UpdateProfile(context.Background(), UpdateProfileParams{CreateProfileParams: CreateProfileParams{Gender: "X"}})
	if !errors.Is(err, ErrInvalidGender) {
		t.Fatalf("expected invalid gender, got %v", err)
	}
}

// TestSoftDeletePlayerReportsAffectedRow verifies durable deletion success.
func TestSoftDeletePlayerReportsAffectedRow(t *testing.T) {
	executor := &fakeExecutor{commandTag: pgconn.NewCommandTag("UPDATE 1")}
	deleted, err := New(executor).SoftDeletePlayer(context.Background(), 7, 1)
	if err != nil || !deleted {
		t.Fatalf("expected deletion, deleted=%t err=%v", deleted, err)
	}
}

// fakeExecutor records repository query calls for tests.
type fakeExecutor struct {
	// row is the row returned by QueryRow.
	row pgx.Row

	// query is the last executed query.
	query string

	// commandTag is returned by Exec.
	commandTag pgconn.CommandTag
}

// Exec executes SQL without returning rows.
func (executor *fakeExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return executor.commandTag, nil
}

// Query executes SQL returning multiple rows.
func (executor *fakeExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

// QueryRow executes SQL returning one row.
func (executor *fakeExecutor) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	executor.query = query

	return executor.row
}

// fakeRow scans fixed values for tests.
type fakeRow struct {
	// values are copied into scan destinations.
	values []any

	// err is returned by Scan before copying values.
	err error
}

// Scan copies values into destinations.
func (row fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}

	for index, value := range row.values {
		assign(destinations[index], value)
	}

	return nil
}

// assign copies one value into one destination.
func assign(destination any, value any) {
	switch target := destination.(type) {
	case *int64:
		*target = value.(int64)
	case *int16:
		*target = value.(int16)
	case *int32:
		*target = value.(int32)
	case *string:
		*target = value.(string)
	case *bool:
		*target = value.(bool)
	case *time.Time:
		*target = value.(time.Time)
	case *pgtype.Timestamptz:
		*target = value.(pgtype.Timestamptz)
	case *pgtype.Int8:
		*target = value.(pgtype.Int8)
	}
}
