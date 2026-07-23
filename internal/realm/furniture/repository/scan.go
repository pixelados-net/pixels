package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// scanDefinition scans a furniture definition row.
func scanDefinition(row pgx.Row) (furnituremodel.Definition, error) {
	var definition furnituremodel.Definition
	var kind string
	var metadata []byte
	var effectMale pgtype.Int4
	var effectFemale pgtype.Int4
	var deletedAt pgtype.Timestamptz
	err := row.Scan(
		&definition.ID,
		&definition.SpriteID,
		&definition.Name,
		&definition.PublicName,
		&definition.Description,
		&kind,
		&definition.Width,
		&definition.Length,
		&definition.StackHeight,
		&definition.AllowStack,
		&definition.AllowWalk,
		&definition.AllowSit,
		&definition.AllowLay,
		&definition.AllowInventoryStack,
		&definition.AllowTrade,
		&definition.AllowMarketplaceSale,
		&definition.AllowRecycle,
		&definition.RedeemableCredits,
		&definition.EffectPool,
		&effectMale,
		&effectFemale,
		&definition.InteractionType,
		&definition.InteractionModesCount,
		&definition.Multiheight,
		&definition.CustomParams,
		&metadata,
		&definition.CreatedAt,
		&definition.UpdatedAt,
		&deletedAt,
		&definition.Version.Version,
	)
	if err != nil {
		return furnituremodel.Definition{}, err
	}
	definition.Kind = furnituremodel.Kind(kind)
	definition.Metadata = json.RawMessage(metadata)
	if effectMale.Valid {
		definition.EffectMale = &effectMale.Int32
	}
	if effectFemale.Valid {
		definition.EffectFemale = &effectFemale.Int32
	}
	definition.DeletedAt = timePointer(deletedAt)

	return definition, nil
}

// scanDefinitions scans furniture definition rows.
func scanDefinitions(rows pgx.Rows) ([]furnituremodel.Definition, error) {
	var definitions []furnituremodel.Definition
	for rows.Next() {
		definition, err := scanDefinition(rows)
		if err != nil {
			return nil, fmt.Errorf("scan furniture definition: %w", err)
		}
		definitions = append(definitions, definition)
	}

	return definitions, rows.Err()
}

// scanItem scans a furniture item row.
func scanItem(row pgx.Row) (furnituremodel.Item, error) {
	var item furnituremodel.Item
	var roomID pgtype.Int8
	var x pgtype.Int2
	var y pgtype.Int2
	var z pgtype.Float8
	var rotation int16
	var wallPosition pgtype.Text
	var rentalOwner pgtype.Int8
	var rentalExpiry pgtype.Timestamptz
	var rentalPrice pgtype.Int4
	var stackHeight pgtype.Int4
	var limitedEditionNumber pgtype.Int4
	var giftSprite pgtype.Int4
	var giftBox pgtype.Int4
	var giftRibbon pgtype.Int4
	var giftSender pgtype.Int8
	var giftMessage pgtype.Text
	var metadata []byte
	var deletedAt pgtype.Timestamptz
	err := row.Scan(
		&item.ID,
		&item.DefinitionID,
		&item.OwnerPlayerID,
		&roomID,
		&x,
		&y,
		&z,
		&rotation,
		&wallPosition,
		&item.ExtraData,
		&rentalOwner,
		&rentalExpiry,
		&rentalPrice,
		&stackHeight,
		&limitedEditionNumber,
		&item.MarketplaceReserved,
		&item.GiftWrapped,
		&giftSprite,
		&giftBox,
		&giftRibbon,
		&giftSender,
		&giftMessage,
		&metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
		&item.Version.Version,
	)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	item.RoomID = int64Pointer(roomID)
	item.RentalOwnerPlayerID = int64Pointer(rentalOwner)
	item.RentalExpiresAt = timePointer(rentalExpiry)
	if rentalPrice.Valid {
		item.RentalPriceCredits = &rentalPrice.Int32
	}
	if stackHeight.Valid {
		item.StackHeightOverrideCM = &stackHeight.Int32
	}
	if limitedEditionNumber.Valid {
		item.LimitedEditionNumber = &limitedEditionNumber.Int32
	}
	if giftSprite.Valid {
		item.GiftWrapSpriteID = &giftSprite.Int32
	}
	if giftBox.Valid {
		item.GiftWrapBoxID = &giftBox.Int32
	}
	if giftRibbon.Valid {
		item.GiftWrapRibbonID = &giftRibbon.Int32
	}
	item.GiftSenderPlayerID = int64Pointer(giftSender)
	if giftMessage.Valid {
		item.GiftMessage = &giftMessage.String
	}
	item.X = int2Pointer(x)
	item.Y = int2Pointer(y)
	item.Z = float64Pointer(z)
	item.Rotation = furnituremodel.Rotation(rotation)
	item.WallPosition = stringPointer(wallPosition)
	item.Metadata = json.RawMessage(metadata)
	item.DeletedAt = timePointer(deletedAt)

	return item, nil
}

// scanItems scans furniture item rows.
func scanItems(rows pgx.Rows) ([]furnituremodel.Item, error) {
	var items []furnituremodel.Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan furniture item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// timePointer converts a PostgreSQL timestamp to an optional time.
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

// int64Pointer converts a PostgreSQL int8 to an optional int64.
func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}

// int2Pointer converts a PostgreSQL int2 to an optional int.
func int2Pointer(value pgtype.Int2) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int16)

	return &result
}

// float64Pointer converts a PostgreSQL float8 to an optional float64.
func float64Pointer(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}

// stringPointer converts a PostgreSQL text value to an optional string.
func stringPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}
