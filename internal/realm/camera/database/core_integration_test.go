package database

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// TestCaptureReplacementAgainstPostgres verifies one reusable active photo and explicit replacement state.
func TestCaptureReplacementAgainstPostgres(t *testing.T) {
	repository, playerID, roomID := cameraRepositoryForTest(t)
	first, err := repository.CreateCapture(context.Background(), camerarecord.Capture{UUID: uuid.NewString(), PlayerID: playerID, RoomID: roomID, Kind: camerarecord.KindPhoto, StorageKey: "integration/first.png", URL: "https://storage/integration/first.png"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.CreateCapture(context.Background(), camerarecord.Capture{UUID: uuid.NewString(), PlayerID: playerID, RoomID: roomID, Kind: camerarecord.KindPhoto, StorageKey: "integration/second.png", URL: "https://storage/integration/second.png"})
	if err != nil {
		t.Fatal(err)
	}
	active, found, err := repository.ActiveCapture(context.Background(), playerID)
	if err != nil || !found || active.ID != second.ID || active.State != camerarecord.StatePending {
		t.Fatalf("unexpected active capture=%+v found=%t err=%v", active, found, err)
	}
	values, err := repository.Captures(context.Background(), playerID, 10)
	if err != nil || len(values) != 2 {
		t.Fatalf("unexpected captures=%+v err=%v", values, err)
	}
	if values[1].ID != first.ID || values[1].State != camerarecord.StateSuperseded || values[1].SupersededAt == nil {
		t.Fatalf("first capture was not superseded: %+v", values[1])
	}
}

// TestCooldownAgainstPostgres verifies durable publication throttling.
func TestCooldownAgainstPostgres(t *testing.T) {
	repository, playerID, _ := cameraRepositoryForTest(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	err := repository.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		if err := repository.SetPublishCooldown(txCtx, playerID, now); err != nil {
			return err
		}
		stored, found, err := repository.PublishCooldown(txCtx, playerID)
		if err != nil || !found || !stored.Equal(now) {
			return camerarecord.ErrCooldown
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestSettingsOptimisticLockAgainstPostgres verifies singleton version conflicts.
func TestSettingsOptimisticLockAgainstPostgres(t *testing.T) {
	repository, _, _ := cameraRepositoryForTest(t)
	settings, err := repository.Settings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("rollback integration settings")
	err = postgres.WithinScope(context.Background(), repository.pool, func(txCtx context.Context) error {
		updated, found, updateErr := repository.UpdateSettings(txCtx, settings, settings.Version)
		if updateErr != nil || !found || updated.Version != settings.Version+1 {
			return errors.New("settings update did not advance version")
		}
		_, found, updateErr = repository.UpdateSettings(txCtx, settings, settings.Version)
		if updateErr != nil || found {
			return errors.New("stale settings version was accepted")
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("unexpected rollback result: %v", err)
	}
}

// cameraRepositoryForTest creates isolated PostgreSQL player and room fixtures.
func cameraRepositoryForTest(t *testing.T) (*Repository, int64, int64) {
	t.Helper()
	dsn := os.Getenv("PIXELS_CAMERA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PIXELS_CAMERA_TEST_DATABASE_URL is not configured")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	username := "camera-test-" + uuid.NewString()
	var playerID, roomID int64
	if err = pool.QueryRow(context.Background(), `insert into players(username) values($1) returning id`, username).Scan(&playerID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(context.Background(), `insert into rooms(owner_player_id,owner_name,name,model_name) values($1,$2,'Camera Test Room','model_a') returning id`, playerID, username).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `delete from camera_publications where player_id=$1`, playerID)
		_, _ = pool.Exec(context.Background(), `delete from camera_publish_cooldowns where player_id=$1`, playerID)
		_, _ = pool.Exec(context.Background(), `delete from camera_captures where player_id=$1`, playerID)
		_, _ = pool.Exec(context.Background(), `delete from rooms where id=$1`, roomID)
		_, _ = pool.Exec(context.Background(), `delete from players where id=$1`, playerID)
	})
	return New(pool), playerID, roomID
}
