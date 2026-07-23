package openapi

import (
	"net/http"

	permissionapi "github.com/niflaot/pixels/pkg/http/openapi/permission"
)

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
		adminRoomTemplate(http.MethodGet, "/api/admin/rooms/bundle-templates", "List room bundle templates", &APIKeyRequest{}, &RoomListResponse{}),
		adminRoomTemplate(http.MethodPost, "/api/admin/rooms/{id}/bundle-template", "Mark room as bundle template", &RoomIDRequest{}, &RoomResponse{}),
		adminRoomTemplate(http.MethodDelete, "/api/admin/rooms/{id}/bundle-template", "Unmark room bundle template", &RoomIDRequest{}, &RoomResponse{}),
		adminRoomRead("/api/admin/rooms/{id}", "Read room metadata", &RoomIDRequest{}, &RoomResponse{}),
		adminRoomRead("/api/admin/rooms/{id}/occupancy", "Read active room occupancy", &RoomIDRequest{}, &RoomOccupancyResponse{}),
		adminRoomRead("/api/admin/rooms/{id}/promotion", "Read active room promotion", &RoomIDRequest{}, &RoomPromotionResponse{}),
		adminRoomMutation(http.MethodDelete, "/api/admin/rooms/{id}/promotion", "Cancel active room promotion", &RoomIDRequest{}, nil),
		adminRoomSettings(http.MethodPatch, "/api/admin/rooms/{id}/roller", "Update room roller speed", &RoomRollerSettingsRequest{}, &RoomResponse{}),
		adminRoomAction("/api/admin/rooms/{id}/close", "Close active room", &RoomIDRequest{}),
		adminRoomAction("/api/admin/rooms/{id}/forward", "Forward active room occupants", &RoomForwardRequest{}),
		adminRoomAction("/api/admin/rooms/players/{playerId}/teleport", "Teleport one live player", &RoomTeleportRequest{}),
		adminRoomAudit("/api/admin/rooms/{id}/rights/history", "Read room rights history", &RoomAuditRequest{}, &RoomRightsAuditResponse{}),
		adminRoomAudit("/api/admin/rooms/{id}/moderation/history", "Read room moderation history", &RoomAuditRequest{}, &RoomModerationAuditResponse{}),
		adminRoomAudit("/api/admin/rooms/{id}/bans", "List active room bans", &RoomAuditRequest{}, &RoomSanctionResponse{}),
		adminRoomAudit("/api/admin/rooms/{id}/mutes", "List active room mutes", &RoomAuditRequest{}, &RoomSanctionResponse{}),
		adminRoomAudit("/api/admin/players/{playerId}/moderation/history", "Read moderation received by player", &PlayerAuditRequest{}, &RoomModerationAuditResponse{}),
		adminRoomAudit("/api/admin/players/{playerId}/moderation/actions", "Read moderation performed by player", &PlayerAuditRequest{}, &RoomModerationAuditResponse{}),
		adminNavigatorRead("/api/admin/navigator/categories", "List navigator categories", &APIKeyRequest{}, &CategoryListResponse{}),
		adminNavigatorRead("/api/admin/navigator/lifted", "List navigator lifted rooms", &APIKeyRequest{}, &LiftedListResponse{}),
		adminNavigatorRead("/api/admin/navigator/history", "Read player navigator history", &NavigatorPlayerRequest{}, &NavigatorRoomIDsResponse{}),
		adminNavigatorRead("/api/admin/navigator/favorites", "Read player navigator favorites", &NavigatorPlayerRequest{}, &NavigatorRoomIDsResponse{}),
		adminNavigatorMutation(http.MethodDelete, "/api/admin/navigator/history", "Delete player navigator history", &NavigatorHistoryDeleteRequest{}, &NavigatorDeleteResponse{}),
		adminNavigatorMutation(http.MethodPost, "/api/admin/navigator/official/{roomId}", "Mark room official", &NavigatorOfficialRequest{}, &RoomResponse{}),
		adminNavigatorMutation(http.MethodDelete, "/api/admin/navigator/official/{roomId}", "Unmark room official", &NavigatorOfficialRequest{}, &RoomResponse{}),
		adminNotificationAction("/api/admin/notifications/send", "Send localized player notification", &NotificationRequest{}),
		adminMessenger(http.MethodGet, "/api/admin/players/{playerId}/friends", "List player friends", &MessengerPlayerRequest{}, &MessengerFriendsResponse{}, http.StatusOK),
		adminMessenger(http.MethodGet, "/api/admin/players/{playerId}/friends/requests", "List player friend requests", &MessengerPlayerRequest{}, &MessengerRequestsResponse{}, http.StatusOK),
		adminMessenger(http.MethodDelete, "/api/admin/players/{playerId}/friends/{friendId}", "Remove player friendship", &MessengerFriendRequest{}, &MessengerMutationResponse{}, http.StatusOK),
		adminMessenger(http.MethodPost, "/api/admin/players/{playerId}/privacy", "Update player messenger privacy", &MessengerPrivacyRequest{}, &MessengerPrivacyResponse{}, http.StatusOK),
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
		adminCatalog(http.MethodGet, "/api/admin/catalog/vouchers", "List catalog vouchers", &APIKeyRequest{}, &VoucherListResponse{}),
		adminCatalog(http.MethodPost, "/api/admin/catalog/vouchers", "Create catalog voucher", &VoucherRequest{}, &VoucherResponse{}),
		adminCatalog(http.MethodPatch, "/api/admin/catalog/vouchers/{id}", "Update catalog voucher", &VoucherPatchRequest{}, &VoucherResponse{}),
		adminCatalog(http.MethodGet, "/api/admin/catalog/vouchers/{id}/redemptions", "List voucher redemptions", &CatalogIDRequest{}, &VoucherRedemptionListResponse{}),
		adminSubscription(http.MethodGet, "/api/admin/subscriptions/{playerId}", "Read player membership", &SubscriptionPlayerRequest{}, &SubscriptionResponse{}, http.StatusOK),
		adminSubscription(http.MethodPost, "/api/admin/subscriptions/{playerId}/grant", "Grant or extend membership", &SubscriptionGrantRequest{}, &SubscriptionResponse{}, http.StatusOK),
		adminSubscription(http.MethodDelete, "/api/admin/subscriptions/{playerId}", "Revoke membership", &SubscriptionPlayerRequest{}, nil, http.StatusNoContent),
		adminSubscription(http.MethodGet, "/api/admin/subscriptions/club-offers", "List club offers", &APIKeyRequest{}, &ClubOfferListResponse{}, http.StatusOK),
		adminSubscription(http.MethodPost, "/api/admin/subscriptions/club-offers", "Create club offer", &ClubOfferRequest{}, &ClubOfferResponse{}, http.StatusOK),
		adminSubscription(http.MethodPatch, "/api/admin/subscriptions/club-offers/{id}", "Update club offer", &ClubOfferPatchRequest{}, &ClubOfferResponse{}, http.StatusOK),
		adminSubscription(http.MethodGet, "/api/admin/subscriptions/targeted-offers", "List targeted offers", &APIKeyRequest{}, &TargetedOfferListResponse{}, http.StatusOK),
		adminSubscription(http.MethodPost, "/api/admin/subscriptions/targeted-offers", "Create targeted offer", &TargetedOfferRequest{}, &TargetedOfferResponse{}, http.StatusOK),
		adminSubscription(http.MethodPatch, "/api/admin/subscriptions/targeted-offers/{id}", "Update targeted offer", &TargetedOfferPatchRequest{}, &TargetedOfferResponse{}, http.StatusOK),
		adminSubscription(http.MethodGet, "/api/admin/subscriptions/calendar/campaigns", "List calendar campaigns", &APIKeyRequest{}, &CampaignListResponse{}, http.StatusOK),
		adminSubscription(http.MethodPost, "/api/admin/subscriptions/calendar/campaigns", "Create calendar campaign", &CampaignRequest{}, &CampaignResponse{}, http.StatusOK),
		adminSubscription(http.MethodPatch, "/api/admin/subscriptions/calendar/campaigns/{id}", "Update calendar campaign", &CampaignPatchRequest{}, &CampaignResponse{}, http.StatusOK),
		adminChat(http.MethodGet, "/api/admin/chat/filters", "List global chat filters", &APIKeyRequest{}, &ChatFilterListResponse{}, http.StatusOK),
		adminChat(http.MethodPost, "/api/admin/chat/filters", "Add global chat filter", &ChatFilterRequest{}, &ChatMutationResponse{}, http.StatusCreated),
		adminChat(http.MethodDelete, "/api/admin/chat/filters/{word}", "Remove global chat filter", &ChatFilterDeleteRequest{}, nil, http.StatusNoContent),
		adminChat(http.MethodGet, "/api/admin/chat/bubbles", "List chat bubble thresholds", &APIKeyRequest{}, &ChatBubbleListResponse{}, http.StatusOK),
		adminChat(http.MethodPut, "/api/admin/chat/bubbles/{bubbleId}", "Set chat bubble threshold", &ChatBubbleRequest{}, &ChatMutationResponse{}, http.StatusOK),
		adminChat(http.MethodDelete, "/api/admin/chat/bubbles/{bubbleId}", "Remove chat bubble threshold", &ChatBubbleDeleteRequest{}, nil, http.StatusNoContent),
		adminChat(http.MethodGet, "/api/admin/rooms/{id}/chat/history", "Read room chat history", &ChatHistoryRoomRequest{}, &ChatHistoryResponse{}, http.StatusOK),
		adminChat(http.MethodGet, "/api/admin/players/{playerId}/chat/history", "Read player chat history", &ChatHistoryPlayerRequest{}, &ChatHistoryResponse{}, http.StatusOK),
		adminPermission(http.MethodGet, "/api/admin/permissions/nodes", "List registered permission nodes", &permissionapi.APIKeyRequest{}, &permissionapi.NodesResponse{}, http.StatusOK),
		adminPermission(http.MethodGet, "/api/admin/permissions/groups", "List permission groups", &permissionapi.APIKeyRequest{}, &permissionapi.GroupsResponse{}, http.StatusOK),
		adminPermission(http.MethodPost, "/api/admin/permissions/groups", "Create permission group", &permissionapi.GroupCreateRequest{}, &permissionapi.GroupResponse{}, http.StatusCreated),
		adminPermission(http.MethodPatch, "/api/admin/permissions/groups/{id}", "Update permission group", &permissionapi.GroupPatchRequest{}, &permissionapi.GroupResponse{}, http.StatusOK),
		adminPermission(http.MethodPost, "/api/admin/permissions/groups/{id}/nodes", "Grant group permission node", &permissionapi.GroupNodeRequest{}, &permissionapi.MutationResponse{}, http.StatusOK),
		adminPermission(http.MethodDelete, "/api/admin/permissions/groups/{id}/nodes/{node}", "Revoke group permission node", &permissionapi.GroupNodeDeleteRequest{}, nil, http.StatusNoContent),
		adminPermission(http.MethodPost, "/api/admin/permissions/players/{playerId}/groups/{groupId}", "Add player to permission group", &permissionapi.MembershipRequest{}, &permissionapi.MutationResponse{}, http.StatusOK),
		adminPermission(http.MethodDelete, "/api/admin/permissions/players/{playerId}/groups/{groupId}", "Remove player from permission group", &permissionapi.MembershipRequest{}, nil, http.StatusNoContent),
		adminPermission(http.MethodPost, "/api/admin/permissions/players/{playerId}/nodes", "Grant direct player permission node", &permissionapi.PlayerNodeRequest{}, &permissionapi.MutationResponse{}, http.StatusOK),
		adminPermission(http.MethodDelete, "/api/admin/permissions/players/{playerId}/nodes/{node}", "Revoke direct player permission node", &permissionapi.PlayerNodeDeleteRequest{}, nil, http.StatusNoContent),
		adminPermission(http.MethodGet, "/api/admin/permissions/players/{playerId}/effective", "List effective player permissions", &permissionapi.PlayerRequest{}, &permissionapi.EffectiveResponse{}, http.StatusOK),
		adminPermission(http.MethodGet, "/api/admin/permissions/players/{playerId}/check", "Check player permission", &permissionapi.CheckRequest{}, &permissionapi.CheckResponse{}, http.StatusOK),
	}
}

// adminRoomSettings creates a protected room settings mutation operation.
func adminRoomSettings(method string, path string, summary string, request any, body any) operation {
	return operation{method: method, path: path, tag: "Rooms", summary: summary,
		description: summary + ".", request: request,
		responses: append([]response{jsonResponse(http.StatusOK, body, summary+".")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)...), secured: true}
}

// adminRoomTemplate creates a protected room bundle template operation.
func adminRoomTemplate(method string, path string, summary string, request any, body any) operation {
	return operation{method: method, path: path, tag: "Room Bundles", summary: summary,
		description: summary + ".", request: request,
		responses: append([]response{jsonResponse(http.StatusOK, body, summary+".")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)...), secured: true}
}

// adminRoomAudit creates a protected room audit read operation.
func adminRoomAudit(path string, summary string, request any, body any) operation {
	return operation{
		method: http.MethodGet, path: path, tag: "Room Audit", summary: summary,
		description: summary + ".", request: request,
		responses: append([]response{jsonResponse(http.StatusOK, body, summary+".")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError)...),
		secured: true,
	}
}

// adminPermission creates a permission administration operation.
func adminPermission(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}

	return operation{method: method, path: path, tag: "Permissions", summary: summary,
		description: summary + ".", request: request, responses: responses, secured: true}
}

// adminRead creates a read-only admin operation.
func adminRead(method string, path string, summary string, request any, body any) operation {
	return operation{
		method:      method,
		path:        path,
		tag:         "Connections",
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
		tag:         "Connections",
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

// adminNotificationAction creates a player notification operation.
func adminNotificationAction(path string, summary string, request any) operation {
	return operation{
		method:      http.MethodPost,
		path:        path,
		tag:         "Notifications",
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
		tag:         "Currencies",
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
		tag:         "Currencies",
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
		tag:         "Rooms",
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
