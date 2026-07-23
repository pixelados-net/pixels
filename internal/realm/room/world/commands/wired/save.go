package wired

import (
	"context"
	"errors"
	"time"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	rewardeffect "github.com/niflaot/pixels/internal/realm/room/world/wired/effect/reward"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	successpacket "github.com/niflaot/pixels/networking/outbound/furniture/wired/save/success"
)

// HandleSave validates, commits, recompiles, and acknowledges one editor save.
func (handler Handler) HandleSave(ctx context.Context, envelope command.Envelope[SaveCommand]) error {
	player, active, roomID, err := handler.actor(envelope.Command.Handler)
	if err != nil {
		return err
	}
	if err = handler.authorize(ctx, player.ID(), active); err != nil {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.no_rights")
	}
	settings := envelope.Command.Settings
	stored, found, err := handler.Store.Find(ctx, roomID, int64(settings.ItemID))
	if err != nil || !found {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.target_missing")
	}
	descriptor, found := handler.Registry.Resolve(stored.Interaction)
	if !found || familyOf(descriptor.Family) != envelope.Command.Family {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.invalid_settings")
	}
	if privilegedProgression(descriptor.Key) {
		if err = handler.authorizeSuperwired(ctx, player.ID()); err != nil {
			return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.no_rights")
		}
	}
	expectedVersion := stored.Version
	stored.IntParams = append([]int32(nil), settings.IntParams...)
	stored.StringParam = settings.StringParam
	stored.SelectionMode = settings.SelectionMode
	stored.DelayPulses = settings.DelayPulses
	stored.Targets = mergeTargets(stored.Targets, settings.ItemIDs)
	if _, compileErr := handler.Compiler.CompileNode(stored); compileErr != nil {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.invalid_settings")
	}
	if descriptor.Key == "wf_act_give_reward" {
		rewards, parseErr := rewardeffect.Parse(stored.StringParam)
		if parseErr != nil {
			return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.invalid_settings")
		}
		_, err = handler.Store.SaveRewardConfig(ctx, stored, expectedVersion, rewards)
	} else {
		_, err = handler.Store.Save(ctx, stored, expectedVersion)
	}
	if err != nil {
		key := "room.wired.save.invalid_settings"
		if errors.Is(err, record.ErrTargetMissing) {
			key = "room.wired.save.target_missing"
		}
		if errors.Is(err, record.ErrConflict) {
			key = "room.wired.save.conflict"
		}
		return handler.sendFailure(ctx, envelope.Command.Handler, key)
	}
	if err = handler.Engine.Reload(ctx, roomID, time.Now()); err != nil {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.technical")
	}
	packet, err := successpacket.Encode()
	if err != nil {
		return err
	}
	return envelope.Command.Handler.Send(ctx, packet)
}

// privilegedProgression reports effects that mutate progression outside ordinary gameplay.
func privilegedProgression(key string) bool {
	return key == "wf_act_progress_achievement" || key == "wf_act_progress_quest" || key == "wf_act_start_quest"
}

// familyOf maps manifest families to editor save families.
func familyOf(family registry.Family) Family {
	switch family {
	case registry.FamilyTrigger:
		return TriggerFamily
	case registry.FamilyEffect:
		return EffectFamily
	case registry.FamilyCondition:
		return ConditionFamily
	default:
		return 0
	}
}

// mergeTargets preserves snapshots for still-selected furniture.
func mergeTargets(previous []record.Target, itemIDs []int32) []record.Target {
	snapshots := make(map[int64]record.Snapshot, len(previous))
	for _, target := range previous {
		snapshots[target.ItemID] = target.Snapshot
	}
	result := make([]record.Target, len(itemIDs))
	for index, itemID := range itemIDs {
		resolved := int64(itemID)
		result[index] = record.Target{ItemID: resolved, Snapshot: snapshots[resolved]}
	}
	return result
}

// classifyCompileError maps compiler validation into stable localized feedback.
func classifyCompileError(err error) string {
	if errors.Is(err, configuration.ErrUnsupported) {
		return "room.wired.save.unsupported_editor"
	}
	return "room.wired.save.invalid_settings"
}
