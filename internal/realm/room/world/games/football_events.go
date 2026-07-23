package games

import (
	"context"

	furnituremoved "github.com/niflaot/pixels/internal/realm/furniture/events/moved"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/pkg/bus"
)

// queueFootball starts one kick unless the ball is completing a goal reset.
func (service *Service) queueFootball(active *roomlive.Room, itemID int64, direction uint8, kickerID int64) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	state := service.states[active.ID()]
	if state == nil {
		return
	}
	if ball := state.footballs[itemID]; ball != nil && ball.resetting {
		return
	}
	state.footballs[itemID] = &footballBall{direction: direction, remaining: 6, kickerID: kickerID}
}

// FurnitureMoved promotes a manually placed football position to its next kickoff point.
func (service *Service) FurnitureMoved(_ context.Context, event bus.Event) error {
	payload, ok := event.Payload.(furnituremoved.Payload)
	if !ok || !service.config.Enabled {
		return nil
	}
	active, found := service.rooms.Find(payload.RoomID)
	if !found {
		return nil
	}
	item, found := active.FurnitureItem(payload.ItemID)
	if !found || item.Definition.InteractionType != "football" {
		return nil
	}
	service.mutex.Lock()
	state := service.stateLocked(active)
	state.footballOrigins[item.ID] = item.Point
	if ball := state.footballs[item.ID]; ball != nil {
		ball.remaining = 0
		ball.resetting = false
	}
	service.mutex.Unlock()
	return nil
}
