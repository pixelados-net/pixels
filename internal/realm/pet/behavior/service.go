package behavior

import (
	"context"
	"slices"

	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
)

// Service parses delivered room chat into validated pet commands.
type Service struct {
	// registry resolves immutable action implementations.
	registry *Registry
	// references resolves learned commands and costs.
	references petreference.Reader
	// runtime executes room-owned actions.
	runtime *petruntime.Service
}

// New creates pet speech behavior.
func New(registry *Registry, references petreference.Reader, runtime *petruntime.Service) *Service {
	return &Service{registry: registry, references: references, runtime: runtime}
}

// HandleSpeech executes one addressed command after chat delivery.
func (service *Service) HandleSpeech(ctx context.Context, roomID int64, playerID int64, message string) error {
	pet, keyword, found := service.runtime.FindNamed(roomID, message)
	if !found || pet.OwnerPlayerID != playerID {
		return nil
	}
	definition, found := service.registry.Resolve(keyword)
	if !found {
		return nil
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		return err
	}
	if pet.TypeID < 0 || pet.TypeID >= int32(len(references.SpeciesCommands)) || !slices.Contains(references.SpeciesCommands[pet.TypeID], definition.ID) {
		service.runtime.Metrics().RecordAction(definition.ID, petobservability.ResultRejected)
		return petrecord.ErrInvalidState
	}
	if definition.ID < 0 || definition.ID >= int32(len(references.Commands)) || !references.CommandPresent[definition.ID] {
		service.runtime.Metrics().RecordAction(definition.ID, petobservability.ResultRejected)
		return petrecord.ErrInvalidState
	}
	if err = service.runtime.ExecuteAction(ctx, roomID, pet.ID, playerID, definition.Action, references.Commands[definition.ID]); err != nil {
		service.runtime.Metrics().RecordAction(definition.ID, petobservability.ResultRejected)
		return err
	}
	service.runtime.Metrics().RecordAction(definition.ID, petobservability.ResultSuccess)
	if definition.ID == 10 {
		return service.runtime.VocalizeByPet(ctx, roomID, pet.ID)
	}
	return nil
}
