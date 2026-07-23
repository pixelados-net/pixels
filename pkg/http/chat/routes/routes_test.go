package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	"github.com/niflaot/pixels/internal/realm/chat/bubble"
	bubblerepo "github.com/niflaot/pixels/internal/realm/chat/bubble/repository"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// filterStoreForTest stores words in memory.
type filterStoreForTest struct{ words []string }

// bubbleStoreForTest stores unlock thresholds.
type bubbleStoreForTest struct{ weights map[int32]int32 }

// List returns bubble thresholds.
func (store *bubbleStoreForTest) List(context.Context) ([]bubblerepo.Unlock, error) {
	items := make([]bubblerepo.Unlock, 0, len(store.weights))
	for id, weight := range store.weights {
		items = append(items, bubblerepo.Unlock{BubbleID: id, MinWeight: weight})
	}
	return items, nil
}

// MinWeight returns one threshold.
func (store *bubbleStoreForTest) MinWeight(_ context.Context, id int32) (int32, bool, error) {
	weight, found := store.weights[id]
	return weight, found, nil
}

// Set stores one threshold.
func (store *bubbleStoreForTest) Set(_ context.Context, id int32, weight int32) error {
	store.weights[id] = weight
	return nil
}

// Delete removes one threshold.
func (store *bubbleStoreForTest) Delete(_ context.Context, id int32) error {
	delete(store.weights, id)
	return nil
}

// profileStoreForTest accepts bubble profile writes.
type profileStoreForTest struct{}

// UpdateBubbleStyle returns one profile.
func (profileStoreForTest) UpdateBubbleStyle(_ context.Context, id int64, style int32) (playermodel.Profile, error) {
	return playermodel.Profile{PlayerID: id, BubbleStyle: style}, nil
}

// permissionsForTest allows no bubble bypass and returns a zero-weight group.
type permissionsForTest struct{}

// HasPermission denies one capability.
func (permissionsForTest) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return false, nil
}

// PrimaryGroup returns one zero-weight group.
func (permissionsForTest) PrimaryGroup(context.Context, int64) (permissionmodel.Group, bool, error) {
	return permissionmodel.Group{}, true, nil
}

// historyStoreForTest stores one response row.
type historyStoreForTest struct{}

// InsertBatch accepts history writes.
func (historyStoreForTest) InsertBatch(context.Context, []historymodel.Entry) error { return nil }

// History returns one response row.
func (historyStoreForTest) History(context.Context, historymodel.Query) ([]historymodel.Entry, error) {
	return []historymodel.Entry{{ID: 1, RoomID: 9, PlayerID: 2, Kind: "talk", Message: "hello"}}, nil
}

// EnsurePartitions accepts history maintenance.
func (historyStoreForTest) EnsurePartitions(context.Context, time.Time, time.Time) error { return nil }

// DropBefore accepts history maintenance.
func (historyStoreForTest) DropBefore(context.Context, time.Time) error { return nil }

// List returns stored words.
func (store *filterStoreForTest) List(context.Context) ([]string, error) {
	return append([]string(nil), store.words...), nil
}

// Add inserts one word.
func (store *filterStoreForTest) Add(_ context.Context, word string) error {
	store.words = append(store.words, word)
	return nil
}

// Remove deletes one word.
func (store *filterStoreForTest) Remove(_ context.Context, word string) error {
	for index, current := range store.words {
		if current == word {
			store.words = append(store.words[:index], store.words[index+1:]...)
		}
	}
	return nil
}

// TestFilterRoutesMutateAndList verifies protected route behavior in isolation.
func TestFilterRoutesMutateAndList(t *testing.T) {
	filter := chatfilter.New(&filterStoreForTest{})
	app := fiber.New()
	Register(app, Dependencies{Filter: filter})
	request, _ := http.NewRequest(http.MethodPost, chatPath+"/filters", bytes.NewBufferString(`{"word":"Bad"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil || response.StatusCode != http.StatusCreated {
		t.Fatalf("post status=%d err=%v", response.StatusCode, err)
	}
	response, err = app.Test(mustRequest(t, http.MethodGet, chatPath+"/filters"))
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("get status=%d err=%v", response.StatusCode, err)
	}
	var body FilterListResponse
	if err = json.NewDecoder(response.Body).Decode(&body); err != nil || body.Total != 1 || body.Items[0] != "bad" {
		t.Fatalf("body=%#v err=%v", body, err)
	}
	response, err = app.Test(mustRequest(t, http.MethodDelete, chatPath+"/filters/bad"))
	if err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status=%d err=%v", response.StatusCode, err)
	}
}

// TestBubbleAndHistoryRoutes verifies threshold mutations and both history scopes.
func TestBubbleAndHistoryRoutes(t *testing.T) {
	bubbles := bubble.New(&bubbleStoreForTest{weights: make(map[int32]int32)}, profileStoreForTest{}, permissionsForTest{}, playerlive.NewRegistry(), "bubble.any")
	app := fiber.New()
	Register(app, Dependencies{Filter: chatfilter.New(&filterStoreForTest{}), Bubbles: bubbles, History: history.NewService(historyStoreForTest{})})
	request, _ := http.NewRequest(http.MethodPut, chatPath+"/bubbles/4", bytes.NewBufferString(`{"minWeight":20}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("put bubble status=%d err=%v", response.StatusCode, err)
	}
	for _, path := range []string{chatPath + "/bubbles", "/api/admin/rooms/9/chat/history?before=2&limit=999", "/api/admin/players/2/chat/history?roomId=9"} {
		response, err = app.Test(mustRequest(t, http.MethodGet, path))
		if err != nil || response.StatusCode != http.StatusOK {
			t.Fatalf("get %s status=%d err=%v", path, response.StatusCode, err)
		}
	}
	response, err = app.Test(mustRequest(t, http.MethodDelete, chatPath+"/bubbles/4"))
	if err != nil || response.StatusCode != http.StatusNoContent {
		t.Fatalf("delete bubble status=%d err=%v", response.StatusCode, err)
	}
}

// TestRouteValidationRejectsMalformedInputs verifies meaningful client errors.
func TestRouteValidationRejectsMalformedInputs(t *testing.T) {
	app := fiber.New()
	bubbles := bubble.New(&bubbleStoreForTest{weights: make(map[int32]int32)}, profileStoreForTest{}, permissionsForTest{}, playerlive.NewRegistry(), "bubble.any")
	Register(app, Dependencies{Filter: chatfilter.New(&filterStoreForTest{}), Bubbles: bubbles, History: history.NewService(historyStoreForTest{})})
	for _, path := range []string{chatPath + "/bubbles/nope", "/api/admin/rooms/nope/chat/history", "/api/admin/players/0/chat/history"} {
		response, err := app.Test(mustRequest(t, http.MethodGet, path))
		if err != nil || response.StatusCode < 400 {
			t.Fatalf("path=%s status=%d err=%v", path, response.StatusCode, err)
		}
	}
}

// TestRouteHelpersNormalizeAndMapExpectedErrors verifies pagination and domain error mapping.
func TestRouteHelpersNormalizeAndMapExpectedErrors(t *testing.T) {
	if id, err := positiveID("7", "player"); err != nil || id != 7 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if _, err := positiveID("0", "player"); err == nil {
		t.Fatal("expected invalid identity")
	}
	if mapped := mapFilterError(chatfilter.ErrInvalidWord); mapped == nil {
		t.Fatal("expected mapped filter error")
	}
	if mapped := mapBubbleError(bubble.ErrInvalidBubble); mapped == nil {
		t.Fatal("expected mapped bubble error")
	}
	app := fiber.New()
	bubbles := bubble.New(&bubbleStoreForTest{weights: make(map[int32]int32)}, profileStoreForTest{}, permissionsForTest{}, playerlive.NewRegistry(), "bubble.any")
	Register(app, Dependencies{Filter: chatfilter.New(&filterStoreForTest{}), Bubbles: bubbles, History: history.NewService(historyStoreForTest{})})
	request, _ := http.NewRequest(http.MethodPost, chatPath+"/filters", bytes.NewBufferString(`{"word":"two words"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil || response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d err=%v", response.StatusCode, err)
	}
}

// mustRequest creates an HTTP request for tests.
func mustRequest(t *testing.T, method string, path string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	return request
}
