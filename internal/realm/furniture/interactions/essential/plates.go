package essential

import (
	"context"
	"strconv"
	"time"

	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/pkg/bus"
)

// walkedOn handles occupancy-driven interaction entry.
func (service *Service) walkedOn(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(furniturewalkedon.Payload)
	if !ok {
		return nil
	}
	active, item, found := service.walkItem(payload.RoomID, payload.ItemID)
	if !found {
		return nil
	}
	switch item.Definition.InteractionType {
	case "pressureplate":
		service.schedulePressure(ctx, active, item)
	case "colorplate":
		return service.changeColorPlate(ctx, active, item, 1)
	case "handitem_tile":
		return service.giveHandItem(ctx, active, payload.PlayerID, item, false)
	}

	return nil
}

// walkedOff handles occupancy-driven interaction exit.
func (service *Service) walkedOff(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(furniturewalkedoff.Payload)
	if !ok {
		return nil
	}
	active, item, found := service.walkItem(payload.RoomID, payload.ItemID)
	if !found {
		return nil
	}
	switch item.Definition.InteractionType {
	case "pressureplate":
		service.schedulePressure(ctx, active, item)
	case "colorplate":
		return service.changeColorPlate(ctx, active, item, -1)
	}

	return nil
}

// walkItem resolves one movement event target.
func (service *Service) walkItem(roomID int64, itemID int64) (*roomlive.Room, worldfurniture.Item, bool) {
	active, found := service.runtime.Find(roomID)
	if !found {
		return nil, worldfurniture.Item{}, false
	}
	item, found := active.FurnitureItem(itemID)

	return active, item, found
}

// schedulePressure debounces one pressure plate occupancy refresh.
func (service *Service) schedulePressure(ctx context.Context, active *roomlive.Room, item worldfurniture.Item) {
	async := context.WithoutCancel(ctx)
	active.ScheduleReplacing(scheduledKey(item.ID, 2), 100*time.Millisecond, func(time.Time) {
		current, found := active.FurnitureItem(item.ID)
		if !found {
			return
		}
		occupied := active.HasUnitInFurnitureFootprint(current)
		value := "0"
		if occupied {
			value = "1"
		}
		_ = service.visual(async, active, current.ID, value)
	})
}

// changeColorPlate applies one bounded occupancy delta.
func (service *Service) changeColorPlate(ctx context.Context, active *roomlive.Room, item worldfurniture.Item, delta int) error {
	current, err := strconv.Atoi(item.ExtraData)
	if err != nil {
		current = 0
	}
	maximum := item.Definition.InteractionModesCount - 1
	if maximum < 0 {
		maximum = 0
	}
	next := current + delta
	if next < 0 {
		next = 0
	}
	if next > maximum {
		next = maximum
	}

	return service.visual(ctx, active, item.ID, strconv.Itoa(next))
}
