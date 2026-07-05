package sso

import "time"

// Ticket contains a created SSO ticket.
type Ticket struct {
	// Value is the opaque ticket sent to the client.
	Value string
	// UserID is the TODO user identifier bound to the ticket.
	UserID string
	// IP is the optional client IP address bound to the ticket.
	IP string
	// ExpiresAt is the time when Redis expires the ticket.
	ExpiresAt time.Time
}

// CreateRequest contains ticket creation input.
type CreateRequest struct {
	// UserID is the TODO user identifier to bind.
	UserID string
	// IP is the optional client IP address to bind.
	IP string
	// TTL overrides the configured default ticket lifetime.
	TTL time.Duration
}

// ConsumeRequest contains ticket consumption input.
type ConsumeRequest struct {
	// Ticket is the opaque ticket value to consume.
	Ticket string
	// IP is the optional client IP address to validate.
	IP string
}

// record is the Redis payload for a ticket.
type record struct {
	// UserID stores the user bound to the ticket.
	UserID string `json:"userId"`
	// IP stores the optional address bound to the ticket.
	IP string `json:"ip,omitempty"`
	// ExpiresAt stores the Redis expiration timestamp.
	ExpiresAt time.Time `json:"expiresAt"`
}
