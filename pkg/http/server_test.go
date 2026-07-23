package http

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/pkg/build"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestServerFields verifies structured startup field construction.
func TestServerFields(t *testing.T) {
	fields := serverFields(testConfig("test"), build.NewInfo("pixels", "1.0.0", "abcdef1234"))

	if len(fields) != 5 {
		t.Fatalf("expected five fields, got %d", len(fields))
	}
}

// TestStartRegistersLifecycle verifies Fiber binds and shuts down from lifecycle hooks.
func TestStartRegistersLifecycle(t *testing.T) {
	lifecycle := fxtest.NewLifecycle(t)
	config := testConfig("test")
	config.App.Port = 0

	Start(lifecycle, fiber.New(), zap.NewNop(), config, testInfo())

	lifecycle.RequireStart()
	lifecycle.RequireStop()
}
