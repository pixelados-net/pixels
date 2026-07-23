package runtime

import (
	"context"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/connection"
	outunitstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outunits "github.com/niflaot/pixels/networking/outbound/room/entities/units"
	outfigure "github.com/niflaot/pixels/networking/outbound/room/pet/figure"
	outinfo "github.com/niflaot/pixels/networking/outbound/room/pet/info"
	outtraining "github.com/niflaot/pixels/networking/outbound/room/pet/training"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// Rider returns the current rider for one active pet.
func (service *Service) Rider(roomID int64, petID int64) (int64, bool) {
	pet, found := service.Active(roomID, petID)
	if !found {
		return 0, false
	}
	pet.mutex.Lock()
	rider := pet.riderPlayerID
	pet.mutex.Unlock()
	return rider, rider != 0
}

// FindByUnit returns the durable pet using one room-local unit identifier.
func (service *Service) FindByUnit(roomID int64, unitID int64) (petrecord.Pet, bool) {
	active, found := service.rooms.Find(roomID)
	if !found {
		return petrecord.Pet{}, false
	}
	unit, found := active.UnitByID(unitID)
	if !found || unit.Kind != worldunit.KindPet || unit.EntityKey < entityBase {
		return petrecord.Pet{}, false
	}
	return service.Snapshot(roomID, unit.EntityKey-entityBase)
}

// Information builds Nitro's complete visible pet information.
func (service *Service) Information(ctx context.Context, pet petrecord.Pet) (outinfo.Info, error) {
	if service.references == nil {
		return outinfo.Info{}, petrecord.ErrInvalidState
	}
	now := service.Now()
	pet = service.materialize(pet, now)
	references, err := service.references.Current(ctx)
	if err != nil {
		return outinfo.Info{}, err
	}
	species := petrecord.Species{}
	if pet.TypeID >= 0 && pet.TypeID < int32(len(references.Species)) && references.SpeciesPresent[pet.TypeID] {
		species = references.Species[pet.TypeID]
	}
	plant := pet.DerivePlantState(now, species)
	level, maximumLevel, experience, experienceGoal := pet.Level, species.MaximumLevel, pet.Experience, petrecord.NextThreshold(pet.Level)
	if species.Plant {
		level, maximumLevel, experience, experienceGoal = plant.GrowthStage, 7, 0, 0
	}
	var thresholds []int32
	if pet.TypeID >= 0 && pet.TypeID < int32(len(references.SpeciesCommands)) {
		thresholds = references.SpeciesCommands[pet.TypeID]
	}
	age := now.Sub(pet.CreatedAt)
	if age < 0 {
		age = 0
	}
	maximumLife := int32(0)
	if pet.GrowAt != nil && pet.DieAt != nil {
		maximumLife = boundedSeconds(pet.DieAt.Sub(*pet.GrowAt))
	}
	return outinfo.Info{
		ID: pet.ID, Name: pet.Name, Level: level, MaximumLevel: maximumLevel, Experience: experience,
		LevelExperienceGoal: experienceGoal, Energy: pet.Energy, MaximumEnergy: petrecord.MaximumEnergy(pet.Level),
		Happiness: pet.Happiness, MaximumHappiness: 100, Respect: pet.Respect, OwnerID: pet.OwnerPlayerID,
		AgeDays: int32(age / (24 * time.Hour)), OwnerName: pet.OwnerName, Rarity: pet.Rarity,
		Saddle: pet.HasSaddle, SkillThresholds: thresholds, PubliclyRideable: boolInt(pet.PublicRide),
		Breedable: species.Breedable && pet.CanBreed && !plant.Dead, FullyGrown: plant.FullyGrown, Dead: plant.Dead,
		MaximumTimeToLive: maximumLife, RemainingTimeToLive: plant.RemainingLifeSeconds,
		RemainingGrowTime: plant.RemainingGrowSeconds, PubliclyBreedable: pet.PublicBreed,
	}, nil
}

// ProjectFigure broadcasts current saddle and riding appearance state.
func (service *Service) ProjectFigure(ctx context.Context, active *roomlive.Room, pet petrecord.Pet) {
	unit, found := active.Unit(EntityKey(pet.ID))
	if !found {
		return
	}
	_, riding := service.Rider(active.ID(), pet.ID)
	packet, err := outfigure.Encode(unit.UnitID, pet.ID, InventoryPet(pet).Figure, pet.HasSaddle, riding)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
	species, plant := service.speciesState(ctx, pet)
	service.syncPlantStatus(ctx, active, pet, species, plant, false)
	service.projectStatus(ctx, active, pet, unit, species, plant)
}

// SendInformation sends one complete pet information packet.
func (service *Service) SendInformation(ctx context.Context, target connection.Context, pet petrecord.Pet) error {
	value, err := service.Information(ctx, pet)
	if err != nil {
		return err
	}
	packet, err := outinfo.Encode(value)
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// SendTraining sends all species commands and currently unlocked commands.
func (service *Service) SendTraining(ctx context.Context, target connection.Context, pet petrecord.Pet) error {
	if service.references == nil {
		return petrecord.ErrInvalidState
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		return err
	}
	all, enabled := petreference.TrainingCommands(references, pet)
	packet, err := outtraining.Encode(pet.ID, all, enabled)
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// species resolves one cached species value.
func (service *Service) species(ctx context.Context, pet petrecord.Pet) petrecord.Species {
	species, _ := service.speciesState(ctx, pet)
	return species
}

// plant resolves one cached derived plant state.
func (service *Service) plant(ctx context.Context, pet petrecord.Pet) petrecord.PlantState {
	_, plant := service.speciesState(ctx, pet)
	return plant
}

// RoomUnit maps one active pet into Nitro's type-two UNIT variant.
func RoomUnit(pet petrecord.Pet, unit roomlive.UnitSnapshot, species petrecord.Species, nowState petrecord.PlantState, riding bool) outunits.Unit {
	figure, visibleLevel, posture := InventoryPet(pet).Figure, pet.Level, pet.Posture
	if species.Plant {
		visibleLevel = nowState.GrowthStage
		posture = plantStatusKey(int8(nowState.GrowthStage))
		if nowState.Dead {
			posture = plantStatusKey(-1)
		}
	}
	return outunits.Unit{Type: outunits.PetType, UserID: pet.ID, Name: pet.Name, Figure: petdata.FigureString(figure), RoomIndex: unit.UnitID,
		X: int32(unit.Position.Point.X), Y: int32(unit.Position.Point.Y), Z: unit.Position.Z.String(), Direction: int32(unit.BodyRotation),
		PetSpecies: pet.TypeID, OwnerID: pet.OwnerPlayerID, OwnerName: pet.OwnerName, PetRarity: pet.Rarity, HasSaddle: pet.HasSaddle, IsRiding: riding,
		CanBreed: species.Breedable && pet.CanBreed && !nowState.Dead, CanHarvest: nowState.CanHarvest, CanRevive: nowState.CanRevive,
		HasBreedingPermission: pet.PublicBreed, PetLevel: visibleLevel, Posture: posture}
}

// syncPlantStatus stores and projects Comet-compatible growth or death status.
func (service *Service) syncPlantStatus(ctx context.Context, active *roomlive.Room, pet petrecord.Pet, species petrecord.Species, plant petrecord.PlantState, force bool) bool {
	if !species.Plant {
		return false
	}
	stage := int8(plant.GrowthStage)
	if plant.Dead {
		stage = -1
	}
	controller, found := service.Active(active.ID(), pet.ID)
	if !found {
		return false
	}
	controller.mutex.Lock()
	unchanged := controller.plantStage == stage
	controller.plantStage = stage
	controller.mutex.Unlock()
	if unchanged && !force {
		return false
	}
	for current := int8(-1); current <= 7; current++ {
		active.ClearUnitStatus(EntityKey(pet.ID), plantStatusKey(current))
	}
	active.SetUnitStatus(EntityKey(pet.ID), plantStatusKey(stage), "")
	unit, found := active.Unit(EntityKey(pet.ID))
	if found {
		_ = broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
	}
	return true
}

// plantStatusKey maps lifecycle stages to Nitro renderer actions.
func plantStatusKey(stage int8) string {
	switch stage {
	case -1:
		return "rip"
	case 1:
		return "grw1"
	case 2:
		return "grw2"
	case 3:
		return "grw3"
	case 4:
		return "grw4"
	case 5:
		return "grw5"
	case 6:
		return "grw6"
	case 7:
		return "std"
	default:
		return "grw"
	}
}

// plantStatusPacket encodes one persistent lifecycle status for a late entrant.
func (service *Service) plantStatusPacket(active *roomlive.Room, petID int64) (codec.Packet, bool) {
	records := roomprojection.Statuses(active, EntityKey(petID))
	if len(records) == 0 {
		return codec.Packet{}, false
	}
	packet, err := outunitstatus.Encode(records)
	return packet, err == nil
}

// boundedSeconds converts one positive duration to protocol seconds.
func boundedSeconds(value time.Duration) int32 {
	seconds := int64(value / time.Second)
	if seconds <= 0 {
		return 0
	}
	if seconds > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(seconds)
}

// boolInt converts a boolean to Nitro's numeric flag.
func boolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
