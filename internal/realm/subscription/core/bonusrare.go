package core

import (
	"context"
	"math"
)

// BonusRareInfo contains the hotel-view reward and currency progress.
type BonusRareInfo struct {
	// ProductType stores the user-visible furniture name.
	ProductType string
	// ProductClassID stores the Nitro furniture sprite id.
	ProductClassID int32
	// Threshold stores the currency balance required for the reward.
	Threshold int32
	// Remaining stores currency still required, clamped at zero.
	Remaining int32
}

// BonusRareInfo returns progress from the durable configured currency balance.
func (service *Service) BonusRareInfo(ctx context.Context, playerID int64) (BonusRareInfo, error) {
	balance, err := service.currencies.Balance(ctx, playerID, service.options.BonusRareCurrencyType)
	if err != nil {
		return BonusRareInfo{}, err
	}
	threshold := clampProtocolInt(service.options.BonusRareThreshold)
	remaining := service.options.BonusRareThreshold - balance
	if remaining < 0 {
		remaining = 0
	}
	result := BonusRareInfo{Threshold: threshold, Remaining: clampProtocolInt(remaining)}
	if service.options.BonusRareProductID == 0 {
		return result, nil
	}
	definition, found, err := service.furniture.FindDefinitionByID(ctx, int64(service.options.BonusRareProductID))
	if err != nil || !found {
		return result, err
	}
	result.ProductType = definition.PublicName
	if result.ProductType == "" {
		result.ProductType = definition.Name
	}
	result.ProductClassID = int32(definition.SpriteID)
	return result, nil
}

// clampProtocolInt converts a durable non-negative count to protocol range.
func clampProtocolInt(value int64) int32 {
	if value <= 0 {
		return 0
	}
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(value)
}
