package rentable

import (
	"context"
	"errors"
	"time"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
)

var (
	// ErrUnavailable reports a missing or occupied rentable target.
	ErrUnavailable = errors.New("rentable furniture unavailable")
	// ErrNotRenter reports a mutation by someone other than the active renter.
	ErrNotRenter = errors.New("active furniture renter required")
)

// Service coordinates atomic rental charging and persistence.
type Service struct {
	store    Store
	currency currencyservice.Granter
	config   Config
	now      func() time.Time
}

// New creates rentable furniture behavior.
func New(config Config, store Store, currency currencyservice.Granter) *Service {
	return &Service{store: store, currency: currency, config: config.Normalize(), now: time.Now}
}

// Status returns one room's current rentable-space state.
func (service *Service) Status(ctx context.Context, roomID int64) (State, bool, error) {
	return service.store.FindRoomSpace(ctx, roomID)
}

// Rent starts or extends a rentable space atomically.
func (service *Service) Rent(ctx context.Context, itemID int64, playerID int64) (State, error) {
	var state State
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		current, found, findErr := service.store.FindItem(txCtx, itemID)
		if findErr != nil {
			return findErr
		}
		if !found || current.ActiveAt(service.now()) && (current.RenterPlayerID == nil || *current.RenterPlayerID != playerID) {
			return ErrUnavailable
		}
		if service.currency != nil && service.config.PriceCredits > 0 {
			if _, grantErr := service.currency.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: -1, Amount: -int64(service.config.PriceCredits), Reason: "rentable furniture", ActorKind: currencyservice.ActorPlayer, ActorID: &playerID}); grantErr != nil {
				return grantErr
			}
		}
		var changed bool
		state, changed, findErr = service.store.Rent(txCtx, itemID, playerID, service.config.PriceCredits, int64(service.config.Duration/time.Second))
		if findErr != nil {
			return findErr
		}
		if !changed {
			return ErrUnavailable
		}
		return nil
	})
	return state, err
}

// Cancel clears one active rental without refunding consumed time.
func (service *Service) Cancel(ctx context.Context, itemID int64, playerID int64) error {
	changed, err := service.store.Cancel(ctx, itemID, playerID)
	if err == nil && !changed {
		return ErrNotRenter
	}
	return err
}

// Buyout charges and transfers one actively rented furniture instance.
func (service *Service) Buyout(ctx context.Context, itemID int64, playerID int64) error {
	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		state, found, err := service.store.FindItem(txCtx, itemID)
		if err != nil {
			return err
		}
		if !found || !state.ActiveAt(service.now()) || state.RenterPlayerID == nil || *state.RenterPlayerID != playerID {
			return ErrNotRenter
		}
		if service.currency != nil && service.config.BuyoutCredits > 0 {
			if _, err = service.currency.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: -1, Amount: -int64(service.config.BuyoutCredits), Reason: "rentable furniture buyout", ActorKind: currencyservice.ActorPlayer, ActorID: &playerID}); err != nil {
				return err
			}
		}
		changed, err := service.store.Buyout(txCtx, itemID, playerID)
		if err == nil && !changed {
			return ErrNotRenter
		}
		return err
	})
}

// Config returns normalized prices for protocol offers.
func (service *Service) Config() Config { return service.config }
