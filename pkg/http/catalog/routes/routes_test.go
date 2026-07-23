package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdated "github.com/niflaot/pixels/networking/outbound/catalog/updated"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// TestCatalogAdministrationRoutes verifies page, item, sanitize, and refresh endpoints.
func TestCatalogAdministrationRoutes(t *testing.T) {
	admin := &routeAdmin{}
	registry := netconn.NewRegistry()
	sent := registerRouteConnection(t, registry)
	app := fiber.New()
	Register(app, Dependencies{Catalog: admin, Connections: registry, Log: zap.NewNop()})

	assertStatus(t, app, http.MethodPost, basePath+"/pages", PageRequest{Name: "chairs", Layout: "default_3x3", Visible: true, Enabled: true}, http.StatusOK)
	pageName := "chairs_updated"
	pageLayout := "spaces"
	pageVisible := false
	assertStatus(t, app, http.MethodPatch, basePath+"/pages/1", PagePatchRequest{Name: &pageName, Layout: &pageLayout, Visible: &pageVisible}, http.StatusOK)
	assertStatus(t, app, http.MethodGet, basePath+"/pages", nil, http.StatusOK)
	assertStatus(t, app, http.MethodPost, basePath+"/items", ItemRequest{PageID: 1, DefinitionID: 2, Name: "chair", CostCredits: 2, PointsType: -1, Amount: 1, Enabled: true}, http.StatusOK)
	itemPage := int64(1)
	definitionID := int64(2)
	itemName := "chair_updated"
	credits := int64(3)
	points := int64(0)
	pointsType := int32(-1)
	amount := int32(1)
	stack := int32(0)
	bundle := true
	giftable := true
	club := false
	order := int32(2)
	enabled := true
	extra := "0"
	assertStatus(t, app, http.MethodPatch, basePath+"/items/1", ItemPatchRequest{PageID: &itemPage, DefinitionID: &definitionID,
		Name: &itemName, CostCredits: &credits, CostPoints: &points, PointsType: &pointsType, Amount: &amount,
		LimitedStack: &stack, BundleDiscountEnabled: &bundle, Giftable: &giftable, ClubOnly: &club,
		OrderNum: &order, Enabled: &enabled, ExtraData: &extra}, http.StatusOK)
	assertStatus(t, app, http.MethodGet, basePath+"/items?pageId=1", nil, http.StatusOK)
	assertStatus(t, app, http.MethodGet, basePath+"/sanitize-list", nil, http.StatusOK)
	assertStatus(t, app, http.MethodPost, basePath+"/refresh", nil, http.StatusOK)
	if admin.refreshes != 1 || len(*sent) != 5 || (*sent)[0].Header != outupdated.Header {
		t.Fatalf("unexpected refreshes=%d packets=%#v", admin.refreshes, *sent)
	}
	assertStatus(t, app, http.MethodDelete, basePath+"/items/1", nil, http.StatusNoContent)
	if len(*sent) != 6 {
		t.Fatalf("expected one publication per mutation, got %d", len(*sent))
	}
}

// TestCatalogErrorMapsAdministrationFailures verifies HTTP status mapping.
func TestCatalogErrorMapsAdministrationFailures(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{err: catalogadmin.ErrItemNotFound, status: fiber.StatusNotFound},
		{err: catalogadmin.ErrConflict, status: fiber.StatusConflict},
		{err: catalogadmin.ErrInvalidPage, status: fiber.StatusBadRequest},
	}
	for _, test := range tests {
		mapped := catalogError(test.err)
		fiberError, ok := mapped.(*fiber.Error)
		if !ok || fiberError.Code != test.status {
			t.Fatalf("unexpected mapped error %#v", mapped)
		}
	}
}

// TestCatalogAdministrationRoutesMapInvalidInput verifies stable client failures.
func TestCatalogAdministrationRoutesMapInvalidInput(t *testing.T) {
	app := fiber.New()
	Register(app, Dependencies{Catalog: &routeAdmin{createItemErr: catalogadmin.ErrDefinitionNotFound}, Connections: netconn.NewRegistry(), Log: zap.NewNop()})
	assertStatus(t, app, http.MethodPost, basePath+"/items", ItemRequest{}, http.StatusBadRequest)
	assertStatus(t, app, http.MethodPatch, basePath+"/items/nope", ItemPatchRequest{}, http.StatusBadRequest)
	assertStatus(t, app, http.MethodGet, basePath+"/items?pageId=nope", nil, http.StatusBadRequest)
}

// routeAdmin stores mutable HTTP catalog fixtures.
type routeAdmin struct {
	// pages stores page fixtures.
	pages []catalogmodel.Page
	// items stores offer fixtures.
	items []catalogmodel.Item
	// refreshes counts refresh calls.
	refreshes int
	// createItemErr stores an optional item creation failure.
	createItemErr error
}

// Pages lists page fixtures.
func (admin *routeAdmin) Pages(context.Context) ([]catalogmodel.Page, error) { return admin.pages, nil }

// CreatePage creates a page fixture.
func (admin *routeAdmin) CreatePage(_ context.Context, input catalogadmin.PageInput) (catalogmodel.Page, error) {
	page := catalogmodel.Page{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, Name: input.Name, Layout: input.Layout,
		RequiredNode: input.RequiredNode, Visible: input.Visible, Enabled: input.Enabled}
	admin.pages = append(admin.pages, page)
	return page, nil
}

// UpdatePage updates a page fixture.
func (admin *routeAdmin) UpdatePage(_ context.Context, id int64, _ catalogadmin.PagePatch) (catalogmodel.Page, error) {
	if len(admin.pages) == 0 {
		return catalogmodel.Page{}, catalogadmin.ErrPageNotFound
	}
	admin.pages[0].ID = id
	return admin.pages[0], nil
}

// Items lists offer fixtures.
func (admin *routeAdmin) Items(context.Context, *int64) ([]catalogmodel.Item, error) {
	return admin.items, nil
}

// CreateItem creates an offer fixture.
func (admin *routeAdmin) CreateItem(_ context.Context, input catalogadmin.ItemInput) (catalogmodel.Item, error) {
	if admin.createItemErr != nil {
		return catalogmodel.Item{}, admin.createItemErr
	}
	item := catalogmodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, PageID: input.PageID,
		DefinitionID: input.DefinitionID, Name: input.Name, CostCredits: input.CostCredits, PointsType: input.PointsType,
		Amount: input.Amount, Enabled: input.Enabled}
	admin.items = append(admin.items, item)
	return item, nil
}

// UpdateItem returns an updated offer fixture.
func (admin *routeAdmin) UpdateItem(_ context.Context, id int64, _ catalogadmin.ItemPatch) (catalogmodel.Item, error) {
	if len(admin.items) == 0 {
		return catalogmodel.Item{}, catalogadmin.ErrItemNotFound
	}
	admin.items[0].ID = id
	return admin.items[0], nil
}

// DeleteItem deletes an offer fixture.
func (admin *routeAdmin) DeleteItem(context.Context, int64) error { admin.items = nil; return nil }

// SanitizeList lists one orphan definition fixture.
func (admin *routeAdmin) SanitizeList(context.Context) ([]furnituremodel.Definition, error) {
	return []furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, SpriteID: 39, Name: "chair"}}, nil
}

// Refresh records one catalog refresh.
func (admin *routeAdmin) Refresh(context.Context) error { admin.refreshes++; return nil }

// assertStatus executes one JSON request and verifies its status.
func assertStatus(t *testing.T, app *fiber.App, method string, path string, body any, expected int) {
	t.Helper()
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	request, _ := http.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	defer response.Body.Close()
	if response.StatusCode != expected {
		t.Fatalf("request %s %s expected %d, got %d", method, path, expected, response.StatusCode)
	}
}

// registerRouteConnection registers a packet-capturing connection.
func registerRouteConnection(t *testing.T, registry *netconn.Registry) *[]codec.Packet {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "catalog", Kind: "websocket", Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("new route connection: %v", err)
	}
	if err := registry.Register(session); err != nil {
		t.Fatalf("register route connection: %v", err)
	}

	return &sent
}
