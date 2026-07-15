// Package projection maps room runtime state into outbound protocol packets.
package projection

import (
	"strconv"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outunits "github.com/niflaot/pixels/networking/outbound/room/entities/units"
)

const (
	// DefaultGender stores the fallback avatar gender.
	DefaultGender = "M"
)

// Units maps room unit snapshots to UNIT records.
func Units(room *roomlive.Room, playerIDs ...int64) []outunits.Unit {
	if room == nil {
		return nil
	}

	occupants := occupantsByID(room.Occupants())
	units := room.Units()
	allowed := allowedPlayers(playerIDs)
	records := make([]outunits.Unit, 0, len(units))
	for _, unit := range units {
		if len(allowed) > 0 && !allowed[unit.EntityKey] {
			continue
		}
		occupant, ok := occupants[unit.PlayerID]
		if !ok {
			continue
		}
		records = append(records, unitRecord(occupant, unit))
	}

	return records
}

// Statuses maps room unit snapshots to UNIT_STATUS records.
func Statuses(room *roomlive.Room, playerIDs ...int64) []outstatus.Unit {
	if room == nil {
		return nil
	}

	units := room.Units()
	allowed := allowedPlayers(playerIDs)
	records := make([]outstatus.Unit, 0, len(units))
	for _, unit := range units {
		if len(allowed) > 0 && !allowed[unit.EntityKey] {
			continue
		}
		records = append(records, statusPositionRecord(unit, unit.Position, statusActions(unit.Statuses)))
	}

	return records
}

// MovementStatuses maps movement ticks to UNIT_STATUS records.
func MovementStatuses(movements []roomlive.Movement) []outstatus.Unit {
	records := make([]outstatus.Unit, 0, len(movements))
	for _, movement := range movements {
		records = append(records, movementStatusRecord(movement))
	}

	return records
}

// unitRecord maps one live unit to one protocol unit.
func unitRecord(occupant roomlive.Occupant, unit roomlive.UnitSnapshot) outunits.Unit {
	return outunits.Unit{
		UserID:           occupant.PlayerID,
		Name:             occupant.Username,
		Motto:            occupant.Motto,
		Figure:           occupant.Figure,
		RoomIndex:        unit.UnitID,
		X:                int32(unit.Position.Point.X),
		Y:                int32(unit.Position.Point.Y),
		Z:                heightValue(unit.Position.Z),
		Direction:        int32(unit.BodyRotation),
		Gender:           genderValue(occupant.Gender),
		GroupID:          -1,
		GroupStatus:      -1,
		AchievementScore: 0,
		Moderator:        false,
	}
}

// statusRecord maps one live unit to one protocol movement status, anchored at the unit's previous
// position so the client animates the step from its origin tile. Static snapshots must use
// statusPositionRecord with the current position instead.
func statusRecord(unit roomlive.UnitSnapshot, actions []outstatus.Action) outstatus.Unit {
	return statusPositionRecord(unit, unit.Previous, actions)
}

// movementStatusRecord maps one movement tick to a protocol status.
func movementStatusRecord(movement roomlive.Movement) outstatus.Unit {
	if !movement.Moved {
		return statusPositionRecord(movement.Unit, movement.Unit.Position, statusActions(movement.Unit.Statuses))
	}

	actions := []outstatus.Action{{
		Key:   worldunit.StatusMove,
		Value: positionValue(movement.Step.Position),
	}}

	return statusRecord(movement.Unit, actions)
}

// statusPositionRecord maps one live unit position to one protocol status.
func statusPositionRecord(unit roomlive.UnitSnapshot, position worldpath.Position, actions []outstatus.Action) outstatus.Unit {
	return outstatus.Unit{
		RoomIndex:     unit.UnitID,
		X:             int32(position.Point.X),
		Y:             int32(position.Point.Y),
		Z:             heightValue(position.Z),
		HeadDirection: int32(unit.HeadRotation),
		BodyDirection: int32(unit.BodyRotation),
		Actions:       actions,
	}
}

// statusActions maps neutral unit statuses to packet actions.
func statusActions(statuses []worldunit.Status) []outstatus.Action {
	actions := make([]outstatus.Action, 0, len(statuses))
	for _, status := range statuses {
		if status.Key == worldunit.StatusDance {
			continue
		}
		actions = append(actions, outstatus.Action{Key: status.Key, Value: status.Value})
	}

	return actions
}

// occupantsByID maps occupants by player id.
func occupantsByID(occupants []roomlive.Occupant) map[int64]roomlive.Occupant {
	mapped := make(map[int64]roomlive.Occupant, len(occupants))
	for _, occupant := range occupants {
		mapped[occupant.PlayerID] = occupant
	}

	return mapped
}

// allowedPlayers maps optional player filters.
func allowedPlayers(playerIDs []int64) map[int64]bool {
	if len(playerIDs) == 0 {
		return nil
	}

	allowed := make(map[int64]bool, len(playerIDs))
	for _, playerID := range playerIDs {
		allowed[playerID] = true
	}

	return allowed
}

// genderValue returns a protocol gender value.
func genderValue(value string) string {
	if value == "" {
		return DefaultGender
	}

	return value
}

// heightValue returns a protocol height value.
func heightValue(value grid.Height) string {
	return value.String()
}

// positionValue returns a movement status position value.
func positionValue(position worldpath.Position) string {
	return strconv.Itoa(int(position.Point.X)) + "," +
		strconv.Itoa(int(position.Point.Y)) + "," +
		position.Z.String()
}
