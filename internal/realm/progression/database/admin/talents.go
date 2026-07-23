package admin

import (
	"context"
	"encoding/json"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// UpsertTalentLevel creates or replaces one talent level.
func (repository *Repository) UpsertTalentLevel(ctx context.Context, value progressionrecord.TalentLevel) error {
	if value.Requirements == nil {
		value.Requirements = []progressionrecord.TalentRequirement{}
	}
	if value.RewardItems == nil {
		value.RewardItems = []int64{}
	}
	if value.RewardPerks == nil {
		value.RewardPerks = []string{}
	}
	if value.RewardBadges == nil {
		value.RewardBadges = []string{}
	}
	requirements, err := json.Marshal(value.Requirements)
	if err != nil {
		return err
	}
	_, err = repository.executorFor(ctx).Exec(ctx, `insert into talent_track_levels(track,level,requirements,reward_items,reward_perks,reward_badges) values($1,$2,$3,$4,$5,$6) on conflict(track,level) do update set requirements=excluded.requirements,reward_items=excluded.reward_items,reward_perks=excluded.reward_perks,reward_badges=excluded.reward_badges`, value.Track, value.Level, requirements, value.RewardItems, value.RewardPerks, value.RewardBadges)
	return err
}
