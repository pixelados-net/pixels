package openapi

// APIKeyRequest contains the API key header.
type APIKeyRequest struct {
	// APIKey stores the configured access key.
	APIKey string `header:"X-API-Key" required:"true" description:"Access key configured by PIXELS_ACCESS_KEY."`
}

// WebSocketUpgradeRequest contains WebSocket upgrade headers.
type WebSocketUpgradeRequest struct {
	// Connection stores the upgrade connection header.
	Connection string `header:"Connection" required:"true" example:"Upgrade"`
	// Upgrade stores the websocket upgrade header.
	Upgrade string `header:"Upgrade" required:"true" example:"websocket"`
}

// ErrorResponse is a JSON error response body.
type ErrorResponse struct {
	// Error stores the human-readable error message.
	Error string `json:"error" required:"true"`
}

// StatusResponse is the public server status response body.
type StatusResponse struct {
	// Status stores the service health label.
	Status string `json:"status" required:"true" example:"ok"`
	// Environment stores the runtime environment.
	Environment string `json:"environment" required:"true" example:"development"`
	// Version stores the running build version.
	Version string `json:"version" required:"true" example:"0.1.0-abcdef12"`
}

// CreateSSOTicketRequest is the private SSO ticket creation body.
type CreateSSOTicketRequest struct {
	APIKeyRequest
	// PlayerID stores the player bound to the ticket.
	PlayerID int64 `json:"playerId" required:"true" minimum:"1" description:"Existing player id bound to the SSO ticket."`
	// IP stores the optional client address bound to the ticket.
	IP string `json:"ip,omitempty" description:"Optional IP address allowed to consume the ticket."`
	// TTLSeconds stores the optional ticket lifetime override.
	TTLSeconds int `json:"ttlSeconds,omitempty" minimum:"1" description:"Optional TTL override in seconds."`
}

// CreateSSOTicketResponse is the private SSO ticket creation response.
type CreateSSOTicketResponse struct {
	// Ticket stores the opaque one-time SSO ticket.
	Ticket string `json:"ticket" required:"true" description:"Opaque one-time SSO ticket."`
	// ExpiresAt stores the ticket expiration in RFC3339 format.
	ExpiresAt string `json:"expiresAt" required:"true" format:"date-time"`
}
