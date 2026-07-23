package runtime

import (
	"context"

	petdismounted "github.com/niflaot/pixels/internal/realm/pet/equipment/events/dismounted"
	petmounted "github.com/niflaot/pixels/internal/realm/pet/equipment/events/mounted"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outeffect "github.com/niflaot/pixels/networking/outbound/room/entities/effect"
)

// Mount attaches or detaches one player from a saddled pet.
func (service *Service) Mount(ctx context.Context, roomID int64, petID int64, playerID int64, mount bool) (petrecord.Pet, error) {
	pet, found := service.Active(roomID, petID)
	if !found {
		return petrecord.Pet{}, petrecord.ErrPetNotFound
	}
	active, activeFound := service.rooms.Find(roomID)
	if !activeFound {
		return petrecord.Pet{}, petrecord.ErrInvalidState
	}
	if mount {
		if _, petFound := active.Unit(EntityKey(petID)); !petFound {
			return petrecord.Pet{}, petrecord.ErrInvalidState
		}
		if _, playerFound := active.Unit(playerID); !playerFound {
			return petrecord.Pet{}, petrecord.ErrInvalidState
		}
	}
	pet.mutex.Lock()
	record := pet.record
	riderPlayerID := pet.riderPlayerID
	if mount {
		if pet.stationary || !record.HasSaddle || riderPlayerID != 0 || record.OwnerPlayerID != playerID && !record.PublicRide {
			pet.mutex.Unlock()
			return petrecord.Pet{}, petrecord.ErrNoRights
		}
		pet.riderPlayerID = playerID
		riderPlayerID = playerID
	} else if riderPlayerID == playerID || record.OwnerPlayerID == playerID {
		if riderPlayerID == 0 {
			pet.mutex.Unlock()
			return record, nil
		}
		pet.riderPlayerID = 0
	} else {
		pet.mutex.Unlock()
		return petrecord.Pet{}, petrecord.ErrNoRights
	}
	pet.mutex.Unlock()
	if mount {
		if _, _, err := active.SetMount(playerID, EntityKey(petID), true); err != nil {
			pet.mutex.Lock()
			if pet.riderPlayerID == playerID {
				pet.riderPlayerID = 0
			}
			pet.mutex.Unlock()
			return petrecord.Pet{}, petrecord.ErrInvalidState
		}
		service.projectEntityStatus(ctx, active, playerID)
		service.projectUnitStatus(ctx, active, petID)
		service.projectRidingEffect(ctx, active, playerID)
	} else if riderPlayerID != 0 {
		_, _, _ = active.SetMount(riderPlayerID, EntityKey(petID), false)
		service.projectEntityStatus(ctx, active, riderPlayerID)
		service.projectUnitStatus(ctx, active, petID)
		service.projectRidingEffect(ctx, active, riderPlayerID)
	}
	service.ProjectFigure(ctx, active, record)
	if mount {
		service.Publish(ctx, petmounted.Name, petmounted.Payload{PetID: petID, RoomID: roomID, PlayerID: playerID})
	} else {
		service.Publish(ctx, petdismounted.Name, petdismounted.Payload{PetID: petID, RoomID: roomID, PlayerID: riderPlayerID})
	}
	return record, nil
}

// DismountPlayer removes one player from every loaded pet using stable room snapshots.
func (service *Service) DismountPlayer(ctx context.Context, playerID int64) {
	service.mutex.RLock()
	states := make([]struct {
		roomID int64
		state  *roomState
	}, 0, len(service.active))
	for roomID, state := range service.active {
		states = append(states, struct {
			roomID int64
			state  *roomState
		}{roomID: roomID, state: state})
	}
	service.mutex.RUnlock()
	for _, entry := range states {
		snapshot := entry.state.snapshot.Load()
		if snapshot == nil {
			continue
		}
		for _, pet := range snapshot.pets {
			pet.mutex.Lock()
			if pet.riderPlayerID != playerID {
				pet.mutex.Unlock()
				continue
			}
			pet.riderPlayerID = 0
			record := pet.record
			pet.mutex.Unlock()
			if active, found := service.rooms.Find(entry.roomID); found {
				_, _, _ = active.SetMount(playerID, EntityKey(record.ID), false)
				service.projectEntityStatus(ctx, active, playerID)
				service.projectUnitStatus(ctx, active, record.ID)
				service.projectRidingEffect(ctx, active, playerID)
				service.ProjectFigure(ctx, active, record)
			}
			service.Publish(ctx, petdismounted.Name, petdismounted.Payload{PetID: record.ID, RoomID: entry.roomID, PlayerID: playerID})
		}
	}
}

// projectRidingEffect projects Nitro's riding effect or restores the selected avatar effect.
func (service *Service) projectRidingEffect(ctx context.Context, active *roomlive.Room, playerID int64) {
	unit, found := active.Unit(playerID)
	if !found {
		return
	}
	packet, err := outeffect.Encode(unit.UnitID, roomprojection.EffectID(unit), 0)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// applyAction mutates only active room state.
func (service *Service) applyAction(active *roomlive.Room, pet *activePet, petID int64, actorID int64, action CommandAction) {
	pet.mutex.Lock()
	switch action.Mode {
	case ActionClear:
		pet.followingPlayerID, pet.stay, pet.silent = 0, false, false
		clearPetStatuses(active, petID)
	case ActionFollow:
		pet.followingPlayerID, pet.stay = actorID, false
		clearPetStatuses(active, petID)
	case ActionStay:
		pet.followingPlayerID, pet.stay = 0, true
		_, _ = active.StopMovement(EntityKey(petID))
	case ActionHere:
		pet.followingPlayerID, pet.stay = 0, false
		_ = service.follow(active, petID, actorID)
	case ActionSilent:
		pet.followingPlayerID, pet.stay, pet.silent = 0, false, true
		clearPetStatuses(active, petID)
	case ActionStatus:
		pet.followingPlayerID, pet.stay = 0, false
		_, _ = active.ReleaseUnitControl(EntityKey(petID))
		clearPetStatuses(active, petID)
		active.SetUnitStatus(EntityKey(petID), action.StatusKey, action.StatusValue)
	}
	pet.mutex.Unlock()
}

// clearPetStatuses removes persistent pet animation statuses.
func clearPetStatuses(active *roomlive.Room, petID int64) {
	for _, key := range []string{worldunit.StatusSit, worldunit.StatusLay, "beg", "ded", "jmp", "pla", "spk", "rlx", "crk", "dip", "wav", "wng", "flm", "eat", "wag", "dan", "trn", "kck"} {
		active.ClearUnitStatus(EntityKey(petID), key)
	}
}
