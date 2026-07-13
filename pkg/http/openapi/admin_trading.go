package openapi

import (
	"net/http"
	"time"
)

// AdminTradePlayerRequest documents one target player.
type AdminTradePlayerRequest struct {
	APIKeyRequest
	// PlayerID identifies the target player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
}

// AdminMarketplaceListingRequest documents one listing id.
type AdminMarketplaceListingRequest struct {
	APIKeyRequest
	// ID identifies the target listing.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// AdminTradeAudit documents one completed trade audit.
type AdminTradeAudit struct {
	// ID identifies the audit row.
	ID int64 `json:"id"`
	// RoomID identifies the settlement room.
	RoomID int64 `json:"roomId"`
	// FirstPlayerID identifies the first participant.
	FirstPlayerID int64 `json:"firstPlayerId"`
	// SecondPlayerID identifies the second participant.
	SecondPlayerID int64 `json:"secondPlayerId"`
	// FirstIP stores the first participant audit address.
	FirstIP string `json:"firstIp,omitempty"`
	// SecondIP stores the second participant audit address.
	SecondIP string `json:"secondIp,omitempty"`
	// FirstItemIDs stores the first offer.
	FirstItemIDs []int64 `json:"firstItemIds"`
	// SecondItemIDs stores the second offer.
	SecondItemIDs []int64 `json:"secondItemIds"`
	// CreatedAt stores settlement time.
	CreatedAt time.Time `json:"createdAt"`
	// FirstRedeemableCredits stores value delivered to the second participant.
	FirstRedeemableCredits int64 `json:"firstRedeemableCredits"`
	// SecondRedeemableCredits stores value delivered to the first participant.
	SecondRedeemableCredits int64 `json:"secondRedeemableCredits"`
}

// AdminTradeAuditList documents bounded trade audit results.
type AdminTradeAuditList struct {
	// Items stores recent audits.
	Items []AdminTradeAudit `json:"items"`
	// Count stores the returned item count.
	Count int `json:"count"`
}

// adminTradingOperations returns protected trading administration operations.
func adminTradingOperations() []operation {
	return []operation{adminTrading(http.MethodGet, "/api/admin/trade/players/{playerId}/log", "List player trade audit", &AdminTradePlayerRequest{}, &AdminTradeAuditList{}, http.StatusOK), adminTrading(http.MethodPost, "/api/admin/trade/players/{playerId}/lock", "Lock player trading", &AdminTradePlayerRequest{}, nil, http.StatusNoContent), adminTrading(http.MethodDelete, "/api/admin/trade/players/{playerId}/lock", "Unlock player trading", &AdminTradePlayerRequest{}, nil, http.StatusNoContent), adminTrading(http.MethodPost, "/api/admin/marketplace/listings/{id}/force-close", "Force-close Marketplace listing", &AdminMarketplaceListingRequest{}, nil, http.StatusNoContent)}
}

// adminTrading creates one protected trading operation.
func adminTrading(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Admin Trading", summary: summary, description: summary + ".", request: request, responses: responses, secured: true}
}
