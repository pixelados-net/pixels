package runtime

import (
	"context"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomtask "github.com/niflaot/pixels/internal/realm/room/runtime/live/task"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/i18n"
)

// commandNeedStatus identifies Nitro's pet consumption animation.
const commandNeedStatus = "eat"

// commandNeedState stores one product selected by a contextual command.
type commandNeedState struct {
	// itemID identifies the selected placed product.
	itemID int64
	// actionID identifies the command cooldown slot.
	actionID int32
	// kind identifies the expected product interaction.
	kind CommandNeed
}

// executeCommandNeed selects a compatible product and reserves pet movement.
func (service *Service) executeCommandNeed(ctx context.Context, active *roomlive.Room, pet *activePet, petID int64, actorID int64, action CommandAction, command petrecord.Command) error {
	now := service.Now()
	pet.mutex.Lock()
	record := pet.record
	invalid := service.needs == nil || pet.stationary || record.Level < command.RequiredLevel || now.Before(pet.cooldowns[action.ID]) || pet.riderPlayerID != 0 || pet.needPending || pet.commandNeed.itemID != 0
	pet.mutex.Unlock()
	if invalid || action.Need == CommandNeedNone {
		return petrecord.ErrInvalidState
	}
	if !active.Snapshot().AllowPetsEat {
		service.stampCommandCooldown(pet, action.ID, command)
		_ = service.SpeakLocalized(ctx, active, record, i18n.Key("pet.command.need.feeding_disabled"))
		return petrecord.ErrFeedingDisabled
	}
	unit, found := active.UnitMotion(EntityKey(petID))
	if !found {
		return petrecord.ErrInvalidState
	}
	item, found := active.NearestFurnitureByInteraction(action.Need.interaction(), unit.Position.Point, -1)
	if !found {
		service.stampCommandCooldown(pet, action.ID, command)
		_ = service.SpeakLocalized(ctx, active, record, action.Need.missingKey())
		return petrecord.ErrInvalidProduct
	}
	pet.mutex.Lock()
	if pet.needPending || pet.commandNeed.itemID != 0 || now.Before(pet.cooldowns[action.ID]) {
		pet.mutex.Unlock()
		return petrecord.ErrInvalidState
	}
	pet.actionGeneration++
	pet.cooldowns[action.ID] = now.Add(command.Cooldown)
	pet.selectedBy = actorID
	pet.followingPlayerID, pet.stay = 0, false
	pet.commandNeed = commandNeedState{itemID: item.ID, actionID: action.ID, kind: action.Need}
	pet.mutex.Unlock()
	clearPetStatuses(active, petID)
	if unit.Position.Point == item.Point {
		return nil
	}
	if _, err := active.MoveControlled(EntityKey(petID), item.Point, worldunit.ControlFurnitureInteraction); err != nil {
		service.cancelCommandNeed(active, pet, petID, item.ID)
		_ = service.SpeakLocalized(ctx, active, record, i18n.Key("pet.command.need.unreachable"))
		return petrecord.ErrInvalidState
	}
	return nil
}

// resolveCommandNeed advances one reserved contextual product workflow.
func (service *Service) resolveCommandNeed(ctx context.Context, active *roomlive.Room, pet *activePet, record petrecord.Pet) {
	pet.mutex.Lock()
	pending := pet.commandNeed
	inFlight := pet.needPending
	stationary := pet.stationary
	pet.mutex.Unlock()
	if pending.itemID == 0 || inFlight {
		return
	}
	if stationary {
		service.cancelCommandNeed(active, pet, record.ID, pending.itemID)
		return
	}
	item, found := active.FurnitureItem(pending.itemID)
	unit, unitFound := active.UnitMotion(EntityKey(record.ID))
	if !active.Snapshot().AllowPetsEat {
		service.cancelCommandNeed(active, pet, record.ID, pending.itemID)
		_ = service.SpeakLocalized(ctx, active, record, i18n.Key("pet.command.need.feeding_disabled"))
		return
	}
	if !found || item.Definition.InteractionType != pending.kind.interaction() || !unitFound {
		service.cancelCommandNeed(active, pet, record.ID, pending.itemID)
		_ = service.SpeakLocalized(ctx, active, record, pending.kind.missingKey())
		return
	}
	if unit.Moving {
		return
	}
	if unit.Position.Point != item.Point {
		if _, err := active.MoveControlled(EntityKey(record.ID), item.Point, worldunit.ControlFurnitureInteraction); err == nil {
			return
		}
		service.cancelCommandNeed(active, pet, record.ID, pending.itemID)
		_ = service.SpeakLocalized(ctx, active, record, i18n.Key("pet.command.need.unreachable"))
		return
	}
	pet.mutex.Lock()
	if pet.commandNeed.itemID != pending.itemID || pet.needPending {
		pet.mutex.Unlock()
		return
	}
	pet.needPending = true
	pet.mutex.Unlock()
	service.dispatchCommandNeed(active, pet, record, pending)
}

// dispatchCommandNeed consumes one arrived-at product outside the room owner loop.
func (service *Service) dispatchCommandNeed(active *roomlive.Room, pet *activePet, record petrecord.Pet, pending commandNeedState) {
	roomID := active.ID()
	queued := service.dispatch(func() {
		err := service.needs.ConsumeNeed(context.Background(), roomID, record.ID, pending.itemID)
		current, found := service.rooms.Find(roomID)
		if !found || current != active {
			return
		}
		active.Schedule(0, func(_ time.Time) {
			service.finishCommandNeed(active, pet, record, pending, err)
		})
	})
	if !queued {
		service.cancelCommandNeed(active, pet, record.ID, pending.itemID)
		_ = service.SpeakLocalized(context.Background(), active, record, i18n.Key("pet.command.need.unavailable"))
	}
}

// finishCommandNeed releases control and projects the contextual result.
func (service *Service) finishCommandNeed(active *roomlive.Room, pet *activePet, record petrecord.Pet, pending commandNeedState, consumeErr error) {
	pet.mutex.Lock()
	if pet.commandNeed.itemID != pending.itemID {
		pet.mutex.Unlock()
		return
	}
	pet.needPending = false
	pet.commandNeed = commandNeedState{}
	generation := pet.actionGeneration
	pet.mutex.Unlock()
	_, _ = active.ReleaseUnitControl(EntityKey(record.ID))
	if consumeErr != nil {
		_ = service.SpeakLocalized(context.Background(), active, record, i18n.Key("pet.command.need.unavailable"))
		return
	}
	clearPetStatuses(active, record.ID)
	active.SetUnitStatus(EntityKey(record.ID), commandNeedStatus, "")
	service.projectUnitStatus(context.Background(), active, record.ID)
	_ = service.SpeakLocalized(context.Background(), active, record, pending.kind.completeKey())
	key := roomtask.Key(uint64(record.ID)<<8 | uint64(pending.actionID+1))
	active.ScheduleReplacing(key, 2*time.Second, func(time.Time) {
		pet.mutex.Lock()
		valid := pet.actionGeneration == generation
		pet.mutex.Unlock()
		if valid {
			active.ClearUnitStatus(EntityKey(record.ID), commandNeedStatus)
			service.projectUnitStatus(context.Background(), active, record.ID)
		}
	})
}

// cancelCommandNeed clears one matching contextual command reservation.
func (service *Service) cancelCommandNeed(active *roomlive.Room, pet *activePet, petID int64, itemID int64) {
	pet.mutex.Lock()
	if pet.commandNeed.itemID == itemID {
		pet.commandNeed = commandNeedState{}
		pet.needPending = false
	}
	pet.mutex.Unlock()
	_, _ = active.ReleaseUnitControl(EntityKey(petID))
}

// stampCommandCooldown applies a rejected contextual command's normal throttle.
func (service *Service) stampCommandCooldown(pet *activePet, actionID int32, command petrecord.Command) {
	pet.mutex.Lock()
	pet.cooldowns[actionID] = service.Now().Add(command.Cooldown)
	pet.mutex.Unlock()
}

// interaction returns the indexed furniture interaction for one need.
func (need CommandNeed) interaction() string {
	if need == CommandNeedDrink {
		return "pet_drink"
	}
	if need == CommandNeedFood {
		return "pet_food"
	}
	return ""
}

// missingKey returns localized missing-product feedback for one need.
func (need CommandNeed) missingKey() i18n.Key {
	if need == CommandNeedDrink {
		return i18n.Key("pet.command.need.drink.missing")
	}
	return i18n.Key("pet.command.need.food.missing")
}

// completeKey returns localized successful-consumption feedback for one need.
func (need CommandNeed) completeKey() i18n.Key {
	if need == CommandNeedDrink {
		return i18n.Key("pet.command.need.drink.complete")
	}
	return i18n.Key("pet.command.need.food.complete")
}

// needInteractions selects two bounded need interactions by current materialized stats.
func needInteractions(pet petrecord.Pet) (string, string, bool) {
	if pet.Energy < petrecord.MaximumEnergy(pet.Level)/2 {
		return "pet_food", "pet_drink", true
	}
	if pet.Happiness < 50 {
		return "pet_toy", "nest", true
	}
	return "", "", false
}
