// Package lobby owns cached Game Center listing and launch behavior.
package lobby

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	gamecenterconfig "github.com/niflaot/pixels/internal/realm/gamecenter/config"
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
	outlist "github.com/niflaot/pixels/networking/outbound/gamecenter/lobby/list"
)

// ErrUnavailable reports a disabled or unknown game.
var ErrUnavailable = errors.New("game center game unavailable")

// Service serves immutable cached game registrations.
type Service struct {
	// config stores feature policy.
	config gamecenterconfig.Config
	// store persists game registrations.
	store gamecenterrecord.Store
	// mutex protects the immutable snapshots.
	mutex sync.RWMutex
	// games stores enabled games by id.
	games map[int32]gamecenterrecord.Game
	// ordered stores enabled games in display order.
	ordered []gamecenterrecord.Game
	// launches counts successful external launches.
	launches atomic.Uint64
}

// New creates an empty Game Center service.
func New(config gamecenterconfig.Config, store gamecenterrecord.Store) *Service {
	return &Service{config: config, store: store, games: make(map[int32]gamecenterrecord.Game)}
}

// Reload atomically replaces the enabled game cache.
func (service *Service) Reload(ctx context.Context) error {
	if !service.config.Enabled {
		return nil
	}
	games, err := service.store.ListGames(ctx, true)
	if err != nil {
		return err
	}
	index := make(map[int32]gamecenterrecord.Game, len(games))
	ordered := make([]gamecenterrecord.Game, 0, len(games))
	for _, game := range games {
		if !game.Enabled {
			continue
		}
		index[game.ID] = game
		ordered = append(ordered, game)
	}
	service.mutex.Lock()
	service.games, service.ordered = index, ordered
	service.mutex.Unlock()
	return nil
}

// List returns a stable client game-list projection.
func (service *Service) List() []outlist.Game {
	service.mutex.RLock()
	result := make([]outlist.Game, len(service.ordered))
	for index, game := range service.ordered {
		result[index] = outlist.Game{ID: game.ID, Name: game.Name, BackgroundColor: game.BackgroundColor, TextColor: game.TextColor, AssetURL: game.AssetURL, SupportURL: game.SupportURL}
	}
	service.mutex.RUnlock()
	return result
}

// FindLaunch returns one enabled launchable game.
func (service *Service) FindLaunch(gameID int32) (gamecenterrecord.Game, error) {
	service.mutex.RLock()
	game, found := service.games[gameID]
	service.mutex.RUnlock()
	if !service.config.Enabled || !found || strings.TrimSpace(game.LaunchURL) == "" {
		return gamecenterrecord.Game{}, ErrUnavailable
	}
	return game, nil
}

// RecordLaunch increments successful external launches.
func (service *Service) RecordLaunch() { service.launches.Add(1) }

// Launches returns the lock-free successful launch count.
func (service *Service) Launches() uint64 { return service.launches.Load() }
