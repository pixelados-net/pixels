package routes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	redispkg "github.com/niflaot/pixels/pkg/redis"
)

// TestCreateReplaysIdempotentResult verifies retries do not create duplicate players.
func TestCreateReplaysIdempotentResult(t *testing.T) {
	app, manager := testApplication(t)
	first := createRequest(t, app, "registration-123456", `{"username":"pixel","look":"hd-180-1","gender":"M"}`)
	second := createRequest(t, app, "registration-123456", `{"username":"pixel","look":"hd-180-1","gender":"M"}`)

	if first.StatusCode != fiber.StatusCreated || second.StatusCode != fiber.StatusOK {
		t.Fatalf("expected create and replay statuses, got %d and %d", first.StatusCode, second.StatusCode)
	}
	if manager.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", manager.createCalls)
	}
	if second.Header.Get(replayHeader) != "true" {
		t.Fatal("expected replay response header")
	}
}

// TestCreateRejectsChangedIdempotentPayload verifies keys cannot be reused for other input.
func TestCreateRejectsChangedIdempotentPayload(t *testing.T) {
	app, _ := testApplication(t)
	_ = createRequest(t, app, "registration-123456", `{"username":"pixel","look":"hd-180-1","gender":"M"}`)
	response := createRequest(t, app, "registration-123456", `{"username":"other","look":"hd-180-1","gender":"M"}`)

	if response.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected conflict, got %d", response.StatusCode)
	}
}

// TestCreateReleasesFailedIdempotencyClaim verifies a corrected retry can proceed.
func TestCreateReleasesFailedIdempotencyClaim(t *testing.T) {
	app, manager := testApplication(t)
	manager.operationErr = playerservice.ErrInvalidUsername
	failed := createRequest(t, app, "registration-retry", `{"username":"pixel","look":"hd-180-1","gender":"M"}`)
	if failed.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", failed.StatusCode)
	}

	manager.operationErr = nil
	retried := createRequest(t, app, "registration-retry", `{"username":"pixel","look":"hd-180-1","gender":"M"}`)
	if retried.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected retry creation, got %d", retried.StatusCode)
	}
}

// TestReadUsesETag verifies unchanged player reads return not modified.
func TestReadUsesETag(t *testing.T) {
	app, _ := testApplication(t)
	request, err := http.NewRequest(http.MethodGet, "/api/admin/players/7", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("read player: %v", err)
	}

	conditional, err := http.NewRequest(http.MethodGet, "/api/admin/players/7", nil)
	if err != nil {
		t.Fatalf("new conditional request: %v", err)
	}
	conditional.Header.Set(fiber.HeaderIfNoneMatch, response.Header.Get(fiber.HeaderETag))
	unchanged, err := app.Test(conditional)
	if err != nil {
		t.Fatalf("conditional read: %v", err)
	}
	if unchanged.StatusCode != fiber.StatusNotModified {
		t.Fatalf("expected not modified, got %d", unchanged.StatusCode)
	}
}

// testApplication creates player routes with deterministic dependencies.
func testApplication(t *testing.T) (*fiber.App, *fakeManager) {
	t.Helper()
	server := miniredis.RunT(t)
	redisClient := redispkg.New(redispkg.Config{Address: server.Addr()})
	t.Cleanup(func() {
		if err := redisClient.Close(); err != nil {
			t.Fatalf("close redis: %v", err)
		}
	})
	manager := &fakeManager{record: testRecord()}
	app := fiber.New()
	Register(app, manager, redisClient, nil, nil)

	return app, manager
}

// createRequest submits one player creation request.
func createRequest(t *testing.T, app *fiber.App, key string, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, "/api/admin/players", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set(idempotencyHeader, key)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	_, _ = io.Copy(io.Discard, response.Body)

	return response
}

// testRecord returns one player representation fixture.
func testRecord() playerservice.Record {
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	return playerservice.Record{
		Player: playermodel.Player{Base: sharedmodel.Base{
			Identity: sharedmodel.Identity{ID: 7}, Timestamps: sharedmodel.Timestamps{CreatedAt: now, UpdatedAt: now},
			Version: sharedmodel.Version{Version: 2},
		}, Username: "pixel"},
		Profile: playermodel.Profile{PlayerID: 7, Look: "hd-180-1", Gender: playermodel.GenderMale,
			Timestamps: sharedmodel.Timestamps{CreatedAt: now, UpdatedAt: now}, Version: sharedmodel.Version{Version: 3}},
	}
}

// fakeManager records route player operations.
type fakeManager struct {
	// Manager supplies behavior outside these focused route tests.
	playerservice.Manager
	// record is returned by player operations.
	record playerservice.Record
	// createCalls counts player creations.
	createCalls int
	// missing makes player lookups report no record.
	missing bool
	// operationErr fails player operations.
	operationErr error
}

// Create creates one fixture player.
func (manager *fakeManager) Create(_ context.Context, params playerservice.CreateParams) (playerservice.Record, error) {
	if manager.operationErr != nil {
		return playerservice.Record{}, manager.operationErr
	}
	manager.createCalls++
	manager.record.Player.Username = strings.TrimSpace(params.Username)
	manager.record.Profile.Look = params.Profile.Look
	manager.record.Profile.Gender = params.Profile.Gender
	return manager.record, nil
}

// FindByID finds the fixture by id.
func (manager *fakeManager) FindByID(_ context.Context, id int64) (playerservice.Record, bool, error) {
	return manager.record, !manager.missing && id == manager.record.Player.ID, manager.operationErr
}

// FindByUsername finds the fixture by username.
func (manager *fakeManager) FindByUsername(_ context.Context, username string) (playerservice.Record, bool, error) {
	return manager.record, !manager.missing && strings.EqualFold(username, manager.record.Player.Username), manager.operationErr
}

// Update applies fixture player changes.
func (manager *fakeManager) Update(_ context.Context, _ int64, params playerservice.UpdateParams) (playerservice.Record, error) {
	if manager.operationErr != nil {
		return playerservice.Record{}, manager.operationErr
	}
	if params.Username != nil {
		manager.record.Player.Username = *params.Username
	}
	if params.Motto != nil {
		manager.record.Profile.Motto = *params.Motto
	}

	return manager.record, nil
}

// SoftDelete records fixture player deletion.
func (manager *fakeManager) SoftDelete(_ context.Context, _ int64) error {
	return manager.operationErr
}

// UpdatePrivacy returns the fixture unchanged.
func (manager *fakeManager) UpdatePrivacy(context.Context, int64, playerservice.PrivacyParams) (playerservice.Record, error) {
	return manager.record, nil
}
