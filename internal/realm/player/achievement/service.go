package achievement

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// MaxEquippedBadges is Nitro's fixed wearable badge capacity.
const MaxEquippedBadges = 5

var (
	// ErrInvalidBadge reports an empty player or malformed badge selection.
	ErrInvalidBadge = errors.New("invalid player badge")
	// ErrBadgeNotOwned reports an attempt to wear another player's badge.
	ErrBadgeNotOwned = errors.New("player badge not owned")
)

// Service coordinates durable achievement state and hot-path badge snapshots.
type Service struct {
	// store persists achievement records.
	store Store
	// mutex protects equipped badge snapshots.
	mutex sync.RWMutex
	// equipped stores equipped badge sets for online players.
	equipped map[int64]map[string]struct{}
}

// New creates an achievement service.
func New(store Store) *Service {
	return &Service{store: store, equipped: make(map[int64]map[string]struct{})}
}

// Load refreshes one player's equipped badge snapshot.
func (service *Service) Load(ctx context.Context, playerID int64) error {
	badges, err := service.store.Badges(ctx, playerID)
	if err != nil {
		return err
	}
	service.replaceSnapshot(playerID, badges)
	return nil
}

// List returns one player's durable badge inventory and active slots.
func (service *Service) List(ctx context.Context, playerID int64) ([]Badge, error) {
	if playerID <= 0 || service.store == nil {
		return nil, ErrInvalidBadge
	}
	return service.store.Badges(ctx, playerID)
}

// SetEquipped validates ownership and atomically replaces active badge slots.
func (service *Service) SetEquipped(ctx context.Context, playerID int64, requested []string) ([]Badge, error) {
	badges, err := service.List(ctx, playerID)
	if err != nil {
		return nil, err
	}
	owned := make(map[string]struct{}, len(badges))
	for _, badge := range badges {
		owned[strings.ToUpper(badge.Code)] = struct{}{}
	}
	codes := make([]string, 0, MaxEquippedBadges)
	seen := make(map[string]struct{}, MaxEquippedBadges)
	for _, value := range requested {
		code := strings.ToUpper(strings.TrimSpace(value))
		if code == "" {
			continue
		}
		if len(code) > 64 || len(codes) == MaxEquippedBadges {
			return badges, ErrInvalidBadge
		}
		if _, found := owned[code]; !found {
			return badges, ErrBadgeNotOwned
		}
		if _, duplicate := seen[code]; duplicate {
			return badges, ErrInvalidBadge
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if err = service.store.SetEquipped(ctx, playerID, codes); err != nil {
		return nil, err
	}
	for index := range badges {
		badges[index].Equipped = false
		badges[index].Slot = 0
		for slot, code := range codes {
			if strings.EqualFold(badges[index].Code, code) {
				badges[index].Equipped = true
				badges[index].Slot = int32(slot + 1)
				break
			}
		}
	}
	service.replaceSnapshot(playerID, badges)
	return badges, nil
}

// replaceSnapshot installs one immutable equipped badge lookup.
func (service *Service) replaceSnapshot(playerID int64, badges []Badge) {
	values := make(map[string]struct{})
	for _, badge := range badges {
		if badge.Equipped {
			values[strings.ToUpper(badge.Code)] = struct{}{}
		}
	}
	service.mutex.Lock()
	service.equipped[playerID] = values
	service.mutex.Unlock()
}

// Unload releases one offline player's badge snapshot.
func (service *Service) Unload(playerID int64) {
	service.mutex.Lock()
	delete(service.equipped, playerID)
	service.mutex.Unlock()
}

// Wearing reports an allocation-free equipped badge lookup.
func (service *Service) Wearing(playerID int64, code string) (bool, bool) {
	service.mutex.RLock()
	badges, loaded := service.equipped[playerID]
	_, found := badges[code]
	service.mutex.RUnlock()
	return found, loaded
}

// GrantBadge grants one durable unequipped badge.
func (service *Service) GrantBadge(ctx context.Context, playerID int64, code string, source string) (bool, error) {
	granted, err := service.store.GrantBadge(ctx, playerID, strings.ToUpper(strings.TrimSpace(code)), source)
	return granted, err
}

// ReplaceBadge replaces one owned badge code while preserving its active slot.
func (service *Service) ReplaceBadge(ctx context.Context, playerID int64, oldCode string, newCode string, source string) (bool, error) {
	replaced, err := service.store.ReplaceBadge(ctx, playerID, strings.ToUpper(strings.TrimSpace(oldCode)), strings.ToUpper(strings.TrimSpace(newCode)), source)
	if err != nil || !replaced {
		return replaced, err
	}
	if err = service.Load(ctx, playerID); err != nil {
		return false, err
	}
	return true, nil
}

// RemoveBadge removes one owned badge and refreshes the hot-path snapshot.
func (service *Service) RemoveBadge(ctx context.Context, playerID int64, code string) (bool, error) {
	removed, err := service.store.RemoveBadge(ctx, playerID, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil || !removed {
		return removed, err
	}
	if err = service.Load(ctx, playerID); err != nil {
		return false, err
	}
	return true, nil
}

// GrantRespect applies one idempotent durable respect grant.
func (service *Service) GrantRespect(ctx context.Context, playerID int64, amount int32, sourceKey string) (bool, error) {
	if amount <= 0 || amount > 1000 || sourceKey == "" {
		return false, nil
	}
	return service.store.GrantRespect(ctx, playerID, amount, sourceKey, "wired")
}
