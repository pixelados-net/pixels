package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	"go.uber.org/zap"
)

// TestAccessEnforcesNamespaceAndDisablesPanickingHandler verifies route isolation.
func TestAccessEnforcesNamespaceAndDisablesPanickingHandler(t *testing.T) {
	registry := pluginroutes.New()
	scope := pluginruntime.NewScope("demo")
	access := NewAccess(registry, scope)
	if err := access.Mount("other", func(fiber.Router) {}); err != pluginruntime.ErrWrongNamespace {
		t.Fatalf("expected namespace rejection, got %v", err)
	}
	if err := access.Describe("other", []byte(`{}`)); err != pluginruntime.ErrWrongNamespace {
		t.Fatalf("expected spec namespace rejection, got %v", err)
	}
	if err := access.Describe("demo", []byte(`{"openapi":"3.0.3"}`)); err != nil {
		t.Fatalf("describe plugin: %v", err)
	}
	if err := access.Mount("demo", func(router fiber.Router) {
		router.Get("/panic", func(*fiber.Ctx) error { panic("boom") })
	}); err != nil {
		t.Fatalf("mount plugin: %v", err)
	}
	app := fiber.New()
	registry.Register(app, zap.NewNop())
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/plugins/demo/panic", nil))
	if err != nil || response.StatusCode != fiber.StatusInternalServerError || scope.Enabled() {
		t.Fatalf("expected isolated route panic, status=%d enabled=%v err=%v", response.StatusCode, scope.Enabled(), err)
	}
}

// TestAccessDisablesPanickingMount verifies startup mount failures disable the scope.
func TestAccessDisablesPanickingMount(t *testing.T) {
	registry := pluginroutes.New()
	scope := pluginruntime.NewScope("mount-panic")
	access := NewAccess(registry, scope)
	if err := access.Mount("mount-panic", func(fiber.Router) { panic("mount") }); err != nil {
		t.Fatalf("store mount: %v", err)
	}
	registry.Register(fiber.New(), zap.NewNop())
	if scope.Enabled() {
		t.Fatal("expected panicking mount scope disabled")
	}
}
