package openapi

import "net/http"

// adminOperations returns protected connection administration operations.
func adminOperations() []operation {
	return []operation{
		adminRead(http.MethodGet, "/api/admin/connections", "List connections", &ConnectionListRequest{}, &ConnectionListResponse{}),
		adminRead(http.MethodGet, "/api/admin/connections/list", "List connections", &ConnectionListRequest{}, &ConnectionListResponse{}),
		adminRead(http.MethodGet, "/api/admin/connections/count", "Count connections", &ConnectionCountRequest{}, &ConnectionCountResponse{}),
		adminRead(http.MethodGet, "/api/admin/connections/reasons", "List disconnect reasons", &APIKeyRequest{}, &ReasonsResponse{}),
		adminDisconnect("/api/admin/connections/disconnect", "Disconnect all connections", &DisconnectAllRequest{}),
		adminDisconnect("/api/admin/connections/{kind}/disconnect", "Disconnect connections by kind", &DisconnectKindRequest{}),
		adminDisconnect("/api/admin/connections/{kind}/{id}/disconnect", "Disconnect one connection", &DisconnectOneRequest{}),
		adminRoomRead("/api/admin/rooms", "List rooms", &RoomListRequest{}, &RoomListResponse{}),
		adminRoomRead("/api/admin/rooms/{id}", "Read room metadata", &RoomIDRequest{}, &RoomResponse{}),
		adminRoomRead("/api/admin/rooms/{id}/occupancy", "Read active room occupancy", &RoomIDRequest{}, &RoomOccupancyResponse{}),
		adminRoomAction("/api/admin/rooms/{id}/close", "Close active room", &RoomIDRequest{}),
		adminRoomAction("/api/admin/rooms/{id}/forward", "Forward active room occupants", &RoomForwardRequest{}),
		adminNavigatorRead("/api/admin/navigator/categories", "List navigator categories", &APIKeyRequest{}, &CategoryListResponse{}),
		adminNavigatorRead("/api/admin/navigator/lifted", "List navigator lifted rooms", &APIKeyRequest{}, &LiftedListResponse{}),
		adminNotificationAction("/api/admin/notifications/send", "Send localized player notification", &NotificationRequest{}),
		adminCurrencyRead("/api/admin/currencies/wallet", "Read player currency wallet", &CurrencyWalletRequest{}, &CurrencyWalletResponse{}),
		adminCurrencyRead("/api/admin/currencies/types", "List configured currency types", &APIKeyRequest{}, &CurrencyTypesResponse{}),
		adminCurrencyAction("/api/admin/currencies/grant", "Grant player currency"),
		adminCurrencyAction("/api/admin/currencies/deduct", "Deduct player currency"),
		adminCurrencyAction("/api/admin/currencies/set", "Set player currency balance"),
		adminCatalog(http.MethodGet, "/api/admin/catalog/pages", "List catalog pages", &APIKeyRequest{}, &CatalogPagesResponse{}),
		adminCatalog(http.MethodPost, "/api/admin/catalog/pages", "Create catalog page", &CatalogPageRequest{}, &CatalogPageResponse{}),
		adminCatalog(http.MethodPatch, "/api/admin/catalog/pages/{id}", "Update catalog page", &CatalogPagePatchRequest{}, &CatalogPageResponse{}),
		adminCatalog(http.MethodGet, "/api/admin/catalog/items", "List catalog offers", &CatalogItemsRequest{}, &CatalogItemsResponse{}),
		adminCatalog(http.MethodPost, "/api/admin/catalog/items", "Create catalog offer", &CatalogItemRequest{}, &CatalogItemResponse{}),
		adminCatalog(http.MethodPatch, "/api/admin/catalog/items/{id}", "Update catalog offer", &CatalogItemPatchRequest{}, &CatalogItemResponse{}),
		adminCatalog(http.MethodDelete, "/api/admin/catalog/items/{id}", "Delete catalog offer", &CatalogIDRequest{}, nil),
		adminCatalog(http.MethodPost, "/api/admin/catalog/refresh", "Refresh and publish catalog", &APIKeyRequest{}, &CatalogRefreshResponse{}),
		adminCatalog(http.MethodGet, "/api/admin/catalog/sanitize-list", "List definitions without offers", &APIKeyRequest{}, &CatalogDefinitionsResponse{}),
	}
}

// adminCatalog creates a catalog administration operation.
func adminCatalog(method string, path string, summary string, request any, body any) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(http.StatusNoContent, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(http.StatusOK, body, summary+".")}, responses...)
	}

	return operation{method: method, path: path, tag: "Admin Catalog", summary: summary,
		description: summary + ".", request: request, responses: responses, secured: true}
}

// adminRead creates a read-only admin operation.
func adminRead(method string, path string, summary string, request any, body any) operation {
	return operation{
		method:      method,
		path:        path,
		tag:         "Admin Connections",
		summary:     summary,
		description: summary + ".",
		request:     request,
		responses: append(
			[]response{jsonResponse(http.StatusOK, body, summary+".")},
			errorResponses(http.StatusUnauthorized)...,
		),
		secured: true,
	}
}

// adminDisconnect creates a disconnect admin operation.
func adminDisconnect(path string, summary string, request any) operation {
	return operation{
		method:      http.MethodPost,
		path:        path,
		tag:         "Admin Connections",
		summary:     summary,
		description: summary + ".",
		request:     request,
		responses: append(
			[]response{jsonResponse(http.StatusOK, &DisconnectResponse{}, "Connections disconnected.")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound)...,
		),
		secured: true,
	}
}

// adminRoomRead creates a read-only room admin operation.
func adminRoomRead(path string, summary string, request any, body any) operation {
	return adminTaggedRead("Admin Rooms", path, summary, request, body)
}

// adminNavigatorRead creates a read-only navigator admin operation.
func adminNavigatorRead(path string, summary string, request any, body any) operation {
	return adminTaggedRead("Admin Navigator", path, summary, request, body)
}

// adminNotificationAction creates a player notification operation.
func adminNotificationAction(path string, summary string, request any) operation {
	return operation{
		method:      http.MethodPost,
		path:        path,
		tag:         "Admin Notifications",
		summary:     summary,
		description: summary + ".",
		request:     request,
		responses: append(
			[]response{jsonResponse(http.StatusOK, &NotificationResponse{}, "Notification sent.")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound)...,
		),
		secured: true,
	}
}

// adminCurrencyRead creates a currency administration read operation.
func adminCurrencyRead(path string, summary string, request any, body any) operation {
	return operation{
		method:      http.MethodGet,
		path:        path,
		tag:         "Admin Currencies",
		summary:     summary,
		description: summary + ".",
		request:     request,
		responses: append(
			[]response{jsonResponse(http.StatusOK, body, summary+".")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError)...,
		),
		secured: true,
	}
}

// adminCurrencyAction creates a currency administration mutation operation.
func adminCurrencyAction(path string, summary string) operation {
	return operation{
		method:      http.MethodPost,
		path:        path,
		tag:         "Admin Currencies",
		summary:     summary,
		description: summary + ". Optional localized alert delivery is disabled by default.",
		request:     &CurrencyMutationRequest{},
		responses: append(
			[]response{jsonResponse(http.StatusOK, &CurrencyMutationResponse{}, "Currency mutation committed.")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)...,
		),
		secured: true,
	}
}

// adminRoomAction creates a room runtime admin operation.
func adminRoomAction(path string, summary string, request any) operation {
	return operation{
		method:      http.MethodPost,
		path:        path,
		tag:         "Admin Rooms",
		summary:     summary,
		description: summary + ".",
		request:     request,
		responses: append(
			[]response{jsonResponse(http.StatusOK, &RoomActionResponse{}, summary+".")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound)...,
		),
		secured: true,
	}
}

// adminTaggedRead creates a tagged read-only admin operation.
func adminTaggedRead(tag string, path string, summary string, request any, body any) operation {
	return operation{
		method:      http.MethodGet,
		path:        path,
		tag:         tag,
		summary:     summary,
		description: summary + ".",
		request:     request,
		responses: append(
			[]response{jsonResponse(http.StatusOK, body, summary+".")},
			errorResponses(http.StatusUnauthorized, http.StatusNotFound)...,
		),
		secured: true,
	}
}
