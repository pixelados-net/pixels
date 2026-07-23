package calendar

import (
	"context"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

// rewardNames resolves calendar reward product names.
func rewardNames(ctx context.Context, furniture furnitureservice.DefinitionFinder, definitionID *int64) (string, string) {
	if definitionID == nil || furniture == nil {
		return "", ""
	}
	definition, found, err := furniture.FindDefinitionByID(ctx, *definitionID)
	if err != nil || !found {
		return "", ""
	}

	return definition.Name, definition.Name
}

// missedDays returns past unclaimed campaign days.
func missedDays(current int32, opened []int32) []int32 {
	set := make(map[int32]struct{}, len(opened))
	for _, day := range opened {
		set[day] = struct{}{}
	}
	missed := make([]int32, 0)
	for day := int32(0); day < current; day++ {
		if _, found := set[day]; !found {
			missed = append(missed, day)
		}
	}

	return missed
}

// mapOffer maps one seasonal catalog offer.
func mapOffer(ctx context.Context, catalog *catalogservice.Service, item catalogmodel.Item) (catalogoffer.Offer, error) {
	products := catalog.Products(ctx, item.ID)
	if len(products) == 0 {
		products = []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}
	}
	definitions := make(map[int64]furnituremodel.Definition, len(products))
	for _, product := range products {
		definition, _, err := catalog.Definition(ctx, product.DefinitionID)
		if err != nil {
			return catalogoffer.Offer{}, err
		}
		definitions[product.DefinitionID] = definition
	}

	return catalogprojection.OfferProducts(item, products, definitions)
}

// send sends one encoded calendar packet.
func send(ctx context.Context, connection netconn.Context, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
