package service

import (
	"context"
	"errors"
	"github.com/niflaot/pixels/internal/realm/player/repository"
)

// ErrTradeWriterUnavailable reports missing durable trade eligibility persistence.
var ErrTradeWriterUnavailable = errors.New("player trade writer unavailable")

// TradeManager changes durable and live player trade eligibility.
type TradeManager interface {
	SetAllowTrade(context.Context, int64, bool) error
}

// SetAllowTrade updates durable player trade eligibility.
func (service *Service) SetAllowTrade(ctx context.Context, playerID int64, allow bool) error {
	writer, ok := service.store.(repository.TradeWriter)
	if !ok {
		return ErrTradeWriterUnavailable
	}
	updated, err := writer.UpdateAllowTrade(ctx, playerID, allow)
	if err != nil {
		return err
	}
	if !updated {
		return ErrPlayerNotFound
	}
	return nil
}

var _ TradeManager = (*Service)(nil)
