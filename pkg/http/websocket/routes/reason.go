package routes

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// reasonCodes stores disconnect reasons accepted by admin routes.
var reasonCodes = []netconn.DisconnectCode{
	netconn.DisconnectUnknown,
	netconn.DisconnectLocalClose,
	netconn.DisconnectRemoteClose,
	netconn.DisconnectTransportError,
	netconn.DisconnectProtocolError,
	netconn.DisconnectAuthenticationFailed,
	netconn.DisconnectAuthenticationTimeout,
	netconn.DisconnectDuplicateSession,
	netconn.DisconnectIdleTimeout,
	netconn.DisconnectRateLimited,
	netconn.DisconnectPolicyViolation,
	netconn.DisconnectKicked,
	netconn.DisconnectBanned,
	netconn.DisconnectServerShutdown,
}

// reasonFromRequest converts a request body into a disconnect reason.
func reasonFromRequest(request DisconnectRequest) (netconn.Reason, error) {
	reason, ok := parseReason(request.Reason)
	if !ok {
		return netconn.Reason{}, fiber.NewError(fiber.StatusBadRequest, "unsupported disconnect reason")
	}

	return netconn.Reason{Code: reason, Message: request.Message}, nil
}

// parseReason returns a disconnect code by stable label.
func parseReason(reason string) (netconn.DisconnectCode, bool) {
	normalized := strings.ToLower(strings.TrimSpace(reason))
	for _, code := range reasonCodes {
		if code.String() == normalized {
			return code, true
		}
	}

	return 0, false
}

// reasonItems returns the supported reason response rows.
func reasonItems() []ReasonResponse {
	items := make([]ReasonResponse, 0, len(reasonCodes))
	for _, code := range reasonCodes {
		items = append(items, ReasonResponse{Code: uint16(code), Reason: code.String()})
	}

	return items
}
