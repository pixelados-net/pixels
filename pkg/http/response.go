// Package http contains the Fiber HTTP transport for Pixels.
package http

// ErrorResponse is a JSON error response body.
type ErrorResponse struct {
	// Error stores the human-readable error message.
	Error string `json:"error"`
}

// StatusResponse is the public server status response body.
type StatusResponse struct {
	// Status stores the service health label.
	Status string `json:"status"`
	// Environment stores the runtime environment.
	Environment string `json:"environment"`
	// Version stores the running build version.
	Version string `json:"version"`
}

// CreateSSOTicketRequest is the private SSO ticket creation body.
type CreateSSOTicketRequest struct {
	// UserID stores the user bound to the ticket.
	UserID string `json:"userId"`
	// IP stores the optional client address bound to the ticket.
	IP string `json:"ip,omitempty"`
	// TTLSeconds stores the optional ticket lifetime override.
	TTLSeconds int `json:"ttlSeconds,omitempty"`
}

// CreateSSOTicketResponse is the private SSO ticket creation response.
type CreateSSOTicketResponse struct {
	// Ticket stores the opaque one-time SSO ticket.
	Ticket string `json:"ticket"`
	// ExpiresAt stores the ticket expiration in RFC3339 format.
	ExpiresAt string `json:"expiresAt"`
}
