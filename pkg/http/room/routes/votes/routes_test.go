package votes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
)

// fakeManager stores deterministic HTTP vote behavior.
type fakeManager struct {
	// mutation stores cast output.
	mutation roomvotes.Mutation
	// state stores status output.
	state roomvotes.State
	// items stores list output.
	items []roomvotes.Vote
	// err stores a domain failure.
	err error
}

// Cast returns configured mutation output.
func (manager fakeManager) Cast(context.Context, int64, int64) (roomvotes.Mutation, error) {
	return manager.mutation, manager.err
}

// State returns configured status output.
func (manager fakeManager) State(context.Context, int64, int64) (roomvotes.State, error) {
	return manager.state, manager.err
}

// List returns configured vote rows.
func (manager fakeManager) List(context.Context, roomvotes.Query) ([]roomvotes.Vote, error) {
	return manager.items, manager.err
}

// TestRoutesExposeCastStatusAndList verifies the room vote HTTP surface.
func TestRoutesExposeCastStatusAndList(t *testing.T) {
	now := time.Now().UTC()
	app := fiber.New()
	Register(app, fakeManager{mutation: roomvotes.Mutation{Score: 3, Inserted: true}, state: roomvotes.State{Score: 3, Voted: true}, items: []roomvotes.Vote{{RoomID: 7, PlayerID: 2, CreatedAt: now}}})
	cast := performVoteRequest(t, app, http.MethodPost, Path+"/cast", CastRequest{RoomID: 7, PlayerID: 2})
	if cast.StatusCode != http.StatusOK {
		t.Fatalf("cast status=%d", cast.StatusCode)
	}
	status := performVoteRequest(t, app, http.MethodGet, Path+"/status?roomId=7&playerId=2", nil)
	if status.StatusCode != http.StatusOK {
		t.Fatalf("status status=%d", status.StatusCode)
	}
	list := performVoteRequest(t, app, http.MethodGet, Path+"/list?roomId=7&playerId=2&before="+now.Format(time.RFC3339)+"&limit=10", nil)
	if list.StatusCode != http.StatusOK {
		t.Fatalf("list status=%d", list.StatusCode)
	}
	var response ListResponse
	if err := json.NewDecoder(list.Body).Decode(&response); err != nil || response.Total != 1 || response.Items[0].PlayerID != 2 {
		t.Fatalf("list response=%+v err=%v", response, err)
	}
}

// TestRoutesRejectMalformedInput verifies meaningful client failures.
func TestRoutesRejectMalformedInput(t *testing.T) {
	app := fiber.New()
	Register(app, fakeManager{})
	response := performVoteRequest(t, app, http.MethodGet, Path+"/status?roomId=nope&playerId=2", nil)
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d", response.StatusCode)
	}
	cases := []string{
		Path + "/status?roomId=7",
		Path + "/list?roomId=7&playerId=nope",
		Path + "/list?roomId=7&before=nope",
	}
	for _, path := range cases {
		response = performVoteRequest(t, app, http.MethodGet, path, nil)
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("path=%s status=%d", path, response.StatusCode)
		}
	}
	response = performVoteRequest(t, app, http.MethodPost, Path+"/cast", "invalid")
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("cast status=%d", response.StatusCode)
	}
}

// TestRoutesMapMissingRooms verifies domain-to-HTTP error mapping.
func TestRoutesMapMissingRooms(t *testing.T) {
	app := fiber.New()
	Register(app, fakeManager{err: roomvotes.ErrRoomNotFound})
	response := performVoteRequest(t, app, http.MethodGet, Path+"/status?roomId=7&playerId=2", nil)
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status=%d", response.StatusCode)
	}
}

// performVoteRequest executes one route request.
func performVoteRequest(t *testing.T, app *fiber.App, method string, path string, body any) *http.Response {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	request := httptest.NewRequest(method, path, &payload)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}

	return response
}
