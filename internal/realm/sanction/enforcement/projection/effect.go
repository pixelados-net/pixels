// Package projection projects active sanctions into live player state.
package projection

import (
	"context"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// ActiveReader loads current sanctions after rare mutations.
type ActiveReader interface {
	// Active returns one player's timestamp-derived projection.
	Active(context.Context, int64) (sanctionrecord.ActiveState, error)
}

// Mute projects active global mute into live player state.
type Mute struct {
	// sanctions loads the aggregate projection.
	sanctions ActiveReader
	// players locates authenticated players.
	players *playerlive.Registry
}

// NewMute creates a mute behavior.
func NewMute(sanctions *sanctioncore.Service, players *playerlive.Registry) *Mute {
	return &Mute{sanctions: sanctions, players: players}
}

// Kind identifies global mute behavior.
func (*Mute) Kind() sanctionrecord.Kind { return sanctionrecord.KindMute }

// Apply refreshes the live projection after persistence.
func (effect *Mute) Apply(ctx context.Context, punishment sanctionrecord.Punishment) error {
	return effect.refresh(ctx, punishment.ReceiverPlayerID)
}

// Revoke refreshes the live projection after revocation.
func (effect *Mute) Revoke(ctx context.Context, punishment sanctionrecord.Punishment) error {
	return effect.refresh(ctx, punishment.ReceiverPlayerID)
}

// refresh maps active state into one live player.
func (effect *Mute) refresh(ctx context.Context, playerID int64) error {
	state, err := effect.sanctions.Active(ctx, playerID)
	if err != nil {
		return err
	}
	player, found := effect.players.Find(playerID)
	if !found {
		return nil
	}
	snapshot := player.Snapshot().Sanctions
	snapshot.MutePermanent = state.MutedPermanently
	snapshot.MuteUntil = time.Time{}
	if state.MuteUntil != nil {
		snapshot.MuteUntil = *state.MuteUntil
	}
	player.SetSanctions(snapshot)
	return nil
}

// TradeLock projects active locks into the live snapshot and compatibility column.
type TradeLock struct {
	// sanctions loads aggregate active state.
	sanctions ActiveReader
	// players persists compatibility state.
	players playerservice.TradeManager
	// live locates authenticated players.
	live *playerlive.Registry
}

// NewTradeLock creates a trade-lock behavior.
func NewTradeLock(sanctions *sanctioncore.Service, players playerservice.TradeManager, live *playerlive.Registry) *TradeLock {
	return &TradeLock{sanctions: sanctions, players: players, live: live}
}

// Kind identifies trade-lock behavior.
func (*TradeLock) Kind() sanctionrecord.Kind { return sanctionrecord.KindTradeLock }

// Apply refreshes durable and live projections.
func (effect *TradeLock) Apply(ctx context.Context, punishment sanctionrecord.Punishment) error {
	return effect.refresh(ctx, punishment.ReceiverPlayerID)
}

// Revoke refreshes durable and live projections.
func (effect *TradeLock) Revoke(ctx context.Context, punishment sanctionrecord.Punishment) error {
	return effect.refresh(ctx, punishment.ReceiverPlayerID)
}

// refresh derives compatibility state from active truth.
func (effect *TradeLock) refresh(ctx context.Context, playerID int64) error {
	state, err := effect.sanctions.Active(ctx, playerID)
	if err != nil {
		return err
	}
	locked := state.TradeLockedPermanently || state.TradeLockUntil != nil
	if err = effect.players.SetAllowTrade(ctx, playerID, !locked); err != nil {
		return err
	}
	player, found := effect.live.Find(playerID)
	if !found {
		return nil
	}
	snapshot := player.Snapshot().Sanctions
	snapshot.TradeLockPermanent = state.TradeLockedPermanently
	snapshot.TradeLockUntil = time.Time{}
	if state.TradeLockUntil != nil {
		snapshot.TradeLockUntil = *state.TradeLockUntil
	}
	player.SetSanctions(snapshot)
	player.SetAllowTrade(!locked)
	return nil
}
