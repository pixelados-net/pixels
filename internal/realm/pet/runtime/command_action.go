package runtime

import (
	"context"
	"time"

	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomtask "github.com/niflaot/pixels/internal/realm/room/runtime/live/task"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// Metrics returns the shared lock-free pet telemetry collector.
func (service *Service) Metrics() *petobservability.Metrics { return service.metrics }

// decisionDelay returns one bounded autonomous decision delay without allocation.
func (service *Service) decisionDelay() time.Duration {
	minimum, maximum := service.config.DecisionMinimum, service.config.DecisionMaximum
	if maximum <= minimum {
		return minimum
	}
	return minimum + time.Duration(service.source.Uint64()%uint64(maximum-minimum+1))
}

// stationarySpecies reports whether one immutable species cannot initiate movement.
func stationarySpecies(references *petreference.Snapshot, typeID int32) bool {
	return references != nil && typeID >= 0 && typeID < int32(len(references.Species)) && references.SpeciesPresent[typeID] && references.Species[typeID].Plant
}

// currentReferences returns the warmed reference generation when configured.
func (service *Service) currentReferences(ctx context.Context) (*petreference.Snapshot, error) {
	if service.references == nil {
		return nil, nil
	}
	return service.references.Current(ctx)
}

// MovePet directs one mobile pet through the shared room world.
func (service *Service) MovePet(active *roomlive.Room, petID int64, point grid.Point) error {
	if active == nil {
		return petrecord.ErrInvalidState
	}
	pet, found := service.Active(active.ID(), petID)
	if !found {
		return petrecord.ErrPetNotFound
	}
	pet.mutex.Lock()
	stationary := pet.stationary
	pet.mutex.Unlock()
	if stationary {
		return petrecord.ErrInvalidState
	}
	_, err := active.MoveTo(EntityKey(petID), point)
	return err
}

// Select records the player currently directing one active pet.
func (service *Service) Select(roomID int64, petID int64, playerID int64) bool {
	pet, found := service.Active(roomID, petID)
	if !found {
		return false
	}
	pet.mutex.Lock()
	pet.selectedBy = playerID
	pet.actionGeneration++
	pet.mutex.Unlock()
	return true
}

// ActionMode identifies one command's world behavior.
type ActionMode uint8

const (
	// ActionStatus projects one temporary renderer status.
	ActionStatus ActionMode = iota
	// ActionClear clears posture, follow, and stay.
	ActionClear
	// ActionFollow follows the commanding player.
	ActionFollow
	// ActionStay stops autonomous movement.
	ActionStay
	// ActionHere walks toward the commanding player.
	ActionHere
	// ActionSilent suppresses autonomous vocalizations until cleared.
	ActionSilent
	// ActionNeed walks to and consumes one contextual pet product.
	ActionNeed
)

// CommandNeed identifies one contextual pet product category.
type CommandNeed uint8

const (
	// CommandNeedNone identifies an action without a contextual product.
	CommandNeedNone CommandNeed = iota
	// CommandNeedDrink identifies pet drink furniture.
	CommandNeedDrink
	// CommandNeedFood identifies pet food furniture.
	CommandNeedFood
)

// CommandAction describes one immutable pet command action.
type CommandAction struct {
	// ID identifies the command.
	ID int32
	// Mode identifies world behavior.
	Mode ActionMode
	// StatusKey stores the renderer status key.
	StatusKey string
	// StatusValue stores the renderer status value.
	StatusValue string
	// Duration stores temporary status lifetime.
	Duration time.Duration
	// Need identifies the contextual product required by this action.
	Need CommandNeed
}

// ExecuteAction validates and applies one learned pet command.
func (service *Service) ExecuteAction(ctx context.Context, roomID int64, petID int64, actorID int64, action CommandAction, command petrecord.Command) error {
	pet, found := service.Active(roomID, petID)
	if !found || action.ID < 0 || action.ID >= int32(len(pet.cooldowns)) {
		return petrecord.ErrPetNotFound
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return petrecord.ErrInvalidState
	}
	if action.Mode == ActionNeed {
		return service.executeCommandNeed(ctx, active, pet, petID, actorID, action, command)
	}
	now := service.Now()
	pet.mutex.Lock()
	before := pet.record
	if (pet.stationary && action.moves()) || before.Level < command.RequiredLevel || before.Energy < command.EnergyCost || before.Happiness < command.HappinessCost || now.Before(pet.cooldowns[action.ID]) || pet.needPending {
		pet.mutex.Unlock()
		return petrecord.ErrInvalidState
	}
	hadCommandNeed := pet.commandNeed.itemID != 0
	pet.commandNeed = commandNeedState{}
	pet.mutex.Unlock()
	if hadCommandNeed {
		_, _ = active.ReleaseUnitControl(EntityKey(petID))
	}
	after := before
	changedStats := command.EnergyCost != 0 || command.HappinessCost != 0 || command.ExperienceReward != 0
	if changedStats {
		var updated bool
		var err error
		after, updated, err = service.store.UpdateStats(ctx, petID, -command.EnergyCost, -command.HappinessCost, command.ExperienceReward, before.Version)
		if err != nil {
			return err
		}
		if !updated {
			return petrecord.ErrConflict
		}
	}
	pet.mutex.Lock()
	pet.actionGeneration++
	generation := pet.actionGeneration
	pet.cooldowns[action.ID] = now.Add(command.Cooldown)
	pet.selectedBy = actorID
	if changedStats {
		pet.record = service.materialize(after, now)
	}
	pet.mutex.Unlock()
	service.applyAction(active, pet, petID, actorID, action)
	service.projectUnitStatus(ctx, active, petID)
	if action.Duration > 0 && action.StatusKey != "" {
		key := roomtask.Key(uint64(petID)<<8 | uint64(action.ID+1))
		active.ScheduleReplacing(key, action.Duration, func(time.Time) {
			current, currentFound := service.Active(roomID, petID)
			if !currentFound {
				return
			}
			current.mutex.Lock()
			valid := current.actionGeneration == generation
			current.mutex.Unlock()
			if valid {
				active.ClearUnitStatus(EntityKey(petID), action.StatusKey)
				service.projectUnitStatus(context.Background(), active, petID)
			}
		})
	}
	if !changedStats {
		return nil
	}
	service.ProjectStatChange(ctx, active, before, after, command.ExperienceReward, false)
	return nil
}

// moves reports whether one command initiates pet locomotion.
func (action CommandAction) moves() bool {
	return action.Mode == ActionFollow || action.Mode == ActionHere || action.Mode == ActionNeed
}
