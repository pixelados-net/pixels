package routes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
)

// pagesHandler lists all active catalog pages.
func pagesHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		pages, err := dependencies.Catalog.Pages(ctx.Context())
		if err != nil {
			return fmt.Errorf("list catalog admin pages: %w", err)
		}
		responses := make([]PageResponse, 0, len(pages))
		for _, page := range pages {
			responses = append(responses, pageResponse(page))
		}

		return ctx.JSON(responses)
	}
}

// createPageHandler creates one catalog page.
func createPageHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request PageRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid catalog page request body")
		}
		page, err := dependencies.Catalog.CreatePage(ctx.Context(), pageInput(request))
		if err != nil {
			return catalogError(err)
		}
		if _, _, err := publishCatalog(ctx.Context(), dependencies); err != nil {
			return err
		}

		return ctx.JSON(pageResponse(page))
	}
}

// updatePageHandler updates one catalog page.
func updatePageHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := routeID(ctx)
		if err != nil {
			return err
		}
		var request PagePatchRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid catalog page patch body")
		}
		page, err := dependencies.Catalog.UpdatePage(ctx.Context(), id, pagePatch(request))
		if err != nil {
			return catalogError(err)
		}
		if _, _, err := publishCatalog(ctx.Context(), dependencies); err != nil {
			return err
		}

		return ctx.JSON(pageResponse(page))
	}
}

// pagePatch maps an HTTP page patch to administration input.
func pagePatch(request PagePatchRequest) catalogadmin.PagePatch {
	patch := catalogadmin.PagePatch{Name: request.Name, Layout: request.Layout, IconColor: request.IconColor,
		IconImage: request.IconImage, OrderNum: request.OrderNum,
		Visible: request.Visible, Enabled: request.Enabled, ClubOnly: request.ClubOnly}
	if request.ParentID != nil {
		parentID := request.ParentID
		patch.ParentID = &parentID
	}
	if request.ClearRequiredNode {
		var node *permission.Node
		patch.RequiredNode = &node
	} else if request.RequiredNode != nil {
		patch.RequiredNode = &request.RequiredNode
	}

	return patch
}
