package wiring

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	wiredcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/wired"
	wiredhandler "github.com/niflaot/pixels/internal/realm/room/world/handlers/wired"
	"go.uber.org/zap"
)

// RegisterHandlers installs all five Nitro WIRED editor packet adapters.
func RegisterHandlers(handlers *realmconn.Handlers, command wiredcmd.Handler, log *zap.Logger) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	wiredhandler.Register(handlers.Inbound, wiredhandler.New(command, log))
}
