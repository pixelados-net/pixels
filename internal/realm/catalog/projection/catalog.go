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
	if !fitsInt32(item.ID) || !fitsInt32(item.CostCredits) || !fitsInt32(item.CostPoints) || definition.SpriteID < 0 || definition.SpriteID > math.MaxInt32 {
		return offer.Offer{}, ErrProtocolRange
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
	remaining := item.LimitedStack - item.LimitedSells
	if remaining < 0 {
		remaining = 0
	}

	return offer.Offer{
		ID: int32(item.ID), LocalizationID: "catalog.item." + item.Name,
		CostCredits: int32(item.CostCredits), CostPoints: int32(item.CostPoints), PointsType: item.PointsType,
		Giftable: false, ClubLevel: clubLevel(item.ClubOnly), BundlePurchaseAllowed: false,
		Products: []offer.Product{{Type: productType, ClassID: int32(definition.SpriteID), ExtraData: item.ExtraData,
			Amount: item.Amount, Limited: item.IsLimited(), LimitedStack: item.LimitedStack, LimitedRemaining: remaining}},
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
