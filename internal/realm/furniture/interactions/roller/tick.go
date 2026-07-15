package roller

import (
	"context"
	"sort"
	"time"

	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/bus"
)

// Cycle advances every due roller in one active room.
func (service *Service) Cycle(ctx context.Context, active *roomlive.Room, _ time.Time) error {
	rollers := active.FurnitureByInteraction("roller")
	if !active.AdvanceRollerCycle(len(rollers) > 0) {
		return nil
	}
	rolledItems := make(map[int64]struct{})
	rolledUnits := make(map[int64]struct{})
	for _, roller := range rollers {
		candidate, valid := service.prepare(active, roller, rolledItems, rolledUnits)
		if !valid {
			continue
		}
		moved := service.apply(active, candidate, rolledItems, rolledUnits)
		if len(moved.items) == 0 && len(moved.units) == 0 {
			continue
		}
		service.scheduleHooks(active, moved)
		if err := service.broadcast(ctx, active, moved); err != nil {
			return err
		}
	}
	return nil
}

// prepare validates one roller and captures eligible mounted entities.
func (service *Service) prepare(active *roomlive.Room, roller worldfurniture.Item, rolledItems map[int64]struct{}, rolledUnits map[int64]struct{}) (step, bool) {
	target, valid := grid.PointInFront(roller.Point, uint8(roller.Rotation))
	if !valid {
		return step{}, false
	}
	column, err := active.SurfaceColumn(target)
	if err != nil {
		return step{}, false
	}
	targetItems := active.FurnitureAt(target)
	unitsInRoom := active.Units()
	if destinationOccupied(unitsInRoom, target) {
		return step{}, false
	}
	targetHeight, stacking, found := destinationSurface(column)
	if !found {
		return step{}, false
	}
	targetRoller, chained := rollerAt(targetItems)
	if chained && !service.validChain(roller, targetRoller, targetItems) {
		return step{}, false
	}
	if chained {
		targetHeight = targetRoller.Top()
	}
	offset := targetHeight - roller.Top()
	items := mountedItems(active.FurnitureAt(roller.Point), roller, rolledItems)
	units := mountedUnits(unitsInRoom, roller, rolledUnits, service.config.MaxAvatarsPerTick)
	if len(items) == 0 && len(units) == 0 {
		return step{}, false
	}
	if !allowsUsers(targetItems) {
		units = nil
	}
	if !stacking || offset > 0 {
		items = nil
	}
	return step{
		roller: roller, target: target, offset: offset, units: units, items: items,
		sourceTop: topItemID(active.FurnitureAt(roller.Point)), targetTop: topItemID(targetItems),
	}, len(items) > 0 || len(units) > 0
}

// apply mutates validated furniture and units while recording tick deduplication.
func (service *Service) apply(active *roomlive.Room, candidate step, rolledItems map[int64]struct{}, rolledUnits map[int64]struct{}) movedStep {
	moved := movedStep{step: candidate}
	dx := int(candidate.target.X) - int(candidate.roller.Point.X)
	dy := int(candidate.target.Y) - int(candidate.roller.Point.Y)
	for _, item := range candidate.items {
		source := item
		point, valid := grid.NewPoint(int(item.Point.X)+dx, int(item.Point.Y)+dy)
		if !valid {
			continue
		}
		item.Point, item.Z = point, item.Z+candidate.offset
		if _, err := active.ReloadFurniture(item.ID, &item); err != nil {
			continue
		}
		rolledItems[item.ID] = struct{}{}
		moved.itemSources = append(moved.itemSources, source)
		moved.items = append(moved.items, item)
		service.enqueuePersistence(active.ID(), item)
	}
	for _, unit := range candidate.units {
		position := worldpath.Position{Point: candidate.target, Z: unit.Position.Z + candidate.offset}
		updated, err := active.RollUnit(unit.EntityKey, position)
		if err != nil {
			continue
		}
		rolledUnits[unit.EntityKey] = struct{}{}
		moved.unitSources = append(moved.unitSources, unit)
		moved.units = append(moved.units, updated)
	}
	return moved
}

// mountedItems returns descending mounted furniture not already rolled this tick.
func mountedItems(items []worldfurniture.Item, roller worldfurniture.Item, rolled map[int64]struct{}) []worldfurniture.Item {
	mounted := make([]worldfurniture.Item, 0, len(items))
	for _, item := range items {
		if item.ID == roller.ID || item.Z < roller.Top() {
			continue
		}
		if _, duplicate := rolled[item.ID]; duplicate {
			continue
		}
		mounted = append(mounted, item)
	}
	sort.Slice(mounted, func(left int, right int) bool { return mounted[left].Z > mounted[right].Z })
	return mounted
}

// mountedUnits returns at most max stationary player units on the roller.
func mountedUnits(units []roomlive.UnitSnapshot, roller worldfurniture.Item, rolled map[int64]struct{}, max int) []roomlive.UnitSnapshot {
	mounted := make([]roomlive.UnitSnapshot, 0, max)
	for _, unit := range units {
		if len(mounted) >= max || unit.Kind != worldunit.KindPlayer || unit.Moving || unit.Position.Point != roller.Point || unit.Position.Z < roller.Top() {
			continue
		}
		if _, duplicate := rolled[unit.EntityKey]; duplicate {
			continue
		}
		mounted = append(mounted, unit)
	}
	return mounted
}

// destinationSurface returns the physical top and stacking policy.
func destinationSurface(column surface.Column) (grid.Height, bool, bool) {
	section, found := column.TopSection()
	if !found {
		return 0, false, false
	}
	return section.Top(), section.Stacking(), true
}

// destinationOccupied reports whether any room unit blocks a target tile.
func destinationOccupied(units []roomlive.UnitSnapshot, target grid.Point) bool {
	for _, unit := range units {
		if unit.Position.Point == target {
			return true
		}
	}
	return false
}

// rollerAt finds a target roller.
func rollerAt(items []worldfurniture.Item) (worldfurniture.Item, bool) {
	for _, item := range items {
		if item.Definition.InteractionType == "roller" {
			return item, true
		}
	}
	return worldfurniture.Item{}, false
}

// validChain enforces equal-height topmost roller chaining unless rules are disabled.
func (service *Service) validChain(source worldfurniture.Item, target worldfurniture.Item, items []worldfurniture.Item) bool {
	if service.config.NoRules {
		return true
	}
	return source.Z == target.Z && topItemID(items) == target.ID
}

// allowsUsers reports whether destination furniture permits rolled avatars.
func allowsUsers(items []worldfurniture.Item) bool {
	for _, item := range items {
		if item.Definition.AllowWalk || item.Definition.AllowSit || item.Definition.AllowLay || openGate(item) {
			continue
		}
		return false
	}
	return true
}

// openGate reports whether a gate exposes its passable state.
func openGate(item worldfurniture.Item) bool {
	if item.Definition.InteractionType != "gate" {
		return false
	}
	if item.Definition.CustomParams == "open_state=0" {
		return item.ExtraData == "0"
	}
	return item.ExtraData == "1"
}

// topItemID returns the physically highest item id on a tile.
func topItemID(items []worldfurniture.Item) int64 {
	var selected int64
	var height grid.Height
	for _, item := range items {
		if selected == 0 || item.Top() > height {
			selected, height = item.ID, item.Top()
		}
	}
	return selected
}

// scheduleHooks defers walk transitions until the client roll animation settles.
func (service *Service) scheduleHooks(active *roomlive.Room, moved movedStep) {
	if service.events == nil || moved.step.sourceTop == moved.step.targetTop {
		return
	}
	for _, unit := range moved.units {
		playerID := unit.PlayerID
		if playerID <= 0 {
			continue
		}
		active.Schedule(service.config.Delay, func(time.Time) {
			ctx := context.Background()
			if moved.step.sourceTop > 0 {
				_ = service.events.Publish(ctx, bus.Event{Name: furniturewalkedoff.Name, Payload: furniturewalkedoff.Payload{RoomID: active.ID(), ItemID: moved.step.sourceTop, PlayerID: playerID}})
			}
			if moved.step.targetTop > 0 {
				_ = service.events.Publish(ctx, bus.Event{Name: furniturewalkedon.Name, Payload: furniturewalkedon.Payload{RoomID: active.ID(), ItemID: moved.step.targetTop, PlayerID: playerID}})
			}
		})
	}
}
