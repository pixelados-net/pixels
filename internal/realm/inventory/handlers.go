package inventory

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	requestcmd "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	requesthandler "github.com/niflaot/pixels/internal/realm/inventory/currency/handlers/request"
	"go.uber.org/zap"
)

// RegisterConnectionHandlers registers inventory packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, request *requestcmd.Handler, log *zap.Logger) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}

	requesthandler.Register(handlers.Inbound, requesthandler.New(request, log))
}
