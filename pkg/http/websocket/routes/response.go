package routes

import (
	"time"

	netconn "github.com/niflaot/pixels/networking/connection"
)

// CountResponse contains connection count data.
type CountResponse struct {
	// Total stores the total active connection count.
	Total int `json:"total"`
	// Kind stores the optional filtered connection kind.
	Kind string `json:"kind,omitempty"`
	// Count stores the optional filtered kind count.
	Count *int `json:"count,omitempty"`
}

// ConnectionResponse contains safe connection data.
type ConnectionResponse struct {
	// ID stores the connection id.
	ID string `json:"id"`
	// Kind stores the connection kind.
	Kind string `json:"kind"`
	// State stores the lifecycle state label.
	State string `json:"state"`
	// StartedAt stores the session start time.
	StartedAt string `json:"startedAt"`
	// AuthenticatedAt stores the optional authentication time.
	AuthenticatedAt string `json:"authenticatedAt,omitempty"`
}

// ListResponse contains connection list data.
type ListResponse struct {
	// Total stores the returned connection count.
	Total int `json:"total"`
	// Items stores safe connection rows.
	Items []ConnectionResponse `json:"items"`
}

// DisconnectRequest contains a disconnect operation body.
type DisconnectRequest struct {
	// Reason stores the stable disconnect reason label.
	Reason string `json:"reason"`
	// Message stores optional operator context.
	Message string `json:"message,omitempty"`
}

// DisconnectResponse contains bulk disconnect results.
type DisconnectResponse struct {
	// Matched stores how many connections matched the request.
	Matched int `json:"matched"`
	// Disconnected stores how many disconnect calls succeeded.
	Disconnected int `json:"disconnected"`
	// Errors stores how many disconnect calls failed.
	Errors int `json:"errors"`
}

// ReasonResponse contains one supported disconnect reason.
type ReasonResponse struct {
	// Code stores the numeric disconnect code.
	Code uint16 `json:"code"`
	// Reason stores the stable disconnect reason label.
	Reason string `json:"reason"`
}

// ReasonsResponse contains supported disconnect reasons.
type ReasonsResponse struct {
	// Items stores supported reason rows.
	Items []ReasonResponse `json:"items"`
}

// connectionResponse converts a connection without exposing its IP.
func connectionResponse(connection netconn.Connection) ConnectionResponse {
	authenticatedAt, authenticated := connection.AuthenticatedAt()
	response := ConnectionResponse{
		ID:        string(connection.ID()),
		Kind:      string(connection.Kind()),
		State:     connection.State().String(),
		StartedAt: formatTime(connection.StartedAt()),
	}
	if authenticated {
		response.AuthenticatedAt = formatTime(authenticatedAt)
	}

	return response
}

// formatTime returns a stable JSON timestamp string.
func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
