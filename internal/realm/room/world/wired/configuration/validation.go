package configuration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// validate enforces shared and descriptor-specific schema rules.
func (compiler *Compiler) validate(stored record.Config, descriptor registry.Descriptor) error {
	if stored.ItemID <= 0 || stored.RoomID <= 0 || stored.DelayPulses < 0 || stored.DelayPulses > compiler.config.MaxDelayPulses {
		return ErrInvalid
	}
	if stored.SelectionMode < 0 || stored.SelectionMode > 3 || len(stored.Targets) > compiler.config.MaxSelection || len(stored.IntParams) > 32 || len(stored.StringParam) > 2048 {
		return ErrInvalid
	}
	if descriptor.Selection == registry.SelectionNone && len(stored.Targets) != 0 {
		return fmt.Errorf("%w: targets rejected", ErrInvalid)
	}
	if descriptor.Selection == registry.SelectionNone && stored.SelectionMode != 0 {
		return fmt.Errorf("%w: selection mode rejected", ErrInvalid)
	}
	if descriptor.Selection == registry.SelectionRequired && len(stored.Targets) == 0 {
		return fmt.Errorf("%w: targets required", ErrInvalid)
	}
	if len(stored.Targets) > 0 && stored.SelectionMode == 0 {
		return fmt.Errorf("%w: selection mode required", ErrInvalid)
	}
	seen := make(map[int64]struct{}, len(stored.Targets))
	for _, target := range stored.Targets {
		if target.ItemID <= 0 {
			return fmt.Errorf("%w: invalid target", ErrInvalid)
		}
		if _, exists := seen[target.ItemID]; exists {
			return fmt.Errorf("%w: duplicate target", ErrInvalid)
		}
		seen[target.ItemID] = struct{}{}
	}
	return validateBehavior(stored, descriptor.Key)
}

// validateBehavior enforces schemas whose editor fields have strict meaning.
func validateBehavior(stored record.Config, key string) error {
	switch key {
	case "wf_trg_says_something":
		if strings.TrimSpace(stored.StringParam) == "" || len(stored.StringParam) > 100 {
			return fmt.Errorf("%w: keyword", ErrInvalid)
		}
	case "wf_trg_periodically", "wf_trg_period_long", "wf_trg_at_given_time", "wf_trg_at_time_long", "wf_cnd_time_more_than", "wf_cnd_time_less_than":
		if !positiveAt(stored.IntParams, 0) {
			return fmt.Errorf("%w: duration", ErrInvalid)
		}
	case "wf_cnd_user_count_in", "wf_cnd_not_user_count":
		if len(stored.IntParams) < 2 || stored.IntParams[0] < 0 || stored.IntParams[1] < stored.IntParams[0] {
			return fmt.Errorf("%w: user range", ErrInvalid)
		}
	case "wf_act_join_team", "wf_act_give_score_tm", "wf_cnd_actor_in_team", "wf_cnd_not_in_team":
		if len(stored.IntParams) == 0 || stored.IntParams[0] < 1 || stored.IntParams[0] > 4 {
			return fmt.Errorf("%w: team", ErrInvalid)
		}
	case "wf_cnd_date_rng_active":
		if len(stored.IntParams) < 2 || stored.IntParams[1] < stored.IntParams[0] {
			return fmt.Errorf("%w: date range", ErrInvalid)
		}
	case "wf_act_mute_triggerer":
		if !positiveAt(stored.IntParams, 0) || stored.IntParams[0] > 1440 {
			return fmt.Errorf("%w: mute duration", ErrInvalid)
		}
	case "wf_act_give_score":
		if len(stored.IntParams) < 2 || stored.IntParams[0] < 1 || stored.IntParams[0] > 100 || stored.IntParams[1] < 1 || stored.IntParams[1] > 10 {
			return fmt.Errorf("%w: score and use limit", ErrInvalid)
		}
	case "wf_act_progress_achievement":
		if strings.TrimSpace(stored.StringParam) == "" {
			return fmt.Errorf("%w: achievement group", ErrInvalid)
		}
	case "wf_act_progress_quest", "wf_act_start_quest":
		value, err := strconv.ParseInt(strings.TrimSpace(stored.StringParam), 10, 64)
		if err != nil || value <= 0 {
			return fmt.Errorf("%w: quest id", ErrInvalid)
		}
	}
	return nil
}

// positiveAt reports whether a setting index stores a positive value.
func positiveAt(values []int32, index int) bool {
	return index >= 0 && index < len(values) && values[index] > 0
}
