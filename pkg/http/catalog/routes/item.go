package routes

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

// itemsHandler lists active catalog offers.
func itemsHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		pageID, err := optionalPageID(ctx.Query("pageId"))
		if err != nil {
			return err
		}
		items, err := dependencies.Catalog.Items(ctx.Context(), pageID)
		if err != nil {
			return fmt.Errorf("list catalog admin items: %w", err)
		}
		responses := make([]ItemResponse, 0, len(items))
		for _, item := range items {
			responses = append(responses, itemResponse(item))
		}

		return ctx.JSON(responses)
	}
}

// createItemHandler creates one catalog offer.
func createItemHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request ItemRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid catalog item request body")
		}
		item, err := dependencies.Catalog.CreateItem(ctx.Context(), itemInput(request))
		if err != nil {
			return catalogError(err)
		}
		if _, _, err := publishCatalog(ctx.Context(), dependencies); err != nil {
			return err
		}

		return ctx.JSON(itemResponse(item))
	}
}

// updateItemHandler updates one catalog offer.
func updateItemHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := routeID(ctx)
		if err != nil {
			return err
		}
		var request ItemPatchRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid catalog item patch body")
		}
		item, err := dependencies.Catalog.UpdateItem(ctx.Context(), id, itemPatch(request))
		if err != nil {
			return catalogError(err)
		}
		if _, _, err := publishCatalog(ctx.Context(), dependencies); err != nil {
			return err
		}

		return ctx.JSON(itemResponse(item))
	}
}

// deleteItemHandler soft deletes one catalog offer.
func deleteItemHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := routeID(ctx)
		if err != nil {
			return err
		}
		if err := dependencies.Catalog.DeleteItem(ctx.Context(), id); err != nil {
			return catalogError(err)
		}
		if _, _, err := publishCatalog(ctx.Context(), dependencies); err != nil {
			return err
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// optionalPageID parses the optional page filter.
func optionalPageID(raw string) (*int64, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid catalog page id")
	}

	return &id, nil
}

// itemPatch maps an HTTP item patch to administration input.
func itemPatch(request ItemPatchRequest) catalogadmin.ItemPatch {
	var templateRoomID **int64
	var grantsEffectID **int32
	if request.RoomBundleTemplateRoomID != nil {
		templateRoomID = &request.RoomBundleTemplateRoomID
	} else if request.ClearRoomBundleTemplate {
		var cleared *int64
		templateRoomID = &cleared
	}
	if request.GrantsEffectID != nil {
		grantsEffectID = &request.GrantsEffectID
	} else if request.ClearGrantsEffect {
		var cleared *int32
		grantsEffectID = &cleared
	}
	patch := catalogadmin.ItemPatch{PageID: request.PageID, DefinitionID: request.DefinitionID, Name: request.Name,
		RoomBundleTemplateRoomID: templateRoomID,
		GrantsEffectID:           grantsEffectID, GrantsEffectDurationSeconds: request.GrantsEffectDurationSeconds,
		CostCredits: request.CostCredits, CostPoints: request.CostPoints, PointsType: request.PointsType,
		Amount: request.Amount, LimitedStack: request.LimitedStack, BundleDiscountEnabled: request.BundleDiscountEnabled,
		Giftable: request.Giftable, ClubOnly: request.ClubOnly,
		OrderNum: request.OrderNum, Enabled: request.Enabled, ExtraData: request.ExtraData}

	return patch
}

// vouchersHandler lists catalog vouchers.
func vouchersHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		vouchers, err := dependencies.Vouchers.Vouchers(ctx.Context())
		if err != nil {
			return catalogError(err)
		}
		return ctx.JSON(vouchers)
	}
}

// saveVoucherHandler creates or updates one voucher.
func saveVoucherHandler(dependencies Dependencies, update bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request VoucherRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid catalog voucher request")
		}
		id := int64(0)
		if update {
			var err error
			id, err = routeID(ctx)
			if err != nil {
				return err
			}
		}
		voucher, err := dependencies.Vouchers.SaveVoucher(ctx.Context(), catalogmodel.Voucher{ID: id, Code: request.Code,
			CostCredits: request.CostCredits, CostPoints: request.CostPoints, PointsType: request.PointsType,
			CatalogItemID: request.CatalogItemID, RedemptionCap: request.RedemptionCap, PerPlayerCap: request.PerPlayerCap,
			Enabled: request.Enabled, ExpiresAt: request.ExpiresAt})
		if err != nil {
			return catalogError(err)
		}
		return ctx.JSON(voucher)
	}
}

// voucherRedemptionsHandler lists voucher use history.
func voucherRedemptionsHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := routeID(ctx)
		if err != nil {
			return err
		}
		redemptions, err := dependencies.Vouchers.VoucherRedemptions(ctx.Context(), id)
		if err != nil {
			return catalogError(err)
		}
		return ctx.JSON(redemptions)
	}
}
