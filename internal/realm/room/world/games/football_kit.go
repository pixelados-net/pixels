package games

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// footballClothing stores the avatar parts replaced by one football uniform.
var footballClothing = map[string]struct{}{"ch": {}, "ca": {}, "cc": {}, "cp": {}, "lg": {}, "wa": {}, "sh": {}}

// mergeFootballKit replaces football clothing parts while preserving identity parts.
func mergeFootballKit(figure string, kit string) string {
	parts := make([]string, 0, len(strings.Split(figure, "."))+len(strings.Split(kit, ".")))
	for _, part := range strings.Split(figure, ".") {
		kind, _, _ := strings.Cut(part, "-")
		if _, replaced := footballClothing[kind]; !replaced && part != "" {
			parts = append(parts, part)
		}
	}
	for _, part := range strings.Split(kit, ".") {
		kind, _, _ := strings.Cut(part, "-")
		if _, accepted := footballClothing[kind]; accepted && part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ".")
}

// toggleFootballKit equips or removes one room-only football uniform at the dressing gate.
func (service *Service) toggleFootballKit(ctx context.Context, active *roomlive.Room, playerID int64, kit string) error {
	occupant, found := active.Occupant(playerID)
	if !found || strings.TrimSpace(kit) == "" {
		return nil
	}
	service.mutex.Lock()
	state := service.stateLocked(active)
	original, equipped := state.footballLooks[playerID]
	if equipped {
		delete(state.footballLooks, playerID)
	} else {
		original = occupant.Figure
		state.footballLooks[playerID] = original
	}
	service.mutex.Unlock()
	figure := original
	if !equipped {
		figure = mergeFootballKit(original, kit)
	}
	active.UpdateOccupantProfile(playerID, figure, occupant.Gender, occupant.Motto)
	return broadcast.RoomSpawn(ctx, service.connections, active, playerID, 0)
}
