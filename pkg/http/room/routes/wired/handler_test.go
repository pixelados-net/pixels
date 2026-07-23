package wired

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestPositiveParamRejectsInvalidIdentifiers verifies HTTP route identifiers fail before persistence access.
func TestPositiveParamRejectsInvalidIdentifiers(t *testing.T) {
	for _, value := range []string{"0", "-1", "invalid"} {
		t.Run(value, func(t *testing.T) {
			app := fiber.New()
			app.Get("/:id", func(ctx *fiber.Ctx) error {
				_, err := positiveParam(ctx, "id")
				return err
			})
			response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/"+value, nil))
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != fiber.StatusBadRequest {
				t.Fatalf("status=%d, want %d", response.StatusCode, fiber.StatusBadRequest)
			}
		})
	}
}

// TestRewardKindAllowlist verifies administration cannot persist unsupported durable reward capabilities.
func TestRewardKindAllowlist(t *testing.T) {
	for _, kind := range []string{"furniture", "badge", "credits", "currency", "respect", "catalog_offer"} {
		if !rewardKind(kind) {
			t.Errorf("implemented reward kind %q was rejected", kind)
		}
	}
	for _, kind := range []string{"", "shell", "admin", "unknown"} {
		if rewardKind(kind) {
			t.Errorf("unsupported reward kind %q was accepted", kind)
		}
	}
}
