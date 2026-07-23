// Package bubble validates and persists player chat bubble styles.
package bubble

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	bubblerepo "github.com/niflaot/pixels/internal/realm/chat/bubble/repository"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

var (
	// ErrInvalidBubble reports a negative protocol style or threshold.
	ErrInvalidBubble = errors.New("invalid chat bubble")
	// ErrBubbleLocked reports a style above the player's primary-group weight.
	ErrBubbleLocked = errors.New("chat bubble is locked")
)

// ProfileStore persists player bubble selections.
type ProfileStore interface {
	// UpdateBubbleStyle persists one validated selection.
	UpdateBubbleStyle(context.Context, int64, int32) (playermodel.Profile, error)
}

// PermissionReader resolves bubble bypasses and primary groups.
type PermissionReader interface {
	permissionservice.Checker
	// PrimaryGroup returns the player's highest-weight group.
	PrimaryGroup(context.Context, int64) (permissionmodel.Group, bool, error)
}

// Service validates thresholds and updates durable and live profile state.
type Service struct {
	// unlocks stores style thresholds.
	unlocks bubblerepo.Store
	// profiles persists profile selections.
	profiles ProfileStore
	// permissions resolves bypasses and primary group weights.
	permissions PermissionReader
	// players stores live player snapshots.
	players *playerlive.Registry
	// bubbleAny stores the unrestricted bubble capability.
	bubbleAny permission.Node
}

// New creates a chat bubble service.
func New(unlocks bubblerepo.Store, profiles ProfileStore, permissions PermissionReader, players *playerlive.Registry, bubbleAny permission.Node) *Service {
	return &Service{unlocks: unlocks, profiles: profiles, permissions: permissions, players: players, bubbleAny: bubbleAny}
}

// Allowed reports whether a player may select one bubble style.
func (service *Service) Allowed(ctx context.Context, playerID int64, bubbleID int32) (bool, error) {
	if playerID <= 0 || bubbleID < 0 {
		return false, ErrInvalidBubble
	}
	bypass, err := service.permissions.HasPermission(ctx, playerID, service.bubbleAny)
	if err != nil || bypass {
		return bypass, err
	}
	minimum, found, err := service.unlocks.MinWeight(ctx, bubbleID)
	if err != nil || !found || minimum <= 0 {
		return err == nil, err
	}
	group, found, err := service.permissions.PrimaryGroup(ctx, playerID)
	if err != nil || !found {
		return false, err
	}

	return group.Weight >= minimum, nil
}

// Select validates and persists one player's bubble style.
func (service *Service) Select(ctx context.Context, playerID int64, bubbleID int32) error {
	allowed, err := service.Allowed(ctx, playerID, bubbleID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrBubbleLocked
	}
	profile, err := service.profiles.UpdateBubbleStyle(ctx, playerID, bubbleID)
	if err != nil {
		return err
	}
	player, found := service.players.Find(playerID)
	if !found {
		return nil
	}
	snapshot := player.Snapshot()
	snapshot.BubbleStyle = profile.BubbleStyle

	return player.ReplaceSnapshot(snapshot)
}

// List returns configured bubble unlock thresholds.
func (service *Service) List(ctx context.Context) ([]bubblerepo.Unlock, error) {
	return service.unlocks.List(ctx)
}

// SetUnlock creates or replaces one bubble threshold.
func (service *Service) SetUnlock(ctx context.Context, bubbleID int32, minWeight int32) error {
	if bubbleID < 0 || minWeight < 0 {
		return ErrInvalidBubble
	}

	return service.unlocks.Set(ctx, bubbleID, minWeight)
}

// DeleteUnlock removes one bubble threshold.
func (service *Service) DeleteUnlock(ctx context.Context, bubbleID int32) error {
	if bubbleID < 0 {
		return ErrInvalidBubble
	}

	return service.unlocks.Delete(ctx, bubbleID)
}
