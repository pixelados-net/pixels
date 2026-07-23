package openapi

import "net/http"

// wiredAdminOperations returns protected WIRED administration operations.
func wiredAdminOperations() []operation {
	return []operation{
		wiredAdmin(http.MethodGet, "/api/admin/wired/registry", "List canonical WIRED registry", &APIKeyRequest{}, &WiredRegistryResponse{}, http.StatusOK),
		wiredAdmin(http.MethodGet, "/api/admin/rooms/{id}/wired", "Inspect room WIRED graph", &WiredRoomRequest{}, &WiredRoomResponse{}, http.StatusOK),
		wiredAdmin(http.MethodGet, "/api/admin/rooms/{id}/wired/{itemId}", "Read WIRED item configuration", &WiredItemRequest{}, &WiredConfigResponse{}, http.StatusOK),
		wiredAdmin(http.MethodPut, "/api/admin/rooms/{id}/wired/{itemId}", "Replace WIRED item configuration", &WiredConfigUpdateRequest{}, &WiredConfigResponse{}, http.StatusOK),
		wiredAdmin(http.MethodPost, "/api/admin/rooms/{id}/wired/{itemId}/apply-snapshot", "Capture WIRED furniture snapshot", &WiredItemRequest{}, &WiredActionResponse{}, http.StatusOK),
		wiredAdmin(http.MethodPost, "/api/admin/rooms/{id}/wired/game/{action}", "Change WIRED game lifecycle", &WiredGameRequest{}, &WiredActionResponse{}, http.StatusOK),
		wiredAdmin(http.MethodGet, "/api/admin/rooms/{id}/wired/traces", "Read sanitized WIRED traces", &WiredRoomRequest{}, &[]WiredTraceResponse{}, http.StatusOK),
		wiredAdmin(http.MethodGet, "/api/admin/rooms/{id}/wired/metrics", "Read WIRED runtime metrics", &WiredRoomRequest{}, &WiredMetricsResponse{}, http.StatusOK),
		wiredAdmin(http.MethodPost, "/api/admin/rooms/{id}/wired/reload", "Reload room WIRED generation", &WiredRoomRequest{}, &WiredActionResponse{}, http.StatusOK),
		wiredAdmin(http.MethodPut, "/api/admin/rooms/{id}/wired/visibility", "Change WIRED box visibility", &WiredVisibilityRequest{}, &WiredVisibilityResponse{}, http.StatusOK),
		wiredAdmin(http.MethodGet, "/api/admin/rooms/{id}/wired/{itemId}/rewards", "List WIRED reward definitions", &WiredItemRequest{}, &[]WiredRewardModel{}, http.StatusOK),
		wiredAdmin(http.MethodPut, "/api/admin/rooms/{id}/wired/{itemId}/rewards", "Replace WIRED reward definitions", &WiredRewardsRequest{}, &WiredActionResponse{}, http.StatusOK),
		wiredAdmin(http.MethodDelete, "/api/admin/rooms/{id}/wired/{itemId}/rewards", "Delete WIRED reward definitions", &WiredItemRequest{}, nil, http.StatusNoContent),
	}
}

// wiredAdmin creates one protected WIRED administration operation.
func wiredAdmin(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError, http.StatusServiceUnavailable)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "WIRED", summary: summary, description: summary + ".", request: request, responses: responses, secured: true}
}
