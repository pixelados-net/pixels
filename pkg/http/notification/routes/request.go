// Package routes contains protected player notification administration routes.
package routes

// NotifyRequest requests one end-user notification.
type NotifyRequest struct {
	// PlayerID stores the target player id.
	PlayerID int64 `json:"playerId"`
	// Kind stores the notification kind, either bubble or alert.
	Kind string `json:"kind"`
	// Key stores the localized message key.
	Key string `json:"key"`
	// Locale stores the optional locale override.
	Locale string `json:"locale,omitempty"`
	// BubbleKey stores the optional bubble alert type.
	BubbleKey string `json:"bubbleKey,omitempty"`
	// Params stores optional translation parameters.
	Params map[string]string `json:"params,omitempty"`
}

// NotifyResponse contains notification delivery status.
type NotifyResponse struct {
	// PlayerID stores the target player id.
	PlayerID int64 `json:"playerId"`
	// Kind stores the delivered notification kind.
	Kind string `json:"kind"`
	// Key stores the localized message key.
	Key string `json:"key"`
	// Sent reports whether a packet was sent.
	Sent bool `json:"sent"`
}
