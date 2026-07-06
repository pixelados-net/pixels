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
	}
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
