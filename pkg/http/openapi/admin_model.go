package openapi

// ConnectionCountRequest contains optional count filters.
type ConnectionCountRequest struct {
	APIKeyRequest
	// Kind stores the optional connection kind filter.
	Kind string `query:"kind"`
}

// ConnectionListRequest contains optional list filters.
type ConnectionListRequest struct {
	APIKeyRequest
	// Kind stores the optional connection kind filter.
	Kind string `query:"kind"`
}

// ConnectionCountResponse contains connection count data.
type ConnectionCountResponse struct {
	// Total stores the total active connection count.
	Total int `json:"total" required:"true"`
	// Kind stores the optional filtered connection kind.
	Kind string `json:"kind,omitempty"`
	// Count stores the optional filtered kind count.
	Count *int `json:"count,omitempty"`
}

// ConnectionListResponse contains connection list data.
type ConnectionListResponse struct {
	// Total stores the returned connection count.
	Total int `json:"total" required:"true"`
	// Items stores safe connection rows.
	Items []ConnectionResponse `json:"items" required:"true"`
}

// ConnectionResponse contains safe connection data.
type ConnectionResponse struct {
	// ID stores the connection id.
	ID string `json:"id" required:"true"`
	// Kind stores the connection kind.
	Kind string `json:"kind" required:"true" example:"websocket"`
	// State stores the lifecycle state label.
	State string `json:"state" required:"true" example:"connected"`
	// StartedAt stores the session start time.
	StartedAt string `json:"startedAt" required:"true" format:"date-time"`
	// AuthenticatedAt stores the optional authentication time.
	AuthenticatedAt string `json:"authenticatedAt,omitempty" format:"date-time"`
}

// ReasonsResponse contains supported disconnect reasons.
type ReasonsResponse struct {
	// Items stores supported reason rows.
	Items []ReasonResponse `json:"items" required:"true"`
}

// ReasonResponse contains one supported disconnect reason.
type ReasonResponse struct {
	// Code stores the numeric disconnect code.
	Code uint16 `json:"code" required:"true"`
	// Reason stores the stable disconnect reason label.
	Reason string `json:"reason" required:"true"`
}

// NotificationRequest contains one localized notification request.
type NotificationRequest struct {
	APIKeyRequest
	// PlayerID stores the target player id.
	PlayerID int64 `json:"playerId" required:"true" minimum:"1"`
	// Kind stores the notification kind.
	Kind string `json:"kind" enum:"bubble,alert" default:"bubble"`
	// Key stores the required i18n message key.
	Key string `json:"key" required:"true" example:"admin.notification.default"`
	// Locale stores an optional locale override.
	Locale string `json:"locale,omitempty" example:"es"`
	// BubbleKey stores an optional bubble alert type.
	BubbleKey string `json:"bubbleKey,omitempty" example:"admin.notification"`
	// Params stores optional translation parameters.
	Params map[string]string `json:"params,omitempty"`
}

// NotificationResponse contains delivery status.
type NotificationResponse struct {
	// PlayerID stores the target player id.
	PlayerID int64 `json:"playerId" required:"true"`
	// Kind stores the delivered notification kind.
	Kind string `json:"kind" required:"true"`
	// Key stores the i18n message key.
	Key string `json:"key" required:"true"`
	// Sent reports whether the packet was sent.
	Sent bool `json:"sent" required:"true"`
}

// CurrencyWalletRequest identifies one player's wallet.
type CurrencyWalletRequest struct {
	APIKeyRequest
	// PlayerID stores the target player id.
	PlayerID int64 `query:"playerId" required:"true" minimum:"1"`
}

// CurrencyMutationRequest contains one administrative currency mutation.
type CurrencyMutationRequest struct {
	APIKeyRequest
	// PlayerID stores the target player id.
	PlayerID int64 `json:"playerId" required:"true" minimum:"1"`
	// CurrencyType stores the signed protocol currency type.
	CurrencyType int32 `json:"currencyType" required:"true" example:"5"`
	// Amount stores a positive delta or non-negative absolute balance.
	Amount int64 `json:"amount" required:"true" minimum:"0"`
	// Reason stores an optional ledger audit reason.
	Reason string `json:"reason,omitempty"`
	// Alert requests an additional localized generic alert.
	Alert bool `json:"alert,omitempty" default:"false" description:"Disabled by default. When true, sends a localized alert if the player is online."`
	// Locale optionally overrides the alert locale.
	Locale string `json:"locale,omitempty" example:"es"`
}

// CurrencyBalanceResponse contains one currency balance.
type CurrencyBalanceResponse struct {
	// CurrencyType identifies the protocol currency.
	CurrencyType int32 `json:"currencyType" required:"true"`
	// Amount stores the absolute balance.
	Amount int64 `json:"amount" required:"true"`
}

// CurrencyWalletResponse contains one player's configured wallet.
type CurrencyWalletResponse struct {
	// PlayerID identifies the wallet owner.
	PlayerID int64 `json:"playerId" required:"true"`
	// Balances stores configured currency balances.
	Balances []CurrencyBalanceResponse `json:"balances" required:"true"`
}

// CurrencyTypeResponse contains one configured currency type.
type CurrencyTypeResponse struct {
	// Type identifies the protocol currency.
	Type int32 `json:"type" required:"true"`
	// Key identifies its localized name.
	Key string `json:"key" required:"true"`
	// Ledger reports whether mutations are audited.
	Ledger bool `json:"ledger" required:"true"`
	// Color stores an optional future client presentation color.
	Color string `json:"color,omitempty"`
}

// CurrencyTypesResponse contains configured currency types.
type CurrencyTypesResponse struct {
	// Types stores configured currency definitions.
	Types []CurrencyTypeResponse `json:"types" required:"true"`
}

// CurrencyMutationResponse contains a committed currency mutation.
type CurrencyMutationResponse struct {
	// PlayerID identifies the affected player.
	PlayerID int64 `json:"playerId" required:"true"`
	// CurrencyType identifies the affected currency.
	CurrencyType int32 `json:"currencyType" required:"true"`
	// Amount stores the resulting absolute balance.
	Amount int64 `json:"amount" required:"true"`
	// AlertRequested reports whether alert delivery was requested.
	AlertRequested bool `json:"alertRequested" required:"true"`
	// AlertSent reports whether the optional alert reached a live player.
	AlertSent bool `json:"alertSent" required:"true"`
}
