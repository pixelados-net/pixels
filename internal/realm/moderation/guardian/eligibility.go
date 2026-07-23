package guardian

import (
	"context"
	"strconv"
)

// eligible selects non-excluded guardians in FIFO pool order.
func (manager *Manager) eligible(ctx context.Context, reporterID int64, reportedID int64) []int64 {
	limit := manager.config.GuardianCount * 4
	if limit < manager.config.GuardianCount {
		limit = manager.config.GuardianCount
	}
	candidates := manager.pool.Guardians(reporterID, limit)
	values := make([]int64, 0, manager.config.GuardianCount)
	for _, playerID := range candidates {
		if playerID == reportedID || manager.byPlayer[playerID] != 0 {
			continue
		}
		if manager.redis != nil {
			_, excluded, err := manager.redis.Find(ctx, guardianExclusionKey(playerID))
			if err != nil || excluded {
				continue
			}
		}
		values = append(values, playerID)
		if len(values) == manager.config.GuardianCount {
			break
		}
	}
	return values
}

// recordIgnored increments ignored offers and applies temporary exclusion.
func (manager *Manager) recordIgnored(ctx context.Context, playerID int64) {
	if manager.redis == nil {
		return
	}
	count, err := manager.redis.Increment(ctx, guardianIgnoredKey(playerID), manager.config.GuardianExclusion)
	if err == nil && count >= int64(manager.config.GuardianIgnoreLimit) {
		_ = manager.redis.Set(ctx, guardianExclusionKey(playerID), []byte{1}, manager.config.GuardianExclusion)
	}
}

// guardianIgnoredKey returns one distributed ignored-offer counter key.
func guardianIgnoredKey(playerID int64) string {
	return "moderation:guardian:ignored:" + strconv.FormatInt(playerID, 10)
}

// guardianExclusionKey returns one distributed temporary exclusion key.
func guardianExclusionKey(playerID int64) string {
	return "moderation:guardian:excluded:" + strconv.FormatInt(playerID, 10)
}
