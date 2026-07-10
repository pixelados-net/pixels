package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// fakeManager records permission route calls.
type fakeManager struct {
	// calls stores invoked manager operations.
	calls []string
	// err stores the injected operation failure.
	err error
	// allowed stores the point check result.
	allowed bool
}

// Groups returns one fixture group.
func (manager *fakeManager) Groups(context.Context) ([]permissionmodel.Group, error) {
	manager.calls = append(manager.calls, "groups")
	return []permissionmodel.Group{fixtureGroup()}, manager.err
}

// EffectiveNodes returns one fixture resolution.
func (manager *fakeManager) EffectiveNodes(context.Context, int64) ([]permissionservice.ResolvedNode, error) {
	manager.calls = append(manager.calls, "effective")
	return []permissionservice.ResolvedNode{{Node: "catalog.admin.manage", Allowed: true, Source: "group:admin"}}, manager.err
}

// EffectivePerks returns no fixture perks.
func (manager *fakeManager) EffectivePerks(context.Context, int64) ([]string, error) {
	return nil, manager.err
}

// PrimaryGroup returns one fixture group.
func (manager *fakeManager) PrimaryGroup(context.Context, int64) (permissionmodel.Group, bool, error) {
	return fixtureGroup(), true, manager.err
}

// AffectedPlayerIDs returns no fixture players.
func (manager *fakeManager) AffectedPlayerIDs(context.Context, int64) ([]int64, error) {
	return nil, manager.err
}

// CreateGroup records group creation.
func (manager *fakeManager) CreateGroup(context.Context, permissionservice.CreateGroupParams) (permissionmodel.Group, error) {
	manager.calls = append(manager.calls, "create")
	return fixtureGroup(), manager.err
}

// UpdateGroup records group mutation.
func (manager *fakeManager) UpdateGroup(context.Context, int64, permissionservice.UpdateGroupParams) (permissionmodel.Group, error) {
	manager.calls = append(manager.calls, "update")
	return fixtureGroup(), manager.err
}

// GrantGroupNode records a group node grant.
func (manager *fakeManager) GrantGroupNode(context.Context, int64, permission.Node, bool) error {
	manager.calls = append(manager.calls, "grant-group-node")
	return manager.err
}

// RevokeGroupNode records a group node revocation.
func (manager *fakeManager) RevokeGroupNode(context.Context, int64, permission.Node) error {
	manager.calls = append(manager.calls, "revoke-group-node")
	return manager.err
}

// AddPlayerToGroup records membership creation.
func (manager *fakeManager) AddPlayerToGroup(context.Context, int64, int64) error {
	manager.calls = append(manager.calls, "add-membership")
	return manager.err
}

// RemovePlayerFromGroup records membership removal.
func (manager *fakeManager) RemovePlayerFromGroup(context.Context, int64, int64) error {
	manager.calls = append(manager.calls, "remove-membership")
	return manager.err
}

// GrantPlayerNode records a direct player grant.
func (manager *fakeManager) GrantPlayerNode(context.Context, int64, permission.Node, bool) error {
	manager.calls = append(manager.calls, "grant-player-node")
	return manager.err
}

// RevokePlayerNode records a direct player revocation.
func (manager *fakeManager) RevokePlayerNode(context.Context, int64, permission.Node) error {
	manager.calls = append(manager.calls, "revoke-player-node")
	return manager.err
}

// HasPermission records one point permission check.
func (manager *fakeManager) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	manager.calls = append(manager.calls, "check")
	return manager.allowed, manager.err
}

// AssignDefaultGroup satisfies the permission manager contract.
func (manager *fakeManager) AssignDefaultGroup(context.Context, int64) error {
	return manager.err
}

// fixtureGroup creates one permission group response fixture.
func fixtureGroup() permissionmodel.Group {
	return permissionmodel.Group{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}, Version: sharedmodel.Version{Version: 1}}, Name: "admin", Weight: 100}
}

// TestPermissionRoutes verifies every permission administration endpoint.
func TestPermissionRoutes(t *testing.T) {
	manager := &fakeManager{allowed: true}
	app := fiber.New()
	Register(app, Dependencies{Permissions: manager})
	node := url.PathEscape("catalog.admin.manage")
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{name: "nodes", method: http.MethodGet, path: basePath + "/nodes", status: http.StatusOK},
		{name: "groups", method: http.MethodGet, path: basePath + "/groups", status: http.StatusOK},
		{name: "create group", method: http.MethodPost, path: basePath + "/groups", body: `{"name":"moderator","weight":50}`, status: http.StatusCreated},
		{name: "update group", method: http.MethodPatch, path: basePath + "/groups/2", body: `{"weight":60}`, status: http.StatusOK},
		{name: "grant group node", method: http.MethodPost, path: basePath + "/groups/2/nodes", body: `{"node":"catalog.admin.manage","allowed":true}`, status: http.StatusOK},
		{name: "revoke group node", method: http.MethodDelete, path: basePath + "/groups/2/nodes/" + node, status: http.StatusNoContent},
		{name: "add membership", method: http.MethodPost, path: basePath + "/players/3/groups/2", status: http.StatusOK},
		{name: "remove membership", method: http.MethodDelete, path: basePath + "/players/3/groups/2", status: http.StatusNoContent},
		{name: "grant player node", method: http.MethodPost, path: basePath + "/players/3/nodes", body: `{"node":"catalog.admin.manage","allowed":false}`, status: http.StatusOK},
		{name: "revoke player node", method: http.MethodDelete, path: basePath + "/players/3/nodes/" + node, status: http.StatusNoContent},
		{name: "effective", method: http.MethodGet, path: basePath + "/players/3/effective", status: http.StatusOK},
		{name: "check", method: http.MethodGet, path: basePath + "/players/3/check?node=catalog.admin.manage", status: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(t, app, test.method, test.path, test.body)
			defer response.Body.Close()
			if response.StatusCode != test.status {
				body, _ := io.ReadAll(response.Body)
				t.Fatalf("expected status %d, got %d: %s", test.status, response.StatusCode, body)
			}
		})
	}
}

// TestPermissionRouteFailures verifies invalid input and domain error mapping.
func TestPermissionRouteFailures(t *testing.T) {
	manager := &fakeManager{err: permissionservice.ErrConflict}
	app := fiber.New()
	Register(app, Dependencies{Permissions: manager})
	tests := []struct {
		path   string
		body   string
		status int
	}{
		{path: basePath + "/groups/nope", body: `{}`, status: http.StatusBadRequest},
		{path: basePath + "/groups/2", body: `{}`, status: http.StatusConflict},
		{path: basePath + "/groups/2", body: `{`, status: http.StatusBadRequest},
	}
	for _, test := range tests {
		response := performRequest(t, app, http.MethodPatch, test.path, test.body)
		response.Body.Close()
		if response.StatusCode != test.status {
			t.Fatalf("expected status %d, got %d for %s", test.status, response.StatusCode, test.path)
		}
	}
}

// performRequest executes one Fiber test request.
func performRequest(t *testing.T, app *fiber.App, method string, path string, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if strings.TrimSpace(body) != "" {
		request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}

	return response
}

// jsonBody decodes one response body for focused assertions.
func jsonBody(t *testing.T, response *http.Response, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
