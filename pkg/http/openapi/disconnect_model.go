package openapi

// DisconnectAllRequest contains a global disconnect body.
type DisconnectAllRequest struct {
	APIKeyRequest
	DisconnectRequest
}

// DisconnectKindRequest contains a kind disconnect request.
type DisconnectKindRequest struct {
	APIKeyRequest
	// Kind stores the target connection kind.
	Kind string `path:"kind" required:"true"`
	DisconnectRequest
}

// DisconnectOneRequest contains a single connection disconnect request.
type DisconnectOneRequest struct {
	APIKeyRequest
	// Kind stores the target connection kind.
	Kind string `path:"kind" required:"true"`
	// ID stores the target connection id.
	ID string `path:"id" required:"true"`
	DisconnectRequest
}

// DisconnectRequest contains a disconnect operation body.
type DisconnectRequest struct {
	// Reason stores the stable disconnect reason label.
	Reason string `json:"reason" required:"true" example:"kicked"`
	// Message stores optional operator context.
	Message string `json:"message,omitempty"`
}

// DisconnectResponse contains bulk disconnect results.
type DisconnectResponse struct {
	// Matched stores how many connections matched the request.
	Matched int `json:"matched" required:"true"`
	// Disconnected stores how many disconnect calls succeeded.
	Disconnected int `json:"disconnected" required:"true"`
	// Errors stores how many disconnect calls failed.
	Errors int `json:"errors" required:"true"`
}
