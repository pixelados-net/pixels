package routes

import (
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

// BalanceResponse contains one player currency balance.
type BalanceResponse struct {
	// CurrencyType identifies the protocol currency.
	CurrencyType int32 `json:"currencyType"`

	// Amount stores the absolute balance.
	Amount int64 `json:"amount"`
}

// WalletResponse contains one player's complete configured wallet.
type WalletResponse struct {
	// PlayerID identifies the wallet owner.
	PlayerID int64 `json:"playerId"`

	// Balances stores configured currency balances.
	Balances []BalanceResponse `json:"balances"`
}

// TypeResponse contains one configured currency definition.
type TypeResponse struct {
	// Type identifies the protocol currency.
	Type int32 `json:"type"`

	// Key identifies the localized currency name.
	Key string `json:"key"`

	// Ledger reports whether mutations are audited.
	Ledger bool `json:"ledger"`

	// Color stores an optional future client presentation color.
	Color string `json:"color,omitempty"`
}

// TypesResponse contains configured currency definitions.
type TypesResponse struct {
	// Types stores configured currency definitions.
	Types []TypeResponse `json:"types"`
}

// MutationResponse contains one committed administrative mutation.
type MutationResponse struct {
	// PlayerID identifies the affected player.
	PlayerID int64 `json:"playerId"`

	// CurrencyType identifies the affected currency.
	CurrencyType int32 `json:"currencyType"`

	// Amount stores the resulting absolute balance.
	Amount int64 `json:"amount"`

	// AlertRequested reports whether alert delivery was requested.
	AlertRequested bool `json:"alertRequested"`

	// AlertSent reports whether the optional alert reached a live connection.
	AlertSent bool `json:"alertSent"`
}

// balanceResponses maps persistent balances to HTTP responses.
func balanceResponses(balances []currencymodel.Balance) []BalanceResponse {
	responses := make([]BalanceResponse, 0, len(balances))
	for _, balance := range balances {
		responses = append(responses, BalanceResponse{CurrencyType: balance.CurrencyType, Amount: balance.Amount})
	}

	return responses
}

// typeResponses maps configured definitions to HTTP responses.
func typeResponses(definitions []currencymodel.Definition) []TypeResponse {
	responses := make([]TypeResponse, 0, len(definitions))
	for _, definition := range definitions {
		responses = append(responses, TypeResponse{
			Type: definition.Type, Key: definition.Key, Ledger: definition.Ledger, Color: definition.Color,
		})
	}

	return responses
}
