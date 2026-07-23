package pluginroutes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// TestRegistryMountsIsolatedRoutesAndSpec verifies plugin path isolation.
func TestRegistryMountsIsolatedRoutesAndSpec(t *testing.T) {
	registry := New()
	if err := registry.Mount("hello", func(router fiber.Router) {
		router.Get("/health", func(ctx *fiber.Ctx) error { return ctx.SendString("ok") })
	}); err != nil {
		t.Fatalf("mount route: %v", err)
	}
	if err := registry.Describe("hello", []byte(`{"openapi":"3.0.3"}`)); err != nil {
		t.Fatalf("describe plugin: %v", err)
	}
	app := fiber.New()
	registry.Register(app, zap.NewNop())

	assertStatus(t, app, "/plugins/hello/health", fiber.StatusOK)
	assertStatus(t, app, "/plugins/hello/openapi.json", fiber.StatusOK)
	assertStatus(t, app, "/plugins/other/health", fiber.StatusNotFound)
}

// TestRegistryRejectsUnsafeAndDuplicateDeclarations verifies namespace ownership.
func TestRegistryRejectsUnsafeAndDuplicateDeclarations(t *testing.T) {
	registry := New()
	if err := registry.Mount("../bad", func(fiber.Router) {}); err == nil {
		t.Fatal("expected unsafe namespace rejection")
	}
	if err := registry.Mount("hello", func(fiber.Router) {}); err != nil {
		t.Fatalf("mount route: %v", err)
	}
	if err := registry.Mount("hello", func(fiber.Router) {}); err == nil {
		t.Fatal("expected duplicate mount rejection")
	}
	if err := registry.Describe("hello", []byte("invalid")); err == nil {
		t.Fatal("expected invalid OpenAPI rejection")
	}
}

// TestRegistryRecoversHandlerPanic verifies plugin failures stay isolated.
func TestRegistryRecoversHandlerPanic(t *testing.T) {
	registry := New()
	_ = registry.Mount("panic", func(router fiber.Router) {
		router.Get("/now", func(*fiber.Ctx) error { panic("broken") })
	})
	app := fiber.New()
	registry.Register(app, zap.NewNop())
	assertStatus(t, app, "/plugins/panic/now", fiber.StatusInternalServerError)
}

// assertStatus checks one in-memory Fiber response status.
func assertStatus(t *testing.T, app *fiber.App, path string, expected int) {
	t.Helper()
	response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("request %s: %v", path, err)
	}
	if response.StatusCode != expected {
		t.Fatalf("expected %d for %s, got %d", expected, path, response.StatusCode)
	}
}
