package routes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// routeEffectManager records administrative effect grants.
type routeEffectManager struct {
	// Manager supplies unused effect operations.
	playereffect.Manager
	// effect stores the last granted stack.
	effect playereffect.Effect
}

// Grant records one administrative effect charge.
func (manager *routeEffectManager) Grant(_ context.Context, playerID int64, effectID int32, duration int32, _ playereffect.Source) (playereffect.Effect, error) {
	manager.effect = playereffect.Effect{PlayerID: playerID, ID: effectID, DurationSeconds: duration, RemainingCharges: 1}
	return manager.effect, nil
}

// TestGrantEffectReturnsDocumentedStack verifies the protected support route.
func TestGrantEffectReturnsDocumentedStack(t *testing.T) {
	effects := &routeEffectManager{}
	app, _ := testApplication(t, effects)
	response, err := app.Test(requestForTest(t, http.MethodPost, "/api/admin/players/7/effects", `{"effectId":101,"durationSeconds":60}`))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != fiber.StatusOK || effects.effect.ID != 101 || !strings.Contains(string(body), `"remainingCharges":1`) {
		t.Fatalf("status=%d effect=%#v body=%s", response.StatusCode, effects.effect, body)
	}
}

// TestReadByUsernameFindsCaseInsensitivePlayer verifies the exact username route.
func TestReadByUsernameFindsCaseInsensitivePlayer(t *testing.T) {
	app, _ := testApplication(t)
	response, err := app.Test(requestForTest(t, http.MethodGet, "/api/admin/players/by-username/PIXEL", ""))
	if err != nil {
		t.Fatalf("find player: %v", err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected ok, got %d", response.StatusCode)
	}
}

// TestUpdateChangesPlayerProfile verifies partial administrative profile changes.
func TestUpdateChangesPlayerProfile(t *testing.T) {
	app, manager := testApplication(t)
	response, err := app.Test(requestForTest(t, http.MethodPatch, "/api/admin/players/7", `{"username":"renamed","motto":"updated"}`))
	if err != nil {
		t.Fatalf("update player: %v", err)
	}
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != fiber.StatusOK || !strings.Contains(string(body), `"username":"renamed"`) || !strings.Contains(string(body), `"motto":"updated"`) {
		t.Fatalf("unexpected update status=%d body=%s", response.StatusCode, body)
	}
	if manager.record.Player.Username != "renamed" {
		t.Fatalf("expected manager update, got %q", manager.record.Player.Username)
	}
}

// TestSoftDeleteReturnsNoContent verifies the administrative deletion route.
func TestSoftDeleteReturnsNoContent(t *testing.T) {
	app, _ := testApplication(t)
	response, err := app.Test(requestForTest(t, http.MethodDelete, "/api/admin/players/7", ""))
	if err != nil {
		t.Fatalf("delete player: %v", err)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected no content, got %d", response.StatusCode)
	}
}

// TestRoutesRejectInvalidRequests verifies malformed administrative input.
func TestRoutesRejectInvalidRequests(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{name: "invalid id", method: http.MethodGet, path: "/api/admin/players/0", status: fiber.StatusBadRequest},
		{name: "invalid patch", method: http.MethodPatch, path: "/api/admin/players/7", body: `{`, status: fiber.StatusBadRequest},
		{name: "conflicting home room", method: http.MethodPatch, path: "/api/admin/players/7", body: `{"homeRoomId":2,"clearHomeRoom":true}`, status: fiber.StatusBadRequest},
		{name: "invalid create", method: http.MethodPost, path: "/api/admin/players", body: `{`, status: fiber.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app, _ := testApplication(t)
			request := requestForTest(t, test.method, test.path, test.body)
			if test.method == http.MethodPost {
				request.Header.Set(idempotencyHeader, "route-validation")
			}
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("request route: %v", err)
			}
			if response.StatusCode != test.status {
				t.Fatalf("expected %d, got %d", test.status, response.StatusCode)
			}
		})
	}
}

// TestRoutesMapMissingAndConflictOutcomes verifies stable HTTP errors.
func TestRoutesMapMissingAndConflictOutcomes(t *testing.T) {
	app, manager := testApplication(t)
	manager.missing = true
	for _, path := range []string{"/api/admin/players/7", "/api/admin/players/by-username/pixel"} {
		response, err := app.Test(requestForTest(t, http.MethodGet, path, ""))
		if err != nil {
			t.Fatalf("read missing player: %v", err)
		}
		if response.StatusCode != fiber.StatusNotFound {
			t.Fatalf("expected not found for %s, got %d", path, response.StatusCode)
		}
	}

	manager.missing = false
	manager.operationErr = playerservice.ErrUsernameTaken
	response, err := app.Test(requestForTest(t, http.MethodPatch, "/api/admin/players/7", `{"username":"duplicate"}`))
	if err != nil {
		t.Fatalf("update duplicate player: %v", err)
	}
	if response.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected conflict, got %d", response.StatusCode)
	}

	manager.operationErr = playerservice.ErrPlayerNotFound
	response, err = app.Test(requestForTest(t, http.MethodDelete, "/api/admin/players/7", ""))
	if err != nil {
		t.Fatalf("delete missing player: %v", err)
	}
	if response.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected not found, got %d", response.StatusCode)
	}
}

// TestUpdateParamsMapsOptionalFields verifies the full patch contract.
func TestUpdateParamsMapsOptionalFields(t *testing.T) {
	username := "renamed"
	look := "hd-190-1"
	gender := "F"
	motto := "updated"
	homeRoomID := int64(9)
	flag := true
	bubble := int32(4)
	params := updateParams(UpdateRequest{Username: &username, Look: &look, Gender: &gender, Motto: &motto,
		HomeRoomID: &homeRoomID, AllowNameChange: &flag, BubbleStyle: &bubble,
		BlockFriendRequests: &flag, BlockRoomInvites: &flag, BlockFollowing: &flag})
	if params.Gender == nil || string(*params.Gender) != gender || params.HomeRoomID == nil || *params.HomeRoomID == nil || **params.HomeRoomID != homeRoomID {
		t.Fatalf("unexpected mapped params %#v", params)
	}

	cleared := updateParams(UpdateRequest{ClearHomeRoom: true})
	if cleared.HomeRoomID == nil || *cleared.HomeRoomID != nil {
		t.Fatalf("expected explicit home room clearing, got %#v", cleared.HomeRoomID)
	}
}

// TestPlayerErrorMapsDomainFailures verifies administrative HTTP status mapping.
func TestPlayerErrorMapsDomainFailures(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{err: playerservice.ErrInvalidGender, status: fiber.StatusBadRequest},
		{err: playerservice.ErrConflict, status: fiber.StatusConflict},
		{err: playerservice.ErrPlayerNotFound, status: fiber.StatusNotFound},
	}
	for _, test := range tests {
		mapped, ok := playerError(test.err).(*fiber.Error)
		if !ok || mapped.Code != test.status {
			t.Fatalf("expected status %d for %v, got %#v", test.status, test.err, mapped)
		}
	}

	expected := errors.New("database failed")
	if mapped := playerError(expected); !errors.Is(mapped, expected) {
		t.Fatalf("expected original error, got %v", mapped)
	}
	if mapped := idempotencyError(expected); !errors.Is(mapped, expected) {
		t.Fatalf("expected original idempotency error, got %v", mapped)
	}
}

// TestNewestTimeSelectsLatestMutation verifies representation timestamps.
func TestNewestTimeSelectsLatestMutation(t *testing.T) {
	first := time.Date(2026, 7, 13, 1, 0, 0, 0, time.UTC)
	second := first.Add(time.Minute)
	if newestTime(first, second) != second || newestTime(second, first) != second {
		t.Fatal("expected newest mutation timestamp")
	}
}

// TestLiveProjectionIgnoresOfflinePlayers verifies optional runtime integration.
func TestLiveProjectionIgnoresOfflinePlayers(t *testing.T) {
	handler := handler{live: playerlive.NewRegistry()}
	handler.project(testRecord())
	handler.disconnectDeleted(context.Background(), 7)
}

// requestForTest creates one JSON route request.
func requestForTest(t *testing.T, method string, path string, body string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}

	return request
}
