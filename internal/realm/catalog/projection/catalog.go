// Package projection maps catalog realm records to protocol response models.
package projection

import (
	"errors"
	"math"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
	outpages "github.com/niflaot/pixels/networking/outbound/catalog/pages"
	"github.com/niflaot/pixels/pkg/i18n"
)

var (
	// ErrProtocolRange reports durable catalog values outside int32 packet range.
	ErrProtocolRange = errors.New("catalog value exceeds protocol range")
	// ErrUnsupportedFurniture reports a furniture kind unsupported by catalog packets.
	ErrUnsupportedFurniture = errors.New("unsupported catalog furniture kind")
)

const productTypeEffect = "e"

// PageTree maps ordered pages into recursive Nitro catalog nodes.
func PageTree(pages []catalogmodel.Page, translations i18n.Translator) ([]outpages.Node, error) {
	children := make(map[int64][]catalogmodel.Page)
	roots := make([]catalogmodel.Page, 0)
	for _, page := range pages {
		if page.ParentID == nil {
			roots = append(roots, page)
			continue
		}
		children[*page.ParentID] = append(children[*page.ParentID], page)
	}

	return mapNodes(roots, children, translations)
}

// Offer maps one catalog item and furniture definition to Nitro offer data.
func Offer(item catalogmodel.Item, definition furnituremodel.Definition) (offer.Offer, error) {
	return OfferProducts(item, []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}, map[int64]furnituremodel.Definition{item.DefinitionID: definition})
}

// OfferProducts maps one catalog item and its products to Nitro offer data.
func OfferProducts(item catalogmodel.Item, products []catalogmodel.Product, definitions map[int64]furnituremodel.Definition) (offer.Offer, error) {
	if !fitsInt32(item.ID) || !fitsInt32(item.CostCredits) || !fitsInt32(item.CostPoints) {
		return offer.Offer{}, ErrProtocolRange
	}
	remaining := item.LimitedStack - item.LimitedSells
	if remaining < 0 {
		remaining = 0
	}

	mapped := make([]offer.Product, 0, len(products)+1)
	for _, product := range products {
		definition, found := definitions[product.DefinitionID]
		if !found || definition.SpriteID < 0 || definition.SpriteID > math.MaxInt32 {
			return offer.Offer{}, ErrUnsupportedFurniture
		}
		productType := ""
		if definition.Kind == furnituremodel.KindFloor {
			productType = "s"
		}
		if definition.Kind == furnituremodel.KindWall {
			productType = "i"
		}
		if productType == "" {
			return offer.Offer{}, ErrUnsupportedFurniture
		}
		mapped = append(mapped, offer.Product{Type: productType, ClassID: int32(definition.SpriteID), ExtraData: item.ExtraData, Amount: product.Quantity, Limited: item.IsLimited(), LimitedStack: item.LimitedStack, LimitedRemaining: remaining})
	}
	if item.GrantsEffectID != nil {
		mapped = append(mapped, offer.Product{Type: productTypeEffect, ClassID: *item.GrantsEffectID, Amount: 1, Limited: item.IsLimited(), LimitedStack: item.LimitedStack, LimitedRemaining: remaining})
	}
	return offer.Offer{
		ID: int32(item.ID), LocalizationID: "catalog.item." + item.Name,
		CostCredits: int32(item.CostCredits), CostPoints: int32(item.CostPoints), PointsType: item.PointsType,
		Giftable: item.Giftable, ClubLevel: clubLevel(item.ClubOnly), BundlePurchaseAllowed: item.BundleDiscountEnabled,
		Products: mapped,
	}, nil
}

// mapNodes recursively maps sibling pages.
func mapNodes(pages []catalogmodel.Page, children map[int64][]catalogmodel.Page, translations i18n.Translator) ([]outpages.Node, error) {
	nodes := make([]outpages.Node, 0, len(pages))
	for _, page := range pages {
		if !fitsInt32(page.ID) {
			return nil, ErrProtocolRange
		}
		nested, err := mapNodes(children[page.ID], children, translations)
		if err != nil {
			return nil, err
		}
		localization := "catalog.page." + page.Name
		if translations != nil {
			localization = translations.Default(i18n.Key(localization))
		}
		nodes = append(nodes, outpages.Node{Visible: page.Visible, IconImage: page.IconImage, PageID: int32(page.ID),
			Name: page.Name, Localization: localization, Children: nested})
	}

	return nodes, nil
}

// fitsInt32 reports whether a signed integer fits the packet primitive.
func fitsInt32(value int64) bool {
	return value >= math.MinInt32 && value <= math.MaxInt32
}

// clubLevel maps catalog club policy to Nitro's numeric level.
func clubLevel(required bool) int32 {
	if required {
		return 1
	}

	return 0
}
