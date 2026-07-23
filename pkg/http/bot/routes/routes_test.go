package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	botadmin "github.com/niflaot/pixels/internal/realm/bot/admin"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// routeStore supplies bartender administration records.
type routeStore struct {
	botrecord.Store
	// items stores current mappings.
	items []botrecord.ServeItem
}

// ListServeItems returns current mappings.
func (store *routeStore) ListServeItems(context.Context) ([]botrecord.ServeItem, error) {
	return store.items, nil
}

// CreateServeItem appends one mapping.
func (store *routeStore) CreateServeItem(_ context.Context, keyword string, definitionID int64) (botrecord.ServeItem, error) {
	item := botrecord.ServeItem{ID: int64(len(store.items) + 1), Keyword: keyword, DefinitionID: definitionID}
	store.items = append(store.items, item)
	return item, nil
}

// TestServeItemRoutesValidateAndReturnJSON verifies protected route behavior after middleware.
func TestServeItemRoutesValidateAndReturnJSON(t *testing.T) {
	store := &routeStore{items: []botrecord.ServeItem{{ID: 1, Keyword: "tea", DefinitionID: 5}}}
	app := fiber.New()
	Register(app, Dependencies{Bots: botadmin.New(store, &botcore.Service{}, nil)})
	request := httptest.NewRequest(http.MethodGet, basePath+"/serve-items", nil)
	response, err := app.Test(request)
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("list status=%d err=%v", response.StatusCode, err)
	}
	request = httptest.NewRequest(http.MethodPost, basePath+"/serve-items", nil)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err = app.Test(request)
	if err != nil || response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid status=%d err=%v", response.StatusCode, err)
	}
}

// TestPositiveIDRejectsMalformedPaths verifies bounded path parsing.
func TestPositiveIDRejectsMalformedPaths(t *testing.T) {
	for _, value := range []string{"", "0", "-1", "abc"} {
		if _, err := positiveID(value); err == nil {
			t.Fatalf("expected rejection for %q", value)
		}
	}
}
