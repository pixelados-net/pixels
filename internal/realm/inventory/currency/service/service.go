package service

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
)

// Service implements currency balance behavior.
type Service struct {
	// store persists currency balances.
	store currencyrepo.Store

	// catalog validates and lists configured currencies.
	catalog *currency.Catalog
}

// New creates a currency service.
func New(store currencyrepo.Store, catalog *currency.Catalog) *Service {
	return &Service{store: store, catalog: catalog}
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

	balance, err := service.store.Grant(ctx, mutation(params, definition.Ledger))
	if errors.Is(err, currencyrepo.ErrInsufficientBalance) {
		return 0, ErrInsufficientBalance
	}
	if errors.Is(err, currencyrepo.ErrBalanceOverflow) {
		return 0, ErrInvalidAmount
	}
	if err != nil {
		return 0, err
	}

	return balance.Amount, nil
}

// Set replaces a currency balance with an absolute amount.
func (service *Service) Set(ctx context.Context, params SetParams) (int64, error) {
	definition, err := service.validateMutation(params.PlayerID, params.CurrencyType, params.Amount, params.ActorKind, true)
	if err != nil {
		return 0, err
	}

	balance, err := service.store.Set(ctx, setMutation(params, definition.Ledger))
	if err != nil {
		return 0, err
	}

	return balance.Amount, nil
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
