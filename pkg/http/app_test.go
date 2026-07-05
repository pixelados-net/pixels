package http

import (
	"io"
	stdhttp "net/http"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	"github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/config"
	appconfig "github.com/niflaot/pixels/pkg/config/app"
	"github.com/niflaot/pixels/pkg/http/openapi"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	"github.com/niflaot/pixels/pkg/logger"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/zap"
)

// TestStatusRouteIsPublic verifies status responses do not require an API key.
func TestStatusRouteIsPublic(t *testing.T) {
	app := testApp(t, "development")
	response := testRequest(t, app, stdhttp.MethodGet, "/status")

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response)
	if !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("expected status body, got %s", body)
	}
}

// TestNewDisablesStartupMessage verifies Fiber's welcome banner is disabled.
func TestNewDisablesStartupMessage(t *testing.T) {
	app := testApp(t, "development")

	if !app.Config().DisableStartupMessage {
		t.Fatal("expected Fiber startup message to be disabled")
	}
}

// TestDocsRoutesAreDevelopmentOnly verifies Scalar docs are public in development.
func TestDocsRoutesAreDevelopmentOnly(t *testing.T) {
	app := testApp(t, "development")
	response := testRequest(t, app, stdhttp.MethodGet, "/docs")

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected docs status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response)
	if !strings.Contains(body, "@scalar/api-reference") {
		t.Fatalf("expected Scalar docs body, got %s", body)
	}
}

// TestDocsRoutesAreHiddenOutsideDevelopment verifies docs return not found outside development.
func TestDocsRoutesAreHiddenOutsideDevelopment(t *testing.T) {
	app := testApp(t, "production")
	response := testRequest(t, app, stdhttp.MethodGet, "/docs")

	if response.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected docs status 404, got %d", response.StatusCode)
	}
}

// TestOpenAPIIsEmbeddedInDocs verifies Scalar receives an inline OpenAPI document.
func TestOpenAPIIsEmbeddedInDocs(t *testing.T) {
	app := testApp(t, "development")
	response := testRequest(t, app, stdhttp.MethodGet, "/docs")

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected docs status 200, got %d", response.StatusCode)
	}

	body := readBody(t, response)
	if !strings.Contains(body, openapi.Spec) {
		t.Fatalf("expected embedded OpenAPI document, got %s", body)
	}
}

// TestPrivateRoutesRequireAccessKey verifies private routes require API keys.
func TestPrivateRoutesRequireAccessKey(t *testing.T) {
	app := testApp(t, "development")
	response := testRequest(t, app, stdhttp.MethodGet, "/api/admin/connections/count")

	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", response.StatusCode)
	}
}

// TestPrivateRoutesAllowAccessKey verifies authenticated admin routes continue.
func TestPrivateRoutesAllowAccessKey(t *testing.T) {
	app := testApp(t, "development")
	request := newRequest(stdhttp.MethodGet, "/api/admin/connections/count")
	request.Header.Set(apiKeyHeader, "secret")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected ok status 200, got %d", response.StatusCode)
	}
}

// TestCreateSSOTicketRoute verifies private SSO ticket creation.
func TestCreateSSOTicketRoute(t *testing.T) {
	app := testApp(t, "development")
	request, err := stdhttp.NewRequest(stdhttp.MethodPost, "/api/sso/tickets", strings.NewReader(`{"playerId":2,"ip":"127.0.0.1","ttlSeconds":60}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	request.Header.Set(apiKeyHeader, "secret")
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}

	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected created status 201, got %d", response.StatusCode)
	}

	body := readBody(t, response)
	if !strings.Contains(body, `"ticket"`) {
		t.Fatalf("expected ticket response, got %s", body)
	}
}

// TestWebsocketRouteRequiresUpgrade verifies websocket entrypoint is public but requires upgrade.
func TestWebsocketRouteRequiresUpgrade(t *testing.T) {
	app := testApp(t, "development")
	response := testRequest(t, app, stdhttp.MethodGet, "/ws")

	if response.StatusCode != fiber.StatusUpgradeRequired {
		t.Fatalf("expected upgrade status 426, got %d", response.StatusCode)
	}
}

// testApp creates a Fiber app for route tests.
func testApp(t *testing.T, environment string) *fiber.App {
	t.Helper()

	service := testSSO(t)
	registry := netconn.NewRegistry()
	adapter := ws.New(ws.Config{}, testConfig(environment).App, registry, realmconn.NewHandlers(service, testFinder{}, live.NewRegistry(), binding.NewRegistry(), bus.New()), zap.NewNop())

	return New(zap.NewNop(), testConfig(environment), testInfo(), service, adapter, registry)
}

// testConfig creates composed configuration for route tests.
func testConfig(environment string) config.AppConfig {
	return config.AppConfig{
		App: appconfig.Config{
			Environment: environment,
			Host:        "127.0.0.1",
			Port:        3000,
			AccessKey:   "secret",
		},
		Logger: logger.Config{
			Level:  "info",
			Format: logger.FormatConsole,
		},
		SSO: sso.Config{
			DefaultTTL: time.Minute,
			Key:        "test-sso-key",
			Prefix:     "pixels:sso",
		},
	}
}

// testSSO creates an SSO service for route tests.
func testSSO(t *testing.T) *sso.Service {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.New(redis.Config{Address: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close redis: %v", err)
		}
	})

	return sso.New(sso.Config{DefaultTTL: time.Minute, Key: "test-sso-key", Prefix: "pixels:sso"}, client)
}

// testInfo creates build metadata for route tests.
func testInfo() build.Info {
	return build.NewInfo("pixels", "1.0.0", "1234567890")
}

// testRequest executes an HTTP request against a Fiber app.
func testRequest(t *testing.T, app *fiber.App, method string, path string) *stdhttp.Response {
	t.Helper()

	response, err := app.Test(newRequest(method, path), -1)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}

	return response
}

// newRequest creates an HTTP request for route tests.
func newRequest(method string, path string) *stdhttp.Request {
	request, err := stdhttp.NewRequest(method, path, nil)
	if err != nil {
		panic(err)
	}

	return request
}

// readBody reads a test response body.
func readBody(t *testing.T, response *stdhttp.Response) string {
	t.Helper()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return string(body)
}
