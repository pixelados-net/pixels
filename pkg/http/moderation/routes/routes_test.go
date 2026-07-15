package routes

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
)

// TestPathID verifies positive identifiers and malformed route values.
func TestPathID(t *testing.T) {
	app := fiber.New()
	app.Get("/:id", func(ctx *fiber.Ctx) error {
		value, err := pathID(ctx, "id")
		if err != nil {
			return err
		}
		return ctx.SendString(string(rune(value)))
	})
	for _, test := range []struct {
		name   string
		path   string
		status int
	}{
		{name: "positive", path: "/7", status: fiber.StatusOK},
		{name: "zero", path: "/0", status: fiber.StatusBadRequest},
		{name: "negative", path: "/-1", status: fiber.StatusBadRequest},
		{name: "malformed", path: "/invalid", status: fiber.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest("GET", test.path, nil))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if response.StatusCode != test.status {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.status)
			}
		})
	}
}

// TestParseLimit verifies default and bounded query limits.
func TestParseLimit(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		_, err := parseLimit(ctx, 25, 100)
		return err
	})
	for _, test := range []struct {
		name   string
		query  string
		status int
	}{
		{name: "default", status: fiber.StatusOK},
		{name: "minimum", query: "?limit=1", status: fiber.StatusOK},
		{name: "maximum", query: "?limit=100", status: fiber.StatusOK},
		{name: "zero", query: "?limit=0", status: fiber.StatusBadRequest},
		{name: "over maximum", query: "?limit=101", status: fiber.StatusBadRequest},
		{name: "malformed", query: "?limit=nope", status: fiber.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest("GET", "/"+test.query, nil))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if response.StatusCode != test.status {
				t.Fatalf("status = %d, want %d", response.StatusCode, test.status)
			}
		})
	}
}

// TestRouteError verifies stable domain-to-HTTP error mappings.
func TestRouteError(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "invalid sanction", err: sanctioncore.ErrInvalidRequest, status: fiber.StatusBadRequest},
		{name: "invalid moderation", err: moderationcore.ErrInvalid, status: fiber.StatusBadRequest},
		{name: "unauthorized", err: sanctioncore.ErrUnauthorized, status: fiber.StatusForbidden},
		{name: "immune", err: sanctioncore.ErrImmune, status: fiber.StatusForbidden},
		{name: "not found", err: moderationcore.ErrNotFound, status: fiber.StatusNotFound},
		{name: "pick conflict", err: moderationcore.ErrPickFailed, status: fiber.StatusConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			var mapped *fiber.Error
			if !errors.As(routeError(test.err), &mapped) {
				t.Fatal("route error did not return a Fiber error")
			}
			if mapped.Code != test.status {
				t.Fatalf("status = %d, want %d", mapped.Code, test.status)
			}
		})
	}
}
