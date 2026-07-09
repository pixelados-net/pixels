package http

import (
	stdhttp "net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/connection"
	"github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestHTTPLogsIncludeFailureReason verifies client errors include useful log context.
func TestHTTPLogsIncludeFailureReason(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	app := testAppWithLogger(t, "development", zap.New(core))
	response := testRequest(t, app, stdhttp.MethodPost, "/api/sso/tickets")

	if response.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", response.StatusCode)
	}

	body := readBody(t, response)
	if !strings.Contains(body, "missing or invalid api key") {
		t.Fatalf("expected meaningful response error, got %s", body)
	}

	if logs.Len() != 1 {
		t.Fatalf("expected one warning log, got %d", logs.Len())
	}

	fields := logs.All()[0].ContextMap()
	if fields["error"] != "missing or invalid api key" {
		t.Fatalf("expected meaningful log error, got %#v", fields["error"])
	}
}

// testAppWithLogger creates a route test app with a custom logger.
func testAppWithLogger(t *testing.T, environment string, log *zap.Logger) *fiber.App {
	t.Helper()

	service := testSSO(t)
	registry := netconn.NewRegistry()
	config := testConfig(environment)
	adapter := ws.New(ws.Config{}, config.App, registry, connection.NewHandlers(service, testFinder{}, live.NewRegistry(), binding.NewRegistry(), bus.New(), nil), zap.NewNop(), config.Logger)

	return New(log, config, testInfo(), service, adapter, testRooms(), testRoomRuntime(), testNavigator(), testCurrencyDependencies(registry, log))
}
