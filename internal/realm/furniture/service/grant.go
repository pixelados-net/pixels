package service

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

// GrantParams contains input for creating player inventory items.
type GrantParams struct {
	// DefinitionID identifies the furniture definition.
	DefinitionID int64
	// OwnerPlayerID identifies the receiving player.
	OwnerPlayerID int64
	// Quantity stores the number of instances to create.
	Quantity int32
	// ExtraData stores the initial protocol-facing state.
	ExtraData string
	// LimitedEditionNumber stores an optional durable LTD serial number.
	LimitedEditionNumber *int32
}

// GiftGrantParams contains wrapped furniture grant input.
type GiftGrantParams struct {
	// GrantParams contains common grant input.
	GrantParams
	// SpriteID stores the selected wrapping furniture sprite.
	SpriteID int32
	// BoxID stores the selected wrapping box.
	BoxID int32
	// RibbonID stores the selected ribbon.
	RibbonID int32
	// SenderPlayerID optionally identifies the visible sender.
	SenderPlayerID *int64
	// Message stores the gift message.
	Message string
}

// Grant creates inventory items for a player from one definition.
func (service *Service) Grant(ctx context.Context, params GrantParams) ([]furnituremodel.Item, error) {
	if params.DefinitionID <= 0 {
		return nil, ErrInvalidDefinitionID
	}
	if params.OwnerPlayerID <= 0 {
		return nil, ErrInvalidPlayerID
	}
	if params.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	_, found, err := service.store.FindDefinitionByID(ctx, params.DefinitionID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrDefinitionNotFound
	}

	return service.store.CreateItems(ctx, params.DefinitionID, params.OwnerPlayerID, params.Quantity, params.ExtraData, params.LimitedEditionNumber)
}

// GrantGift creates wrapped inventory items for one recipient.
func (service *Service) GrantGift(ctx context.Context, params GiftGrantParams) ([]furnituremodel.Item, error) {
	if params.DefinitionID <= 0 {
		return nil, ErrInvalidDefinitionID
	}
	if params.OwnerPlayerID <= 0 {
		return nil, ErrInvalidPlayerID
	}
	if params.Quantity <= 0 || params.SpriteID <= 0 || len(params.Message) > 255 {
		return nil, ErrInvalidQuantity
	}
	_, found, err := service.store.FindDefinitionByID(ctx, params.DefinitionID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrDefinitionNotFound
	}
	writer, ok := service.store.(repository.GiftItemWriter)
	if !ok {
		return nil, ErrGiftWriterUnavailable
	}

	return writer.CreateGiftItems(ctx, repository.GiftItemParams{
		DefinitionID:   params.DefinitionID,
		OwnerPlayerID:  params.OwnerPlayerID,
		Quantity:       params.Quantity,
		ExtraData:      params.ExtraData,
		SpriteID:       params.SpriteID,
		BoxID:          params.BoxID,
		RibbonID:       params.RibbonID,
		SenderPlayerID: params.SenderPlayerID,
		Message:        params.Message,
	})
}
