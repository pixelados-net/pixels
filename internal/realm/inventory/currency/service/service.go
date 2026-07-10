package service

import (
	"context"
	"errors"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/zap"
)

// Service implements currency balance behavior.
type Service struct {
	// store persists currency balances.
	store currencyrepo.Store

	// catalog validates and lists configured currencies.
	catalog *currency.Catalog

	// events publishes committed currency facts.
	events bus.Publisher

	// log records non-critical projection failures.
	log *zap.Logger
	// permissions resolves infinite balance capability.
	permissions permissionservice.Checker
}

// New creates a currency service.
func New(store currencyrepo.Store, catalog *currency.Catalog, events bus.Publisher, log *zap.Logger, checkers ...permissionservice.Checker) *Service {
	if log == nil {
		log = zap.NewNop()
	}

	service := &Service{store: store, catalog: catalog, events: events, log: log}
	if len(checkers) > 0 {
		service.permissions = checkers[0]
	}

	return service
}

// Wallet returns every configured currency balance for a player.
func (service *Service) Wallet(ctx context.Context, playerID int64) ([]currencymodel.Balance, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}

	stored, err := service.store.ListBalances(ctx, playerID)
	if err != nil {
		return nil, err
	}
	amounts := make(map[int32]currencymodel.Balance, len(stored))
	for _, balance := range stored {
		amounts[balance.CurrencyType] = balance
	}

	definitions := service.catalog.Types()
	wallet := make([]currencymodel.Balance, 0, len(definitions))
	for _, definition := range definitions {
		balance := amounts[definition.Type]
		balance.PlayerID = playerID
		balance.CurrencyType = definition.Type
		wallet = append(wallet, balance)
	}

	return wallet, nil
}

// Balance returns one configured currency balance.
func (service *Service) Balance(ctx context.Context, playerID int64, currencyType int32) (int64, error) {
	if err := service.validateIdentity(playerID, currencyType); err != nil {
		return 0, err
	}

	balance, found, err := service.store.FindBalance(ctx, playerID, currencyType)
	if err != nil || !found {
		return 0, err
	}

	return balance.Amount, nil
}

// Grant applies a signed currency delta.
func (service *Service) Grant(ctx context.Context, params GrantParams) (int64, error) {
	definition, err := service.validateMutation(params.PlayerID, params.CurrencyType, params.Amount, params.ActorKind, false)
	if err != nil {
		return 0, err
	}
	infinite, err := service.infiniteBalance(ctx, params)
	if err != nil {
		return 0, err
	}
	if infinite {
		balance, found, err := service.store.FindBalance(ctx, params.PlayerID, params.CurrencyType)
		if err != nil || !found {
			return 0, err
		}

		return balance.Amount, nil
	}

	result, err := service.store.Grant(ctx, mutation(params, definition.Ledger))
	if errors.Is(err, currencyrepo.ErrInsufficientBalance) {
		return 0, ErrInsufficientBalance
	}
	if errors.Is(err, currencyrepo.ErrBalanceOverflow) {
		return 0, ErrInvalidAmount
	}
	if err != nil {
		return 0, err
	}
	service.publish(ctx, params.PlayerID, params.CurrencyType, result.Balance.Amount, result.Delta, params.ActorKind)

	return result.Balance.Amount, nil
}

// infiniteBalance reports whether a player-originated deduction bypasses persistence.
func (service *Service) infiniteBalance(ctx context.Context, params GrantParams) (bool, error) {
	if service.permissions == nil || params.ActorKind != ActorPlayer || params.Amount >= 0 {
		return false, nil
	}

	return service.permissions.HasPermission(ctx, params.PlayerID, currency.InfiniteBalance)
}

// Set replaces a currency balance with an absolute amount.
func (service *Service) Set(ctx context.Context, params SetParams) (int64, error) {
	definition, err := service.validateMutation(params.PlayerID, params.CurrencyType, params.Amount, params.ActorKind, true)
	if err != nil {
		return 0, err
	}

	result, err := service.store.Set(ctx, setMutation(params, definition.Ledger))
	if err != nil {
		return 0, err
	}
	service.publish(ctx, params.PlayerID, params.CurrencyType, result.Balance.Amount, result.Delta, params.ActorKind)

	return result.Balance.Amount, nil
}

// Types returns configured currency definitions.
func (service *Service) Types(context.Context) ([]currencymodel.Definition, error) {
	return service.catalog.Types(), nil
}

// validateIdentity validates a player and configured currency type.
func (service *Service) validateIdentity(playerID int64, currencyType int32) error {
	if playerID <= 0 {
		return ErrInvalidPlayerID
	}
	if _, found := service.catalog.Type(currencyType); !found {
		return ErrInvalidCurrencyType
	}

	return nil
}

// validateMutation validates shared mutation fields.
func (service *Service) validateMutation(playerID int64, currencyType int32, amount int64, actor string, absolute bool) (currencymodel.Definition, error) {
	if err := service.validateIdentity(playerID, currencyType); err != nil {
		return currencymodel.Definition{}, err
	}
	if (!absolute && amount == 0) || (absolute && amount < 0) {
		return currencymodel.Definition{}, ErrInvalidAmount
	}
	if actor != ActorSystem && actor != ActorAdmin && actor != ActorPlayer {
		return currencymodel.Definition{}, ErrInvalidActor
	}

	definition, _ := service.catalog.Type(currencyType)

	return definition, nil
}

// mutation maps grant parameters into repository data.
func mutation(params GrantParams, ledger bool) currencyrepo.Mutation {
	return currencyrepo.Mutation{
		PlayerID: params.PlayerID, CurrencyType: params.CurrencyType, Amount: params.Amount,
		Ledger: ledger, Reason: params.Reason, ActorKind: params.ActorKind, ActorID: params.ActorID,
	}
}

// setMutation maps set parameters into repository data.
func setMutation(params SetParams, ledger bool) currencyrepo.Mutation {
	return currencyrepo.Mutation{
		PlayerID: params.PlayerID, CurrencyType: params.CurrencyType, Amount: params.Amount,
		Ledger: ledger, Reason: params.Reason, ActorKind: params.ActorKind, ActorID: params.ActorID,
	}
}

// publish emits a committed currency change without rolling back persistence on projection failure.
func (service *Service) publish(ctx context.Context, playerID int64, currencyType int32, amount int64, delta int64, actorKind string) {
	if postgres.AfterCommit(ctx, func(commitCtx context.Context) {
		service.publishNow(commitCtx, playerID, currencyType, amount, delta, actorKind)
	}) {
		return
	}

	service.publishNow(ctx, playerID, currencyType, amount, delta, actorKind)
}

// publishNow emits one committed currency fact.
func (service *Service) publishNow(ctx context.Context, playerID int64, currencyType int32, amount int64, delta int64, actorKind string) {
	if service.events == nil {
		return
	}

	err := service.events.Publish(ctx, bus.Event{
		Name: currencychanged.Name,
		Payload: currencychanged.Payload{
			PlayerID: playerID, CurrencyType: currencyType, Amount: amount, Delta: delta, ActorKind: actorKind,
		},
	})
	if err != nil {
		service.log.Warn("currency change projection failed",
			zap.Int64("player_id", playerID),
			zap.Int32("currency_type", currencyType),
			zap.Int64("amount", amount),
			zap.Int64("delta", delta),
			zap.Error(err),
		)
	}
}
