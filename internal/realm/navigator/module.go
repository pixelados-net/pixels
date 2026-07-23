// Package navigator contains navigator realm persistence and runtime wiring.
package navigator

import (
	navhistory "github.com/niflaot/pixels/internal/realm/navigator/browse/history"
	navruntime "github.com/niflaot/pixels/internal/realm/navigator/browse/runtime"
	"github.com/niflaot/pixels/internal/realm/navigator/core"
	"github.com/niflaot/pixels/internal/realm/navigator/database"
	"github.com/niflaot/pixels/internal/realm/navigator/record"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides navigator realm persistence state.
var Module = fx.Module(
	"realm-navigator",
	fx.Provide(
		LoadConfig,
		NewStore,
		navruntime.NewCategoryCountBroadcaster,
		navruntime.NewRoomCountBroadcaster,
		core.New,
		NewManager,
		NewHistoryWriter,
		NewPreferenceWriter,
	),
	fx.Invoke(RegisterConnectionHandlers),
	fx.Invoke(navruntime.RegisterCategoryCounts),
	fx.Invoke(navruntime.RegisterRoomCounts),
	fx.Invoke(navhistory.Register),
	fx.Invoke(navsession.RegisterWriter),
)

// NewHistoryWriter creates the asynchronous visit writer from Navigator config.
func NewHistoryWriter(manager core.Manager, log *zap.Logger, config Config) *navhistory.Writer {
	return navhistory.NewConfigured(manager, log, config.HistoryQueueSize, config.HistoryDedupeWindow)
}

// NewPreferenceWriter creates the coalesced Navigator preference writer.
func NewPreferenceWriter(manager core.Manager, log *zap.Logger, config Config) *navsession.PreferenceWriter {
	return navsession.NewPreferenceWriter(manager, log, config.PreferenceFlushInterval, config.PreferencePendingLimit)
}

// NewStore creates the navigator persistence store.
func NewStore(pool *postgres.Pool) record.Store {
	return database.New(pool)
}

// NewManager exposes the navigator management contract.
func NewManager(service *core.Service) core.Manager {
	return service
}
