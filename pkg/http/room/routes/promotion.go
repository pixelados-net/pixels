package routes

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	roompromotion "github.com/niflaot/pixels/internal/realm/room/promotion"
)

// promotionManager exposes room promotion administration behavior.
type promotionManager interface {
	// Active returns one active promotion.
	Active(context.Context, int64) (roompromotion.Promotion, bool, error)
	// Cancel force-removes one promotion.
	Cancel(context.Context, int64) (bool, error)
}

// PromotionResponse contains safe room promotion state.
type PromotionResponse struct {
	// ID identifies the promotion.
	ID int64 `json:"id"`
	// RoomID identifies the promoted room.
	RoomID int64 `json:"roomId"`
	// CategoryID identifies the visible event category.
	CategoryID int32 `json:"categoryId"`
	// Title stores the event title.
	Title string `json:"title"`
	// Description stores the event description.
	Description string `json:"description"`
	// StartsAt stores the activation boundary.
	StartsAt string `json:"startsAt"`
	// EndsAt stores the expiration boundary.
	EndsAt string `json:"endsAt"`
	// CreatedBy identifies the purchaser.
	CreatedBy int64 `json:"createdBy"`
	// Version stores optimistic mutation order.
	Version int64 `json:"version"`
}

// promotionHandler returns one active room promotion.
func promotionHandler(service promotionManager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := promotionRoomID(ctx)
		if err != nil {
			return err
		}
		value, found, err := service.Active(ctx.Context(), roomID)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "room promotion not found")
		}
		return ctx.JSON(promotionResponse(value))
	}
}

// cancelPromotionHandler force-cancels one active room promotion.
func cancelPromotionHandler(service promotionManager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := promotionRoomID(ctx)
		if err != nil {
			return err
		}
		deleted, err := service.Cancel(ctx.Context(), roomID)
		if err != nil {
			return err
		}
		if !deleted {
			return fiber.NewError(fiber.StatusNotFound, "room promotion not found")
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// promotionRoomID parses one positive room path parameter.
func promotionRoomID(ctx *fiber.Ctx) (int64, error) {
	roomID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || roomID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid room id")
	}
	return roomID, nil
}

// promotionResponse maps durable promotion state.
func promotionResponse(value roompromotion.Promotion) PromotionResponse {
	return PromotionResponse{ID: value.ID, RoomID: value.RoomID, CategoryID: value.CategoryID,
		Title: value.Title, Description: value.Description, StartsAt: value.StartsAt.UTC().Format(time.RFC3339),
		EndsAt: value.EndsAt.UTC().Format(time.RFC3339), CreatedBy: value.CreatedBy, Version: value.Version}
}
