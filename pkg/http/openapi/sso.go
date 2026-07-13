package openapi

import "net/http"

// ssoOperations returns protected SSO route operations.
func ssoOperations() []operation {
	return []operation{
		{
			method:      http.MethodPost,
			path:        "/api/sso/tickets",
			tag:         "SSO",
			summary:     "Create SSO ticket",
			description: "Creates a Redis-backed one-time SSO ticket. Idempotency-Key makes transport retries replay-safe.",
			request:     &CreateSSOTicketRequest{},
			responses: append(
				[]response{jsonResponse(http.StatusCreated, &CreateSSOTicketResponse{}, "SSO ticket created.")},
				errorResponses(http.StatusBadRequest, http.StatusConflict, http.StatusUnauthorized, http.StatusInternalServerError)...,
			),
			secured: true,
		},
	}
}
