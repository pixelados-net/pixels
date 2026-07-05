// Package http contains the Fiber HTTP transport for Pixels.
package http

// ErrorResponse is a JSON error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatusResponse is the public server status response body.
type StatusResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

// CreateSSOTicketRequest is the private SSO ticket creation body.
type CreateSSOTicketRequest struct {
	UserID     string `json:"userId"`
	IP         string `json:"ip,omitempty"`
	TTLSeconds int    `json:"ttlSeconds,omitempty"`
}

// CreateSSOTicketResponse is the private SSO ticket creation response.
type CreateSSOTicketResponse struct {
	Ticket    string `json:"ticket"`
	ExpiresAt string `json:"expiresAt"`
}
