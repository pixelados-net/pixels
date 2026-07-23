package routes

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
)

// checkerStub records one route authorization decision.
type checkerStub struct {
	// allowed stores the returned permission decision.
	allowed bool
	// playerID stores the resolved actor.
	playerID int64
	// node stores the resolved permission node.
	node permission.Node
}

// HasPermission records and returns one configured permission decision.
func (checker *checkerStub) HasPermission(_ context.Context, playerID int64, node permission.Node) (bool, error) {
	checker.playerID = playerID
	checker.node = node
	return checker.allowed, nil
}

// TestAuthorizeRead validates actor attribution and permission enforcement.
func TestAuthorizeRead(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		allowed    bool
		wantStatus int
	}{
		{name: "missing actor", wantStatus: fiber.StatusBadRequest},
		{name: "denied actor", header: "7", wantStatus: fiber.StatusForbidden},
		{name: "allowed actor", header: "7", allowed: true, wantStatus: fiber.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checker := &checkerStub{allowed: test.allowed}
			dependencies := Dependencies{Permissions: checker}
			app := fiber.New()
			app.Get("/authorization", func(ctx *fiber.Ctx) error {
				if err := dependencies.authorizeRead(ctx, petpolicy.ManageAny); err != nil {
					return err
				}
				return ctx.SendStatus(fiber.StatusNoContent)
			})
			request := httptest.NewRequest(fiber.MethodGet, "/authorization", nil)
			if test.header != "" {
				request.Header.Set(actorHeader, test.header)
			}
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("test request: %v", err)
			}
			if response.StatusCode != test.wantStatus {
				t.Fatalf("expected status %d, got %d", test.wantStatus, response.StatusCode)
			}
			if test.header != "" && checker.playerID != 7 {
				t.Fatalf("expected actor 7, got %d", checker.playerID)
			}
			if test.header != "" && checker.node != petpolicy.ManageAny {
				t.Fatalf("expected manage node, got %q", checker.node)
			}
		})
	}
}
