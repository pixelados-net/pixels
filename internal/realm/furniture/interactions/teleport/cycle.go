package teleport

import (
	"context"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

const phaseDelay = 500 * time.Millisecond

// startApproach walks toward the source or reports a soft rejected approach.
func (service *Service) startApproach(ctx context.Context, active *roomlive.Room, transit Transit) (bool, error) {
	unit, found := active.Unit(transit.PlayerID)
	if !found {
		return false, service.fail(ctx, active.ID(), transit, "unit_not_found")
	}
	front, valid := frontPoint(transit.Source)
	if !valid || unit.Position.Point == front || unit.Position.Point == transit.Source.Point {
		return true, nil
	}
	if _, err := active.MoveTo(transit.PlayerID, front); err != nil {
		return false, service.fail(ctx, active.ID(), transit, "approach_unreachable")
	}

	return true, nil
}

// Cycle advances due transitions for one room owner tick.
func (service *Service) Cycle(ctx context.Context, active *roomlive.Room, now time.Time) error {
	loaded, found := service.rooms.Load(active.ID())
	if !found {
		return nil
	}
	state := loaded.(*roomState)
	var storage [8]int64
	playerIDs := storage[:0]
	state.mutex.Lock()
	for playerID := range state.transits {
		playerIDs = append(playerIDs, playerID)
	}
	state.mutex.Unlock()
	var result error
	for _, playerID := range playerIDs {
		state.mutex.Lock()
		transit, exists := state.transits[playerID]
		state.mutex.Unlock()
		if !exists {
			continue
		}
		if transit.Phase == PhaseResolving {
			continue
		}
		complete := false
		for range 6 {
			if transit.Phase != PhaseApproach && now.Before(transit.Deadline) {
				break
			}
			var stepErr error
			transit, complete, stepErr = service.advance(ctx, active, transit, now)
			result = firstError(result, stepErr)
			if complete || delayFor(transit) > 0 {
				break
			}
		}
		if complete {
			state.mutex.Lock()
			delete(state.transits, playerID)
			state.mutex.Unlock()
			if !transit.Handoff {
				service.releasePair(transit)
			}
			if transit.NextItemID > 0 {
				result = firstError(result, service.Start(ctx, StartRequest{
					PlayerID: transit.PlayerID, Room: active, ItemID: transit.NextItemID,
				}))
			}
		} else {
			state.mutex.Lock()
			state.transits[playerID] = transit
			state.mutex.Unlock()
		}
	}
	state.mutex.Lock()
	empty := len(state.transits) == 0
	state.mutex.Unlock()
	if empty {
		service.rooms.Delete(active.ID())
	}

	return result
}

// advance advances one transition without allocating a scheduler task.
func (service *Service) advance(ctx context.Context, active *roomlive.Room, transit Transit, now time.Time) (Transit, bool, error) {
	switch transit.Phase {
	case PhaseApproach:
		unit, found := active.Unit(transit.PlayerID)
		front, valid := frontPoint(transit.Source)
		if !found {
			return transit, true, service.fail(ctx, active.ID(), transit, "unit_left")
		}
		if unit.Moving || (valid && unit.Position.Point != front && unit.Position.Point != transit.Source.Point) {
			return transit, false, nil
		}
		if err := service.openSource(ctx, active, transit); err != nil {
			return transit, true, err
		}
		transit.Phase = PhaseEnter
		return transit, false, nil
	case PhaseEnter:
		unit, found := active.Unit(transit.PlayerID)
		if !found {
			return transit, true, service.fail(ctx, active.ID(), transit, "unit_left")
		}
		if unit.Moving || unitHasStatus(unit, worldunit.StatusMove) {
			return transit, false, nil
		}
		transit.Phase, transit.Deadline = PhaseCross, now.Add(delayFor(transit))
		return transit, false, nil
	case PhaseCross:
		crossRoom, err := service.cross(ctx, active, transit)
		if err != nil {
			return transit, true, err
		}
		if crossRoom {
			transit.Phase, transit.Deadline = PhaseForward, now.Add(delayFor(transit))
			return transit, false, nil
		}
		transit.Phase, transit.Deadline = PhaseExit, now.Add(delayFor(transit))
		return transit, false, nil
	case PhaseForward:
		if err := service.closeSource(ctx, active, transit); err != nil {
			return transit, true, err
		}
		if err := service.forward(ctx, active, transit); err != nil {
			return transit, true, err
		}
		transit.Handoff = service.pendingFor(transit.PlayerID, transit.TargetRoomID)
		return transit, true, nil
	case PhaseArrival:
		if err := service.showArrival(ctx, active, transit); err != nil {
			return transit, true, err
		}
		transit.Phase, transit.Deadline = PhaseExit, now.Add(delayFor(transit))
		return transit, false, nil
	case PhaseExit:
		if delayFor(transit) == 0 {
			if next, found := active.OtherInteractionAt(transit.Target.Point, transit.Target.ID); found && next.Definition.InteractionType == "teleport_tile" {
				transit.NextItemID = next.ID
				return transit, true, service.finish(ctx, active, transit)
			}
		}
		if err := service.beginExit(ctx, active, transit); err != nil {
			return transit, true, service.finish(ctx, active, transit)
		}
		transit.Phase = PhaseSettle
		return transit, false, nil
	case PhaseSettle:
		unit, found := active.Unit(transit.PlayerID)
		if !found {
			return transit, true, service.fail(ctx, active.ID(), transit, "unit_left")
		}
		if unit.Moving || unitHasStatus(unit, worldunit.StatusMove) {
			return transit, false, nil
		}
		return transit, true, service.finish(ctx, active, transit)
	default:
		return transit, true, service.fail(ctx, active.ID(), transit, "invalid_phase")
	}
}

// unitHasStatus reports whether a room unit snapshot contains one status key.
func unitHasStatus(unit roomlive.UnitSnapshot, key string) bool {
	for _, status := range unit.Statuses {
		if status.Key == key {
			return true
		}
	}

	return false
}

// firstError preserves the first cycle failure without allocating an error aggregate.
func firstError(current error, next error) error {
	if current != nil {
		return current
	}

	return next
}

// delayFor returns the visual phase delay for one teleport variant.
func delayFor(transit Transit) time.Duration {
	if transit.Source.Definition.InteractionType == "teleport_tile" {
		return 0
	}

	return phaseDelay
}
