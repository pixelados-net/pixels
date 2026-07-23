package engine

import (
	"context"
	"time"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"go.uber.org/zap"
)

// playerTrigger identifies queued achievement deltas across goal metadata.
type playerTrigger struct {
	// playerID identifies the affected player.
	playerID int64
	// key identifies the shared trigger.
	key string
}

// playerAchievement identifies one cached durable progress value.
type playerAchievement struct {
	// playerID identifies the affected player.
	playerID int64
	// definitionID identifies the achievement group.
	definitionID int64
}

// run owns trigger batching and periodic flushes.
func (service *Service) run() {
	ticker := time.NewTicker(service.config.FlushInterval)
	defer ticker.Stop()
	defer close(service.done)
	for {
		select {
		case value := <-service.queue:
			service.metrics.AddQueue(-1)
			if service.add(value) {
				if err := service.flush(context.Background(), value.playerID); err != nil {
					service.log.Error("flush progression threshold", zap.Int64("player_id", value.playerID), zap.Error(err))
				}
			}
		case <-ticker.C:
			if err := service.flush(context.Background(), 0); err != nil {
				service.log.Error("flush progression", zap.Error(err))
			}
		case <-service.stop:
			service.drain()
			return
		}
	}
}

// add accumulates one trigger delta.
func (service *Service) add(value trigger) bool {
	service.mutex.Lock()
	key := triggerKey{playerID: value.playerID, key: value.key, data: value.data, daily: value.daily}
	added := value.amount
	if value.daily {
		if current := service.pending[key]; current >= value.amount {
			added = 0
		} else {
			added = value.amount - current
			service.pending[key] = value.amount
		}
	} else {
		service.pending[key] += value.amount
	}
	forecastKey := playerTrigger{playerID: value.playerID, key: value.key}
	service.forecast[forecastKey] += added
	flush := service.crossesThresholdLocked(value.playerID, value.key, service.forecast[forecastKey])
	service.metrics.SetPending(len(service.pending))
	service.mutex.Unlock()
	return flush
}

// drain moves queued values into the pending batch.
func (service *Service) drain() {
	for {
		select {
		case value := <-service.queue:
			service.metrics.AddQueue(-1)
			service.add(value)
		default:
			return
		}
	}
}

// FlushPlayer synchronously persists one player's pending progress.
func (service *Service) FlushPlayer(ctx context.Context, playerID int64) error {
	return service.flush(ctx, playerID)
}

// flush swaps and persists the selected pending batch.
func (service *Service) flush(ctx context.Context, playerID int64) error {
	service.mutex.Lock()
	if len(service.pending) == 0 {
		service.mutex.Unlock()
		return nil
	}
	selected := make(map[triggerKey]int64, len(service.pending))
	for key, amount := range service.pending {
		if playerID == 0 || key.playerID == playerID {
			selected[key] = amount
			delete(service.pending, key)
			forecastKey := playerTrigger{playerID: key.playerID, key: key.key}
			service.forecast[forecastKey] -= amount
			if service.forecast[forecastKey] <= 0 {
				delete(service.forecast, forecastKey)
			}
		}
	}
	service.mutex.Unlock()
	service.metrics.SetPending(service.pendingCount())
	for key, amount := range selected {
		if err := service.ProgressNowData(ctx, key.playerID, key.key, key.data, amount, key.daily); err != nil {
			service.restore(key, amount)
			return err
		}
	}
	return nil
}

// restore returns one failed delta to the pending batch.
func (service *Service) restore(key triggerKey, amount int64) {
	service.mutex.Lock()
	added := amount
	if key.daily {
		if current := service.pending[key]; current >= amount {
			added = 0
		} else {
			added = amount - current
			service.pending[key] = amount
		}
	} else {
		service.pending[key] += amount
	}
	service.forecast[playerTrigger{playerID: key.playerID, key: key.key}] += added
	service.metrics.SetPending(len(service.pending))
	service.mutex.Unlock()
}

// pendingCount returns the current aggregated pending key count.
func (service *Service) pendingCount() int {
	service.mutex.Lock()
	count := len(service.pending)
	service.mutex.Unlock()
	return count
}

// HydratePlayer loads one player's durable achievement positions once at login.
func (service *Service) HydratePlayer(ctx context.Context, playerID int64) error {
	if service == nil || playerID <= 0 {
		return nil
	}
	values, err := service.store.PlayerAchievements(ctx, playerID)
	if err != nil {
		return err
	}
	service.mutex.Lock()
	for key := range service.known {
		if key.playerID == playerID {
			delete(service.known, key)
		}
	}
	for _, value := range values {
		service.known[playerAchievement{playerID: playerID, definitionID: value.DefinitionID}] = value.Progress
	}
	service.hydrated[playerID] = true
	service.mutex.Unlock()
	return nil
}

// ForgetPlayer releases one disconnected player's forecast cache.
func (service *Service) ForgetPlayer(playerID int64) {
	if service == nil || playerID <= 0 {
		return
	}
	service.mutex.Lock()
	delete(service.hydrated, playerID)
	for key := range service.known {
		if key.playerID == playerID {
			delete(service.known, key)
		}
	}
	service.mutex.Unlock()
}

// observeProgress stores one committed monotonic progress position.
func (service *Service) observeProgress(playerID int64, definitionID int64, progress int64) {
	service.mutex.Lock()
	service.known[playerAchievement{playerID: playerID, definitionID: definitionID}] = progress
	service.mutex.Unlock()
}

// crossesThresholdLocked reports whether queued deltas reach any next level.
func (service *Service) crossesThresholdLocked(playerID int64, key string, amount int64) bool {
	if !service.hydrated[playerID] || amount <= 0 {
		return false
	}
	for _, definition := range service.catalog.Achievements(key) {
		progress := service.known[playerAchievement{playerID: playerID, definitionID: definition.ID}]
		if reachesNextLevel(definition.Levels, progress, amount) {
			return true
		}
	}
	return false
}

// reachesNextLevel reports whether one positive delta reaches the next threshold.
func reachesNextLevel(levels []progressionrecord.AchievementLevel, progress int64, amount int64) bool {
	for _, level := range levels {
		if level.ProgressNeeded > progress {
			return progress+amount >= level.ProgressNeeded
		}
	}
	return false
}
