// Package sanction wires global punishment persistence and effects.
package sanction

import (
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	sanctiondb "github.com/niflaot/pixels/internal/realm/sanction/database"
	"github.com/niflaot/pixels/internal/realm/sanction/enforcement/projection"
	enforcementsession "github.com/niflaot/pixels/internal/realm/sanction/enforcement/session"
	"github.com/niflaot/pixels/internal/realm/sanction/enforcement/warning"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	sanctionsession "github.com/niflaot/pixels/internal/realm/sanction/session"
	"go.uber.org/fx"
)

// Module provides global punishment behavior.
var Module = fx.Module("realm-sanction", fx.Provide(sanctiondb.New, NewStore, sanctioncore.New, warning.NewWarn, enforcementsession.NewBan, enforcementsession.NewKick, projection.NewMute, projection.NewTradeLock), fx.Invoke(RegisterAppliers, sanctionsession.RegisterLifecycle, sanctionsession.ConfigureAuthentication))

// NewStore exposes PostgreSQL persistence through the domain boundary.
func NewStore(repository *sanctiondb.Repository) sanctionrecord.Store { return repository }

// RegisterAppliers installs every initial punishment behavior.
func RegisterAppliers(service *sanctioncore.Service, warn *warning.Warn, ban *enforcementsession.Ban, kick *enforcementsession.Kick, mute *projection.Mute, tradeLock *projection.TradeLock) error {
	if err := service.Register(warn); err != nil {
		return err
	}
	if err := service.Register(ban); err != nil {
		return err
	}
	if err := service.Register(kick); err != nil {
		return err
	}
	if err := service.Register(mute); err != nil {
		return err
	}
	if err := service.Register(tradeLock); err != nil {
		return err
	}
	return nil
}
