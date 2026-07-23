package routes

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	roompromotion "github.com/niflaot/pixels/internal/realm/room/promotion"
)

// promotionManagerForTest supplies deterministic route behavior.
type promotionManagerForTest struct {
	// value stores the active promotion.
	value roompromotion.Promotion
	// found reports whether the promotion exists.
	found bool
	// canceled records a delete request.
	canceled bool
}

// Active returns deterministic promotion state.
func (manager *promotionManagerForTest) Active(context.Context, int64) (roompromotion.Promotion, bool, error) {
	return manager.value, manager.found, nil
}

// Cancel records a deterministic cancellation.
func (manager *promotionManagerForTest) Cancel(context.Context, int64) (bool, error) {
	manager.canceled = true
	return manager.found, nil
}

// TestPromotionRoutesReadAndCancel verifies admin projection and forced cancellation.
func TestPromotionRoutesReadAndCancel(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	manager := &promotionManagerForTest{found: true, value: roompromotion.Promotion{ID: 1, RoomID: 7, Title: "QA", StartsAt: now, EndsAt: now.Add(time.Hour), Version: 1}}
	app := fiber.New()
	app.Get("/api/admin/rooms/:id/promotion", promotionHandler(manager))
	app.Delete("/api/admin/rooms/:id/promotion", cancelPromotionHandler(manager))
	response, err := app.Test(httptest.NewRequest("GET", "/api/admin/rooms/7/promotion", nil))
	if err != nil {
		t.Fatal(err)
	}
	var body PromotionResponse
	if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK || body.RoomID != 7 || body.Title != "QA" {
		t.Fatalf("status=%d body=%#v", response.StatusCode, body)
	}
	response, err = app.Test(httptest.NewRequest("DELETE", "/api/admin/rooms/7/promotion", nil))
	if err != nil || response.StatusCode != fiber.StatusNoContent || !manager.canceled {
		t.Fatalf("status=%d canceled=%t err=%v", response.StatusCode, manager.canceled, err)
	}
}
