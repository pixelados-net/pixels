package runtime

import (
	"context"
	"time"

	petdismounted "github.com/niflaot/pixels/internal/realm/pet/equipment/events/dismounted"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"go.uber.org/zap"
)

// Cycle advances due pets from the room's existing owner tick.
func (service *Service) Cycle(ctx context.Context, active *roomlive.Room, now time.Time) error {
	if err := service.EnsureRoom(ctx, active); err != nil {
		return err
	}
	dueStarted := time.Time{}
	for _, pet := range service.roomPets(active.ID()) {
		pet.mutex.Lock()
		riderPlayerID := pet.riderPlayerID
		record := pet.record
		stationary := pet.stationary
		commandNeedItemID := pet.commandNeed.itemID
		if riderPlayerID != 0 {
			pet.mutex.Unlock()
			service.syncRider(ctx, active, pet, record.ID, riderPlayerID, now)
			continue
		}
		if commandNeedItemID != 0 {
			pet.mutex.Unlock()
			if stationary {
				service.cancelCommandNeed(active, pet, record.ID, commandNeedItemID)
				continue
			}
			service.resolveCommandNeed(ctx, active, pet, record)
			continue
		}
		if now.Before(pet.nextDue) {
			pet.mutex.Unlock()
			continue
		}
		if dueStarted.IsZero() {
			dueStarted = time.Now()
		}
		pet.nextDue = now.Add(service.decisionDelay())
		followingPlayerID := pet.followingPlayerID
		stay := pet.stay
		silent := pet.silent
		nextVocal := pet.nextVocal
		pet.mutex.Unlock()
		unit, found := active.UnitMotion(EntityKey(record.ID))
		if !found {
			continue
		}
		if stationary {
			species := service.species(ctx, record)
			plant := record.DerivePlantState(now, species)
			silent = silent || plant.Dead
			if service.syncPlantStatus(ctx, active, record, species, plant, false) {
				service.projectStatus(ctx, active, record, unit, species, plant)
			}
		}
		if unit.Moving || stay {
			continue
		}
		if !stationary && service.resolveNeed(ctx, active, pet, record, unit) {
			service.metrics.RecordDecision(petobservability.DecisionNeed)
			continue
		}
		if !stationary && followingPlayerID != 0 && service.follow(active, record.ID, followingPlayerID) {
			service.metrics.RecordDecision(petobservability.DecisionFollow)
			service.capturePosition(active, pet, unit, now)
			continue
		}
		if !silent && !now.Before(nextVocal) && service.source.Uint64()%8 == 0 {
			cooldown, spoken, vocalErr := service.vocalize(ctx, active, record)
			if vocalErr != nil && service.log != nil {
				service.log.Debug("pet vocal", zap.Int64("pet_id", record.ID), zap.Error(vocalErr))
			}
			if spoken {
				service.metrics.RecordDecision(petobservability.DecisionVocal)
				pet.mutex.Lock()
				pet.nextVocal = now.Add(cooldown)
				pet.mutex.Unlock()
				continue
			}
		}
		if !stationary {
			if goal, valid := active.RandomWalkablePoint(EntityKey(record.ID), service.config.WalkRadius, service.source.Uint64()); valid {
				moveErr := service.MovePet(active, record.ID, goal)
				service.metrics.RecordDecision(petobservability.DecisionWander)
				service.metrics.RecordPathError(moveErr)
			}
		}
		service.capturePosition(active, pet, unit, now)
	}
	if !dueStarted.IsZero() {
		service.metrics.ObserveBehaviorDue(time.Since(dueStarted))
	}
	return nil
}

// resolveNeed moves toward or consumes one indexed need target without blocking the room tick.
func (service *Service) resolveNeed(_ context.Context, active *roomlive.Room, pet *activePet, record petrecord.Pet, unit roomlive.UnitSnapshot) bool {
	if service.needs == nil || !active.Snapshot().AllowPetsEat {
		return false
	}
	primary, secondary, needed := needInteractions(record)
	if !needed {
		return false
	}
	item, found := active.NearestFurnitureByInteraction(primary, unit.Position.Point, service.config.WalkRadius)
	if !found {
		item, found = active.NearestFurnitureByInteraction(secondary, unit.Position.Point, service.config.WalkRadius)
	}
	if !found {
		return false
	}
	if unit.Position.Point != item.Point {
		err := service.MovePet(active, record.ID, item.Point)
		service.metrics.RecordPathError(err)
		return err == nil
	}
	pet.mutex.Lock()
	if pet.needPending {
		pet.mutex.Unlock()
		return true
	}
	pet.needPending = true
	pet.mutex.Unlock()
	roomID, petID, itemID := active.ID(), record.ID, item.ID
	queued := service.dispatch(func() {
		current, currentFound := service.rooms.Find(roomID)
		if !currentFound || current != active {
			return
		}
		_ = service.needs.ConsumeNeed(context.Background(), roomID, petID, itemID)
		current, currentFound = service.rooms.Find(roomID)
		if !currentFound || current != active {
			return
		}
		active.Schedule(0, func(time.Time) {
			controller, controllerFound := service.Active(roomID, petID)
			if controllerFound {
				controller.mutex.Lock()
				controller.needPending = false
				controller.mutex.Unlock()
			}
		})
	})
	if !queued {
		pet.mutex.Lock()
		pet.needPending = false
		pet.mutex.Unlock()
	}
	return true
}

// syncRider validates one room-owned mount and persists its shared physical position.
func (service *Service) syncRider(ctx context.Context, active *roomlive.Room, pet *activePet, petID int64, riderPlayerID int64, now time.Time) {
	rider, riderFound := active.UnitMotion(riderPlayerID)
	unit, petFound := active.UnitMotion(EntityKey(petID))
	if !riderFound || !petFound {
		pet.mutex.Lock()
		pet.riderPlayerID = 0
		record := pet.record
		pet.mutex.Unlock()
		if riderFound {
			_, _, _ = active.SetMount(riderPlayerID, EntityKey(petID), false)
			service.projectEntityStatus(ctx, active, riderPlayerID)
			service.projectRidingEffect(ctx, active, riderPlayerID)
		}
		service.ProjectFigure(ctx, active, record)
		service.Publish(ctx, petdismounted.Name, petdismounted.Payload{PetID: petID, RoomID: active.ID(), PlayerID: riderPlayerID})
		return
	}
	if rider.RenderOffset != worldunit.RidingHeightOffset || unit.Position != rider.Position || unit.BodyRotation != rider.BodyRotation {
		repairedRider, repairedPet, err := active.SetMount(riderPlayerID, EntityKey(petID), true)
		if err == nil {
			rider, unit = repairedRider, repairedPet
			service.projectEntityStatus(ctx, active, riderPlayerID)
			service.projectUnitStatus(ctx, active, petID)
			service.projectRidingEffect(ctx, active, riderPlayerID)
		}
	}
	service.capturePosition(active, pet, unit, now)
}

// follow moves one pet toward an adjacent tile around its live target.
func (service *Service) follow(active *roomlive.Room, petID int64, playerID int64) bool {
	target, found := active.Unit(playerID)
	if !found {
		return false
	}
	for rotation := uint8(0); rotation < 8; rotation++ {
		point, valid := grid.PointInFront(target.Position.Point, rotation)
		if !valid {
			continue
		}
		if err := service.MovePet(active, petID, point); err == nil {
			service.metrics.RecordPath(petobservability.PathMoved)
			return true
		}
	}
	service.metrics.RecordPath(petobservability.PathBlocked)
	return false
}

// capturePosition coalesces changed positions onto the shared persistence pool.
func (service *Service) capturePosition(active *roomlive.Room, pet *activePet, unit roomlive.UnitSnapshot, now time.Time) {
	roomID := active.ID()
	pet.mutex.Lock()
	if unit.Position.Point == pet.lastPoint || now.Sub(pet.lastFlush) < service.config.PositionFlushInterval {
		pet.mutex.Unlock()
		return
	}
	pet.lastPoint, pet.lastFlush = unit.Position.Point, now
	petID, version := pet.record.ID, pet.record.Version
	pet.mutex.Unlock()
	queued := service.dispatch(func() {
		current, roomFound := service.rooms.Find(roomID)
		controller, petFound := service.Active(roomID, petID)
		if !roomFound || current != active || !petFound || controller != pet {
			return
		}
		next, saved, err := service.store.SavePosition(context.Background(), petID, roomID, int(unit.Position.Point.X), int(unit.Position.Point.Y), unit.Position.Z.Units(), int16(unit.BodyRotation), version)
		if err != nil {
			service.metrics.RecordStatFlush(petobservability.ResultFailed)
			if service.log != nil {
				service.log.Debug("save pet position", zap.Int64("pet_id", petID), zap.Error(err))
			}
			return
		}
		if saved {
			service.metrics.RecordStatFlush(petobservability.ResultSuccess)
			pet.mutex.Lock()
			if pet.record.Version == version {
				pet.record.Version = next
			}
			pet.mutex.Unlock()
		}
	})
	if !queued {
		service.metrics.RecordStatFlush(petobservability.ResultRejected)
	}
}
