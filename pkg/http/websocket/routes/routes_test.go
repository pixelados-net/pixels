package routes

import (
	"context"
	"io"
	stdhttp "net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestCountListReasons verifies read-only admin routes.
func TestCountListReasons(t *testing.T) {
	registry := netconn.NewRegistry()
	disconnected := 0
	mustRegister(t, registry, testConnection(t, "one", "websocket", &disconnected))
	mustRegister(t, registry, testConnection(t, "two", "websocket", &disconnected))
	mustRegister(t, registry, testConnection(t, "raw", "raw", &disconnected))
	app := testApp(registry)

	response := testRequest(t, app, stdhttp.MethodGet, "/api/admin/connections/count?kind=websocket", "")
	body := readBody(t, response)
	if response.StatusCode != fiber.StatusOK || !strings.Contains(body, `"count":2`) {
		t.Fatalf("expected websocket count, got %d %s", response.StatusCode, body)
	}

	response = testRequest(t, app, stdhttp.MethodGet, "/api/admin/connections?kind=websocket", "")
	body = readBody(t, response)
	if !strings.Contains(body, `"total":2`) || strings.Contains(body, "remote") {
		t.Fatalf("expected safe websocket list, got %s", body)
	}

	response = testRequest(t, app, stdhttp.MethodGet, "/api/admin/connections/list?kind=websocket", "")
	body = readBody(t, response)
	if !strings.Contains(body, `"total":2`) || strings.Contains(body, "remote") {
		t.Fatalf("expected safe websocket list alias, got %s", body)
	}

	response = testRequest(t, app, stdhttp.MethodGet, "/api/admin/connections/reasons", "")
	body = readBody(t, response)
	if !strings.Contains(body, `"reason":"kicked"`) {
		t.Fatalf("expected disconnect reasons, got %s", body)
	}
}

// TestDisconnectRoutes verifies single, kind, and all disconnect operations.
func TestDisconnectRoutes(t *testing.T) {
	registry := netconn.NewRegistry()
	disconnected := 0
	mustRegister(t, registry, testConnection(t, "one", "websocket", &disconnected))
	mustRegister(t, registry, testConnection(t, "two", "websocket", &disconnected))
	mustRegister(t, registry, testConnection(t, "raw", "raw", &disconnected))
	app := testApp(registry)

	body := `{"reason":"kicked","message":"test"}`
	response := testRequest(t, app, stdhttp.MethodPost, "/api/admin/connections/websocket/one/disconnect", body)
	if response.StatusCode != fiber.StatusOK || registry.Count("websocket") != 1 {
		t.Fatalf("expected single disconnect, got %d", response.StatusCode)
	}

	response = testRequest(t, app, stdhttp.MethodPost, "/api/admin/connections/websocket/disconnect", body)
	if response.StatusCode != fiber.StatusOK || registry.Count("websocket") != 0 {
		t.Fatalf("expected kind disconnect, got %d", response.StatusCode)
	}

	response = testRequest(t, app, stdhttp.MethodPost, "/api/admin/connections/disconnect", body)
	if response.StatusCode != fiber.StatusOK || registry.CountAll() != 0 || disconnected != 3 {
		t.Fatalf("expected all disconnect, got %d count %d disconnected %d", response.StatusCode, registry.CountAll(), disconnected)
	}
}

// TestDisconnectRoutesRejectInvalidInput verifies bad route inputs.
func TestDisconnectRoutesRejectInvalidInput(t *testing.T) {
	registry := netconn.NewRegistry()
	app := testApp(registry)

	response := testRequest(t, app, stdhttp.MethodPost, "/api/admin/connections/websocket/missing/disconnect", `{"reason":"kicked"}`)
	if response.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected not found, got %d", response.StatusCode)
	}

	response = testRequest(t, app, stdhttp.MethodPost, "/api/admin/connections/disconnect", `{"reason":"nope"}`)
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", response.StatusCode)
	}
}

// testApp creates an admin route app.
func testApp(registry *netconn.Registry) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	Register(app, registry)

	return app
}

// testConnection creates a registry connection.
func testConnection(t *testing.T, id netconn.ID, kind netconn.Kind, disconnected *int) *netconn.Session {
	t.Helper()
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:        id,
		Kind:      kind,
		StartedAt: time.Unix(10, 0),
		Sender: func(context.Context, codec.Packet) error {
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			*disconnected = *disconnected + 1
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	return session
}

// mustRegister registers a connection or fails.
func mustRegister(t *testing.T, registry *netconn.Registry, connection netconn.Connection) {
	t.Helper()
	if err := registry.Register(connection); err != nil {
		t.Fatalf("register connection: %v", err)
	}
}

// testRequest sends a route test request.
func testRequest(t *testing.T, app *fiber.App, method string, path string, body string) *stdhttp.Response {
	t.Helper()
	request, err := stdhttp.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}

	return response
}

// readBody reads a response body.
func readBody(t *testing.T, response *stdhttp.Response) string {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return string(body)
}
