package runtime

import (
	"context"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

const entityBase int64 = -(1 << 62)

// rebuildSnapshot publishes a stable controller generation.
func (state *roomState) rebuildSnapshot() {
	items := make([]*activePet, 0, len(state.pets))
	for _, pet := range state.pets {
		items = append(items, pet)
	}
	state.snapshot.Store(&petSnapshot{pets: items})
}

// EnsureRoom loads and attaches placed pets exactly once.
func (service *Service) EnsureRoom(ctx context.Context, active *roomlive.Room) error {
	_, err := service.ensureRoom(ctx, active)
	return err
}

// ensureRoom loads one active generation and reports first publication.
func (service *Service) ensureRoom(ctx context.Context, active *roomlive.Room) (bool, error) {
	if active == nil {
		return false, petrecord.ErrInvalidState
	}
	roomID := active.ID()
	service.mutex.Lock()
	state := service.active[roomID]
	if state != nil {
		ready := state.ready
		service.mutex.Unlock()
		if ready != nil {
			<-ready
		}
		return false, state.loadErr
	}
	state = &roomState{ready: make(chan struct{})}
	service.active[roomID] = state
	service.mutex.Unlock()
	started := time.Now()
	defer func() { service.metrics.ObserveRoomLoad(time.Since(started)) }()
	references, err := service.currentReferences(ctx)
	if err != nil {
		service.finishRoomLoad(roomID, state, err)
		return false, err
	}
	pets, err := service.store.Room(ctx, roomID)
	if err != nil {
		service.finishRoomLoad(roomID, state, err)
		return false, err
	}
	now := service.Now()
	controllers := make(map[int64]*activePet, len(pets))
	service.mutex.Lock()
	if service.active[roomID] != state {
		service.mutex.Unlock()
		service.finishRoomLoad(roomID, state, petrecord.ErrInvalidState)
		return false, petrecord.ErrInvalidState
	}
	for _, pet := range pets {
		pet = service.materialize(pet, now)
		if pet.X == nil || pet.Y == nil || pet.Z == nil || pet.Rotation == nil {
			continue
		}
		point, valid := grid.NewPoint(*pet.X, *pet.Y)
		if !valid {
			continue
		}
		position := worldpath.Position{Point: point, Z: grid.HeightFromUnits(*pet.Z)}
		if _, addErr := active.AddEntity(EntityKey(pet.ID), pet.OwnerPlayerID, worldunit.KindPet, position, worldunit.Rotation(*pet.Rotation)); addErr != nil {
			continue
		}
		controllers[pet.ID] = &activePet{record: pet, stationary: stationarySpecies(references, pet.TypeID), nextDue: now.Add(service.decisionDelay()), lastFlush: now, lastPoint: point}
	}
	state.pets = controllers
	state.rebuildSnapshot()
	close(state.ready)
	state.ready = nil
	service.mutex.Unlock()
	for _, pet := range state.pets {
		pet.mutex.Lock()
		record := pet.record
		pet.mutex.Unlock()
		service.ProjectSpawn(ctx, active, record)
	}
	service.metrics.AddRoom(1)
	return true, nil
}

// finishRoomLoad publishes one load failure and removes its generation.
func (service *Service) finishRoomLoad(roomID int64, state *roomState, err error) {
	service.mutex.Lock()
	state.loadErr = err
	if state.ready != nil {
		close(state.ready)
		state.ready = nil
	}
	if service.active[roomID] == state {
		delete(service.active, roomID)
	}
	service.mutex.Unlock()
}

// UnloadRoom releases one active room generation.
func (service *Service) UnloadRoom(roomID int64) {
	service.mutex.Lock()
	_, found := service.active[roomID]
	delete(service.active, roomID)
	service.mutex.Unlock()
	if found {
		service.metrics.AddRoom(-1)
	}
}

// Active returns one placed pet controller without allocation.
func (service *Service) Active(roomID int64, petID int64) (*activePet, bool) {
	service.mutex.RLock()
	state := service.active[roomID]
	var pet *activePet
	if state != nil {
		pet = state.pets[petID]
	}
	service.mutex.RUnlock()
	return pet, pet != nil
}

// AddPlaced adds one newly placed pet to the active generation.
func (service *Service) AddPlaced(ctx context.Context, pet petrecord.Pet) {
	if pet.RoomID == nil {
		return
	}
	service.mutex.RLock()
	state := service.active[*pet.RoomID]
	var ready chan struct{}
	if state != nil {
		ready = state.ready
	}
	service.mutex.RUnlock()
	if ready != nil {
		<-ready
	}
	references, _ := service.currentReferences(ctx)
	service.mutex.Lock()
	state = service.active[*pet.RoomID]
	if state != nil {
		now := service.Now()
		pet = service.materialize(pet, now)
		point := grid.Point{}
		if pet.X != nil && pet.Y != nil {
			point, _ = grid.NewPoint(*pet.X, *pet.Y)
		}
		state.pets[pet.ID] = &activePet{record: pet, stationary: stationarySpecies(references, pet.TypeID), nextDue: now.Add(service.decisionDelay()), lastFlush: now, lastPoint: point}
		state.rebuildSnapshot()
	}
	service.mutex.Unlock()
}

// RemovePlaced removes one controller from the active generation.
func (service *Service) RemovePlaced(roomID int64, petID int64) {
	service.mutex.Lock()
	state := service.active[roomID]
	if state != nil {
		delete(state.pets, petID)
		state.rebuildSnapshot()
	}
	service.mutex.Unlock()
}

// ReplacePlaced refreshes one active pet after a durable mutation.
func (service *Service) ReplacePlaced(saved petrecord.Pet) {
	if saved.RoomID == nil {
		return
	}
	pet, found := service.Active(*saved.RoomID, saved.ID)
	if !found {
		return
	}
	pet.mutex.Lock()
	pet.record = service.materialize(saved, service.Now())
	pet.mutex.Unlock()
}

// PlacedCount returns the number of pets loaded in one room without allocation.
func (service *Service) PlacedCount(roomID int64) int {
	service.mutex.RLock()
	state := service.active[roomID]
	count := 0
	if state != nil {
		count = len(state.pets)
	}
	service.mutex.RUnlock()
	return count
}

// OwnerPlacedCount returns one owner's loaded pet count without allocation.
func (service *Service) OwnerPlacedCount(roomID int64, ownerID int64) int {
	count := 0
	for _, pet := range service.roomPets(roomID) {
		pet.mutex.Lock()
		if pet.record.OwnerPlayerID == ownerID {
			count++
		}
		pet.mutex.Unlock()
	}
	return count
}

// Snapshot returns one immutable active pet value.
func (service *Service) Snapshot(roomID int64, petID int64) (petrecord.Pet, bool) {
	pet, found := service.Active(roomID, petID)
	if !found {
		return petrecord.Pet{}, false
	}
	pet.mutex.Lock()
	value := pet.record
	pet.mutex.Unlock()
	return value, true
}

// roomPets returns the immutable hot-path snapshot.
func (service *Service) roomPets(roomID int64) []*activePet {
	service.mutex.RLock()
	state := service.active[roomID]
	var snapshot *petSnapshot
	if state != nil {
		snapshot = state.snapshot.Load()
	}
	service.mutex.RUnlock()
	if snapshot == nil {
		return nil
	}
	return snapshot.pets
}
