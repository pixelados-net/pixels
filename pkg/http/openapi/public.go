package openapi

import "net/http"

// publicOperations returns public route operations.
func publicOperations() []operation {
	return []operation{
		{
			method:      http.MethodGet,
			path:        "/status",
			tag:         "Public",
			summary:     "Read server status",
			description: "Returns public runtime status without requiring an API key.",
			responses:   []response{jsonResponse(http.StatusOK, &StatusResponse{}, "Server status.")},
		},
		{
			method:      http.MethodGet,
			path:        "/ws",
			tag:         "WebSocket",
			summary:     "Open websocket session",
			description: "Upgrades an HTTP request to the pixel-protocol websocket entrypoint.",
			request:     &WebSocketUpgradeRequest{},
			responses: append(
				[]response{emptyResponse(http.StatusSwitchingProtocols, "Websocket upgrade accepted.")},
				errorResponses(http.StatusUpgradeRequired)...,
			),
		},
		{
			method:      http.MethodGet,
			path:        "/docs",
			tag:         "Public",
			summary:     "Read Scalar API documentation",
			description: "Serves public Scalar documentation in development only.",
			responses: []response{
				{status: http.StatusOK, body: "", description: "Scalar documentation HTML.", contentType: "text/html"},
				jsonResponse(http.StatusNotFound, &ErrorResponse{}, "Documentation is disabled outside development."),
			},
		},
		{
			method:      http.MethodGet,
			path:        "/client/ui-config.json",
			tag:         "Client Config",
			summary:     "Read Nitro currency configuration",
			description: "Returns a partial Nitro UI configuration containing every configured currency type.",
			responses:   []response{jsonResponse(http.StatusOK, &CurrencyUIConfigResponse{}, "Nitro currency and room model UI configuration.")},
		},
		{
			method:      http.MethodGet,
			path:        "/client/texts/{locale}/ExternalTexts.json",
			tag:         "Client Config",
			summary:     "Read localized Nitro currency texts",
			description: "Returns ExternalTexts-compatible currency name entries for the requested locale.",
			request:     &CurrencyExternalTextsRequest{},
			responses:   []response{jsonResponse(http.StatusOK, &CurrencyExternalTextsResponse{}, "Localized Nitro currency names.")},
		},
	}
}

// fallbackOperation returns the authenticated fallback route operation.
func fallbackOperation() operation {
	return operation{
		method:      http.MethodGet,
		path:        "/*",
		tag:         "Fallback",
		summary:     "Private route fallback",
		description: "Represents protected endpoints added after public route registration.",
		request:     &APIKeyRequest{},
		responses:   errorResponses(http.StatusUnauthorized, http.StatusNotFound),
		secured:     true,
	}
}
