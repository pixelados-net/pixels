package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
)

// PageRequest contains catalog page mutation fields.
type PageRequest struct {
	// ParentID identifies the optional parent page.
	ParentID *int64 `json:"parentId"`
	// Name stores the stable localization slug.
	Name string `json:"name"`
	// Layout identifies the client layout.
	Layout string `json:"layout"`
	// IconColor stores the client icon color.
	IconColor int32 `json:"iconColor"`
	// IconImage stores the client icon image.
	IconImage int32 `json:"iconImage"`
	// MinRank stores the minimum visible rank.
	MinRank int32 `json:"minRank"`
	// OrderNum stores sibling display order.
	OrderNum int32 `json:"orderNum"`
	// Visible reports whether the page appears in the tree.
	Visible bool `json:"visible"`
	// Enabled reports whether the page can be opened.
	Enabled bool `json:"enabled"`
	// ClubOnly reports whether club membership is required.
	ClubOnly bool `json:"clubOnly"`
}

// PagePatchRequest contains optional catalog page mutation fields.
type PagePatchRequest struct {
	// ParentID replaces the optional parent page.
	ParentID *int64 `json:"parentId"`
	// Name replaces the stable localization slug.
	Name *string `json:"name"`
	// Layout replaces the client layout.
	Layout *string `json:"layout"`
	// IconColor replaces the client icon color.
	IconColor *int32 `json:"iconColor"`
	// IconImage replaces the client icon image.
	IconImage *int32 `json:"iconImage"`
	// MinRank replaces the minimum visible rank.
	MinRank *int32 `json:"minRank"`
	// OrderNum replaces sibling display order.
	OrderNum *int32 `json:"orderNum"`
	// Visible replaces page tree visibility.
	Visible *bool `json:"visible"`
	// Enabled replaces page availability.
	Enabled *bool `json:"enabled"`
	// ClubOnly replaces club access policy.
	ClubOnly *bool `json:"clubOnly"`
}

// ItemRequest contains catalog offer mutation fields.
type ItemRequest struct {
	// PageID identifies the owning catalog page.
	PageID int64 `json:"pageId"`
	// DefinitionID identifies the granted furniture definition.
	DefinitionID int64 `json:"definitionId"`
	// Name stores the stable localization slug.
	Name string `json:"name"`
	// CostCredits stores the credits price.
	CostCredits int64 `json:"costCredits"`
	// CostPoints stores the activity-points price.
	CostPoints int64 `json:"costPoints"`
	// PointsType identifies the activity-points currency.
	PointsType int32 `json:"pointsType"`
	// Amount stores the granted furniture quantity.
	Amount int32 `json:"amount"`
	// LimitedStack stores numbered stock.
	LimitedStack int32 `json:"limitedStack"`
	// OfferID stores an optional future grouping id.
	OfferID *int64 `json:"offerId"`
	// ClubOnly reports whether club membership is required.
	ClubOnly bool `json:"clubOnly"`
	// OrderNum stores page display order.
	OrderNum int32 `json:"orderNum"`
	// Enabled reports whether the offer can be purchased.
	Enabled bool `json:"enabled"`
	// ExtraData stores initial furniture state.
	ExtraData string `json:"extraData"`
}

// ItemPatchRequest contains optional catalog offer mutation fields.
type ItemPatchRequest struct {
	// PageID replaces the owning catalog page.
	PageID *int64 `json:"pageId"`
	// DefinitionID replaces the furniture definition.
	DefinitionID *int64 `json:"definitionId"`
	// Name replaces the localization slug.
	Name *string `json:"name"`
	// CostCredits replaces the credits price.
	CostCredits *int64 `json:"costCredits"`
	// CostPoints replaces the activity-points price.
	CostPoints *int64 `json:"costPoints"`
	// PointsType replaces the activity-points currency.
	PointsType *int32 `json:"pointsType"`
	// Amount replaces the furniture quantity.
	Amount *int32 `json:"amount"`
	// LimitedStack replaces numbered stock.
	LimitedStack *int32 `json:"limitedStack"`
	// OfferID replaces the optional future grouping id.
	OfferID *int64 `json:"offerId"`
	// ClubOnly replaces club access policy.
	ClubOnly *bool `json:"clubOnly"`
	// OrderNum replaces page display order.
	OrderNum *int32 `json:"orderNum"`
	// Enabled replaces offer availability.
	Enabled *bool `json:"enabled"`
	// ExtraData replaces initial furniture state.
	ExtraData *string `json:"extraData"`
}

// routeID parses a positive route id.
func routeID(ctx *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid catalog record id")
	}

	return id, nil
}

// pageInput maps an HTTP page request to administration input.
func pageInput(request PageRequest) catalogadmin.PageInput {
	return catalogadmin.PageInput{ParentID: request.ParentID, Name: request.Name, Layout: request.Layout,
		IconColor: request.IconColor, IconImage: request.IconImage, MinRank: request.MinRank, OrderNum: request.OrderNum,
		Visible: request.Visible, Enabled: request.Enabled, ClubOnly: request.ClubOnly}
}

// itemInput maps an HTTP offer request to administration input.
func itemInput(request ItemRequest) catalogadmin.ItemInput {
	return catalogadmin.ItemInput{PageID: request.PageID, DefinitionID: request.DefinitionID, Name: request.Name,
		CostCredits: request.CostCredits, CostPoints: request.CostPoints, PointsType: request.PointsType,
		Amount: request.Amount, LimitedStack: request.LimitedStack, OfferID: request.OfferID, ClubOnly: request.ClubOnly,
		OrderNum: request.OrderNum, Enabled: request.Enabled, ExtraData: request.ExtraData}
}
