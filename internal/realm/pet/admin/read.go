package admin

import (
	"context"
	"sort"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// List returns one bounded protected pet page.
func (service *Service) List(ctx context.Context, filter petrecord.AdminFilter) ([]petrecord.Pet, error) {
	return service.store.ListAdmin(ctx, filter)
}

// Find returns one non-deleted pet aggregate.
func (service *Service) Find(ctx context.Context, petID int64) (petrecord.Pet, bool, error) {
	return service.store.Find(ctx, petID)
}

// Species returns the current immutable species manifest.
func (service *Service) Species(ctx context.Context) ([]petrecord.Species, error) {
	snapshot, err := service.references.Current(ctx)
	if err != nil {
		return nil, err
	}
	values := make([]petrecord.Species, 0, len(snapshot.Species))
	for typeID, present := range snapshot.SpeciesPresent {
		if present {
			values = append(values, snapshot.Species[typeID])
		}
	}
	return values, nil
}

// Breeds returns current breeds optionally restricted by species.
func (service *Service) Breeds(ctx context.Context, typeID *int32, sellableOnly bool) ([]petrecord.Breed, error) {
	snapshot, err := service.references.Current(ctx)
	if err != nil {
		return nil, err
	}
	values := make([]petrecord.Breed, 0, len(snapshot.Breeds))
	for _, breed := range snapshot.Breeds {
		if typeID != nil && breed.TypeID != *typeID || sellableOnly && !breed.Sellable {
			continue
		}
		values = append(values, breed)
	}
	sort.Slice(values, func(left int, right int) bool {
		if values[left].TypeID != values[right].TypeID {
			return values[left].TypeID < values[right].TypeID
		}
		if values[left].BreedID != values[right].BreedID {
			return values[left].BreedID < values[right].BreedID
		}
		return values[left].PaletteID < values[right].PaletteID
	})
	return values, nil
}

// Commands returns the complete command registry and species mappings.
func (service *Service) Commands(ctx context.Context) ([]petrecord.Command, map[int32][]int32, error) {
	snapshot, err := service.references.Current(ctx)
	if err != nil {
		return nil, nil, err
	}
	values := make([]petrecord.Command, 0, len(snapshot.Commands))
	links := make(map[int32][]int32, len(snapshot.SpeciesCommands))
	for id, present := range snapshot.CommandPresent {
		if present {
			values = append(values, snapshot.Commands[id])
		}
	}
	for typeID, commandIDs := range snapshot.SpeciesCommands {
		if commandIDs != nil {
			links[int32(typeID)] = append([]int32(nil), commandIDs...)
		}
	}
	return values, links, nil
}

// RefreshReference validates and publishes one new reference generation.
func (service *Service) RefreshReference(ctx context.Context, audit Audit) error {
	if err := audit.Validate(); err != nil {
		return err
	}
	if err := service.references.Refresh(ctx); err != nil {
		return err
	}
	return service.store.AppendGlobalAudit(ctx, audit.ActorPlayerID, "reference_refreshed", audit.Reason)
}
