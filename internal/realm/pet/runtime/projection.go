package runtime

import (
	"context"

	petleveled "github.com/niflaot/pixels/internal/realm/pet/care/events/leveled"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/connection"
	outadd "github.com/niflaot/pixels/networking/outbound/inventory/pet/add"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/pet/list"
	outreceived "github.com/niflaot/pixels/networking/outbound/inventory/pet/received"
	outremoveinventory "github.com/niflaot/pixels/networking/outbound/inventory/pet/remove"
	outlevelnotification "github.com/niflaot/pixels/networking/outbound/notification/pet/level"
	outremoved "github.com/niflaot/pixels/networking/outbound/room/entities/removed"
	outunits "github.com/niflaot/pixels/networking/outbound/room/entities/units"
	outexperience "github.com/niflaot/pixels/networking/outbound/room/pet/experience"
	outlevel "github.com/niflaot/pixels/networking/outbound/room/pet/level"
	outrespected "github.com/niflaot/pixels/networking/outbound/room/pet/respected"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/pet/status"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// InventoryPet maps durable data to Nitro inventory data.
func InventoryPet(pet petrecord.Pet) petdata.Pet {
	parts := make([]petdata.CustomPart, len(pet.Parts))
	for index, part := range pet.Parts {
		parts[index] = petdata.CustomPart{LayerID: part.LayerID, PartID: part.PartID, PaletteID: part.PaletteID}
	}
	paletteID, color := pet.PaletteID, pet.Color
	if pet.TypeID == 16 {
		paletteID, color = 0, "FFFFFF"
	}
	return petdata.Pet{ID: pet.ID, Name: pet.Name, Level: pet.Level, Figure: petdata.Figure{TypeID: pet.TypeID, PaletteID: paletteID, Color: color, BreedID: pet.BreedID, CustomParts: parts}}
}

// ProjectStatChange broadcasts experience, level, and respect projections after commit.
func (service *Service) ProjectStatChange(ctx context.Context, active *roomlive.Room, before petrecord.Pet, after petrecord.Pet, experience int32, respected bool) {
	service.ReplacePlaced(after)
	unit, found := active.Unit(EntityKey(after.ID))
	if !found {
		return
	}
	if experience != 0 {
		if packet, err := outexperience.Encode(after.ID, unit.UnitID, experience); err == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
	}
	if before.Level != after.Level {
		if packet, err := outlevel.Encode(unit.UnitID, after.ID, after.Level); err == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
		if packet, err := outlevelnotification.Encode(after.ID, after.Name, after.Level, InventoryPet(after).Figure); err == nil {
			service.sendPlayer(ctx, after.OwnerPlayerID, packet)
		}
		service.Publish(ctx, petleveled.Name, petleveled.Payload{PetID: after.ID, OwnerPlayerID: after.OwnerPlayerID, PreviousLevel: before.Level, Level: after.Level})
	}
	if respected {
		if packet, err := outrespected.Encode(after.Respect, after.OwnerPlayerID, InventoryPet(after)); err == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
	}
	service.projectStatus(ctx, active, after, unit, service.species(ctx, after), service.plant(ctx, after))
}

// sendPlayer sends one best-effort packet to an online owner.
func (service *Service) sendPlayer(ctx context.Context, playerID int64, packet codec.Packet) {
	if target, found := service.playerConnection(playerID); found {
		_ = target.Send(ctx, packet)
	}
}

// ProjectSpawn broadcasts one pet spawn and status.
func (service *Service) ProjectSpawn(ctx context.Context, active *roomlive.Room, pet petrecord.Pet) {
	unit, found := active.Unit(EntityKey(pet.ID))
	if !found {
		return
	}
	species, plant := service.speciesState(ctx, pet)
	_, riding := service.Rider(active.ID(), pet.ID)
	packet, err := outunits.Encode([]outunits.Unit{RoomUnit(pet, unit, species, plant, riding)})
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
	service.syncPlantStatus(ctx, active, pet, species, plant, false)
	service.projectStatus(ctx, active, pet, unit, species, plant)
}

// ProjectRemove broadcasts one room-local pet removal.
func (service *Service) ProjectRemove(ctx context.Context, active *roomlive.Room, roomIndex int64) {
	packet, err := outremoved.Encode(roomIndex)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// SendInventory sends ordered inventory fragments to one connection.
func (service *Service) SendInventory(ctx context.Context, target connection.Context, pets []petrecord.Pet) error {
	fragmentSize := service.config.InventoryFragmentSize
	now := service.Now()
	total := (len(pets) + fragmentSize - 1) / fragmentSize
	if total == 0 {
		total = 1
	}
	for fragment := 0; fragment < total; fragment++ {
		start, end := fragment*fragmentSize, (fragment+1)*fragmentSize
		if end > len(pets) {
			end = len(pets)
		}
		records := make([]petdata.Pet, end-start)
		for index := start; index < end; index++ {
			records[index-start] = InventoryPet(service.materialize(pets[index], now))
		}
		packet, err := outlist.Encode(int32(total), int32(fragment), records)
		if err != nil {
			return err
		}
		if err = target.Send(ctx, packet); err != nil {
			return err
		}
	}
	return nil
}

// SendInventoryAdd projects one incremental inventory addition.
func (service *Service) SendInventoryAdd(ctx context.Context, playerID int64, pet petrecord.Pet) {
	if service.inventories != nil {
		service.inventories.Invalidate(playerID)
	}
	connection, found := service.playerConnection(playerID)
	if !found {
		return
	}
	packet, err := outadd.Encode(InventoryPet(service.materialize(pet, service.Now())), false)
	if err == nil {
		_ = connection.Send(ctx, packet)
	}
}

// SendInventoryReceived projects one post-purchase pet notification.
func (service *Service) SendInventoryReceived(ctx context.Context, playerID int64, pet petrecord.Pet) {
	connection, found := service.playerConnection(playerID)
	if !found {
		return
	}
	packet, err := outreceived.Encode(false, InventoryPet(service.materialize(pet, service.Now())))
	if err == nil {
		_ = connection.Send(ctx, packet)
	}
}

// SendInventoryRemove projects one incremental inventory removal.
func (service *Service) SendInventoryRemove(ctx context.Context, playerID int64, petID int64) {
	if service.inventories != nil {
		service.inventories.Invalidate(playerID)
	}
	connection, found := service.playerConnection(playerID)
	if !found {
		return
	}
	packet, err := outremoveinventory.Encode(petID)
	if err == nil {
		_ = connection.Send(ctx, packet)
	}
}

// playerConnection resolves one authenticated player connection.
func (service *Service) playerConnection(playerID int64) (connection.Connection, bool) {
	if service.connections == nil || service.players == nil {
		return nil, false
	}
	player, found := service.players.Find(playerID)
	if !found {
		return nil, false
	}
	peer := player.Peer()
	return service.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
}

// projectStatus broadcasts breeding and plant flags.
func (service *Service) projectStatus(ctx context.Context, active *roomlive.Room, pet petrecord.Pet, unit roomlive.UnitSnapshot, species petrecord.Species, plant petrecord.PlantState) {
	packet, err := outstatus.Encode(unit.UnitID, pet.ID, species.Breedable && pet.CanBreed && !plant.Dead, plant.CanHarvest, plant.CanRevive, pet.PublicBreed)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// speciesState resolves cached species and derived plant lifecycle.
func (service *Service) speciesState(ctx context.Context, pet petrecord.Pet) (petrecord.Species, petrecord.PlantState) {
	if service.references == nil {
		return petrecord.Species{}, petrecord.PlantState{}
	}
	references, err := service.references.Current(ctx)
	if err != nil || pet.TypeID < 0 || pet.TypeID >= int32(len(references.Species)) {
		return petrecord.Species{}, petrecord.PlantState{}
	}
	species := references.Species[pet.TypeID]
	return species, pet.DerivePlantState(service.Now(), species)
}

// projectSpawnConnection sends one pet snapshot to a late entrant.
func (service *Service) projectSpawnConnection(ctx context.Context, target connection.Connection, active *roomlive.Room, pet petrecord.Pet) {
	unit, found := active.Unit(EntityKey(pet.ID))
	if !found || target == nil {
		return
	}
	species, plant := service.speciesState(ctx, pet)
	_, riding := service.Rider(active.ID(), pet.ID)
	packets := make([]codec.Packet, 0, 3)
	if packet, err := outunits.Encode([]outunits.Unit{RoomUnit(pet, unit, species, plant, riding)}); err == nil {
		packets = append(packets, packet)
	}
	if packet, err := outstatus.Encode(unit.UnitID, pet.ID, species.Breedable && pet.CanBreed && !plant.Dead, plant.CanHarvest, plant.CanRevive, pet.PublicBreed); err == nil {
		packets = append(packets, packet)
	}
	if packet, encoded := service.plantStatusPacket(active, pet.ID); encoded {
		packets = append(packets, packet)
	}
	for _, packet := range packets {
		_ = target.Send(ctx, packet)
	}
}

// SyncPlayer sends every active pet to one late entrant.
func (service *Service) SyncPlayer(ctx context.Context, roomID int64, playerID int64) error {
	active, found := service.rooms.Find(roomID)
	if !found {
		return nil
	}
	loaded, err := service.ensureRoom(ctx, active)
	if err != nil || loaded {
		return err
	}
	target, found := service.playerConnection(playerID)
	if !found {
		return nil
	}
	for _, current := range service.roomPets(roomID) {
		current.mutex.Lock()
		pet := current.record
		current.mutex.Unlock()
		service.projectSpawnConnection(ctx, target, active, pet)
	}
	return nil
}
