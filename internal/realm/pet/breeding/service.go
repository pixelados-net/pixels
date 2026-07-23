// Package breeding owns nest sessions and offspring grants.
package breeding

import (
	"context"
	"fmt"
	"sort"

	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	petidentity "github.com/niflaot/pixels/internal/realm/pet/identity"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrequest "github.com/niflaot/pixels/networking/outbound/room/pet/breeding/request"
	outstate "github.com/niflaot/pixels/networking/outbound/room/pet/breeding/state"
)

// Service coordinates durable nest-owned breeding sessions.
type Service struct {
	// config stores breeding age and timeout policy.
	config petpolicy.Config
	// store persists sessions and offspring.
	store petrecord.Store
	// references resolves species and breeds.
	references petreference.Reader
	// runtime resolves active parent pets.
	runtime *petruntime.Service
	// rooms resolves active room generations.
	rooms *roomlive.Registry
	// players resolves owner connections.
	players *playerlive.Registry
	// connections sends owner-specific packets.
	connections *netconn.Registry
	// filter rejects censored offspring names.
	filter *chatfilter.Service
}

// New creates breeding behavior.
func New(config petpolicy.Config, store petrecord.Store, references petreference.Reader, runtime *petruntime.Service, rooms *roomlive.Registry, players *playerlive.Registry, connections *netconn.Registry, filter *chatfilter.Service) *Service {
	return &Service{config: config.Normalize(), store: store, references: references, runtime: runtime, rooms: rooms, players: players, connections: connections, filter: filter}
}

// Start creates or confirms one compatible nest session.
func (service *Service) Start(ctx context.Context, target netconn.Context, roomID int64, actorID int64, firstID int64, secondID int64) (err error) {
	result := petobservability.ResultSuccess
	defer func() { service.runtime.Metrics().RecordBreeding(petobservability.BreedingStart, result) }()
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}
	first, second, err := service.parents(ctx, roomID, actorID, firstID, secondID)
	if err != nil {
		result = petobservability.ResultRejected
		return service.sendFailure(ctx, target, 1)
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		result = petobservability.ResultRejected
		return service.sendFailure(ctx, target, 1)
	}
	nestID, found := breedingNest(active)
	if !found {
		result = petobservability.ResultRejected
		return service.sendFailure(ctx, target, 2)
	}
	value := petrecord.BreedingSession{NestItemID: nestID, RoomID: roomID, GenerationToken: fmt.Sprintf("%d:%d:%d:%d", roomID, nestID, firstID, secondID), ParentOneID: firstID, ParentTwoID: secondID, ExpiresAt: service.runtime.Now().Add(service.config.BreedingTimeout)}
	saved, stored, err := service.store.SaveBreedingSession(ctx, value, actorID)
	if err != nil || !stored {
		result = petobservability.Classify(err, err == nil)
		return firstError(err, service.sendFailure(ctx, target, 3))
	}
	packet, err := outstate.Encode(stateCode(saved.State), ownedPet(actorID, first, second), otherPet(actorID, first, second))
	if err == nil {
		err = target.Send(ctx, packet)
	}
	if err != nil || saved.State != "confirmed" {
		if err != nil {
			result = petobservability.ResultFailed
		}
		return err
	}
	err = service.sendRequest(ctx, saved, first, second)
	if err != nil {
		result = petobservability.ResultFailed
	}
	return err
}

// sendRequest sends the protocol-native confirmation payload to both owners.
func (service *Service) sendRequest(ctx context.Context, session petrecord.BreedingSession, first petrecord.Pet, second petrecord.Pet) error {
	resultTypeID, err := service.resultType(ctx, first.TypeID, second.TypeID)
	if err != nil {
		return err
	}
	categories, err := service.rarityCategories(ctx, resultTypeID)
	if err != nil {
		return err
	}
	packet, err := outrequest.Encode(session.NestItemID, parent(first), parent(second), categories, resultTypeID)
	if err != nil {
		return err
	}
	service.sendOwner(ctx, first.OwnerPlayerID, packet)
	if second.OwnerPlayerID != first.OwnerPlayerID {
		service.sendOwner(ctx, second.OwnerPlayerID, packet)
	}
	return nil
}

// parents validates room visibility, compatibility, and owner consent eligibility.
func (service *Service) parents(ctx context.Context, roomID int64, actorID int64, firstID int64, secondID int64) (petrecord.Pet, petrecord.Pet, error) {
	first, firstFound := service.runtime.Snapshot(roomID, firstID)
	second, secondFound := service.runtime.Snapshot(roomID, secondID)
	if !firstFound || !secondFound || first.ID == second.ID || first.OwnerPlayerID != actorID && second.OwnerPlayerID != actorID {
		return petrecord.Pet{}, petrecord.Pet{}, petrecord.ErrInvalidState
	}
	if first.OwnerPlayerID != actorID && !first.PublicBreed || second.OwnerPlayerID != actorID && !second.PublicBreed {
		return petrecord.Pet{}, petrecord.Pet{}, petrecord.ErrNoRights
	}
	if !first.CanBreed || !second.CanBreed {
		return petrecord.Pet{}, petrecord.Pet{}, petrecord.ErrInvalidState
	}
	references, err := service.references.Current(ctx)
	if err != nil || first.TypeID < 0 || first.TypeID >= int32(len(references.Species)) || second.TypeID < 0 || second.TypeID >= int32(len(references.Species)) || !references.SpeciesPresent[first.TypeID] || !references.SpeciesPresent[second.TypeID] || !references.Species[first.TypeID].Breedable || !references.Species[second.TypeID].Breedable {
		return petrecord.Pet{}, petrecord.Pet{}, firstError(err, petrecord.ErrInvalidState)
	}
	if _, compatible := references.BreedingResult(first.TypeID, second.TypeID); !compatible {
		return petrecord.Pet{}, petrecord.Pet{}, petrecord.ErrInvalidState
	}
	now := service.runtime.Now()
	if now.Sub(first.CreatedAt) < service.config.BreedingMinimumAge || now.Sub(second.CreatedAt) < service.config.BreedingMinimumAge {
		return petrecord.Pet{}, petrecord.Pet{}, petrecord.ErrInvalidState
	}
	species := references.Species[first.TypeID]
	if species.Plant && (first.DerivePlantState(now, species).Dead || second.DerivePlantState(now, species).Dead) {
		return petrecord.Pet{}, petrecord.Pet{}, petrecord.ErrInvalidState
	}
	return first, second, nil
}

// resultType resolves one enabled canonical parent compatibility rule.
func (service *Service) resultType(ctx context.Context, firstTypeID int32, secondTypeID int32) (int32, error) {
	references, err := service.references.Current(ctx)
	if err != nil {
		return 0, err
	}
	resultTypeID, found := references.BreedingResult(firstTypeID, secondTypeID)
	if !found {
		return 0, petrecord.ErrInvalidState
	}
	return resultTypeID, nil
}

// rarityCategories groups enabled breeds by rarity for Nitro's dialog.
func (service *Service) rarityCategories(ctx context.Context, typeID int32) ([]outrequest.RarityCategory, error) {
	references, err := service.references.Current(ctx)
	if err != nil {
		return nil, err
	}
	groups := make(map[int32][]int32)
	for _, breed := range references.Breeds {
		if breed.TypeID == typeID && breed.Enabled {
			groups[breed.Rarity] = append(groups[breed.Rarity], breed.BreedID)
		}
	}
	keys := make([]int, 0, len(groups))
	for rarity := range groups {
		keys = append(keys, int(rarity))
	}
	sort.Ints(keys)
	result := make([]outrequest.RarityCategory, 0, len(keys))
	for _, rarity := range keys {
		result = append(result, outrequest.RarityCategory{Chance: int32(max(1, 100/(rarity+1))), Breeds: groups[int32(rarity)]})
	}
	return result, nil
}

// validateName applies the shared approval and censorship policy.
func (service *Service) validateName(value string) (string, int32) {
	name, code := petidentity.ValidateName(value)
	if code == petidentity.NameApproved && service.filter != nil {
		if _, censored := service.filter.Censor(name); censored {
			code = petidentity.NameCensored
		}
	}
	return name, code
}
