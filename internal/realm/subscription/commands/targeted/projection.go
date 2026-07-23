package targeted

import (
	"context"
	"math"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

// mapCatalogOffer maps one catalog offer and its products.
func (handler Handler) mapCatalogOffer(ctx context.Context, item catalogmodel.Item) (catalogoffer.Offer, error) {
	products := handler.Catalog.Products(ctx, item.ID)
	if len(products) == 0 {
		products = []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}
	}
	definitions := make(map[int64]furnituremodel.Definition, len(products))
	for _, product := range products {
		definition, _, err := handler.Catalog.Definition(ctx, product.DefinitionID)
		if err != nil {
			return catalogoffer.Offer{}, err
		}
		definitions[product.DefinitionID] = definition
	}

	return catalogprojection.OfferProducts(item, products, definitions)
}

// clampInt32 clamps durable prices to protocol range.
func clampInt32(value int64) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}

	return int32(value)
}

// send sends one encoded targeted offer packet.
func send(ctx context.Context, connection netconn.Context, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
