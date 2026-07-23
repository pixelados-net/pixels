// Package catalog owns pet palette, name, package, and catalog grant behavior.
package catalog

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	petidentity "github.com/niflaot/pixels/internal/realm/pet/identity"
	petcreated "github.com/niflaot/pixels/internal/realm/pet/identity/events/created"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbreeds "github.com/niflaot/pixels/networking/outbound/catalog/pet/breeds"
	outname "github.com/niflaot/pixels/networking/outbound/catalog/pet/nameapproval"
)

// Service coordinates pet catalog-facing operations.
type Service struct {
	// config stores package and seed lifecycle policy.
	config petpolicy.Config
	// store persists grants and transaction scope.
	store petrecord.Store
	// references resolves immutable palette and product data.
	references petreference.Reader
	// furniture validates and consumes package items.
	furniture furnitureservice.TradingManager
	// furnitureRooms picks up placed packages before consumption.
	furnitureRooms furnitureservice.Manager
	// runtime projects online inventory changes.
	runtime *petruntime.Service
	// filter rejects globally censored names.
	filter *chatfilter.Service
	// rooms resolves active package furniture.
	rooms *roomlive.Registry
	// connections broadcasts consumed package removal.
	connections *netconn.Registry
	// packageMutex protects bounded package naming handshakes.
	packageMutex sync.Mutex
	// packageRequests stores pending naming prompts by furniture item.
	packageRequests map[int64]packageRequest
}

// New creates pet catalog behavior.
func New(config petpolicy.Config, store petrecord.Store, references petreference.Reader, furniture furnitureservice.TradingManager, furnitureRooms furnitureservice.Manager, runtime *petruntime.Service, filter *chatfilter.Service, rooms *roomlive.Registry, connections *netconn.Registry) *Service {
	return &Service{config: config.Normalize(), store: store, references: references, furniture: furniture, furnitureRooms: furnitureRooms, runtime: runtime, filter: filter, rooms: rooms, connections: connections, packageRequests: make(map[int64]packageRequest)}
}

// packageRequest stores one bounded furniture naming prompt.
type packageRequest struct {
	// ownerID identifies the player who opened the prompt.
	ownerID int64
	// roomID identifies the room generation containing the package.
	roomID int64
	// expiresAt stores the absolute acceptance deadline.
	expiresAt time.Time
}

// ValidateName returns Nitro's shared pet name result.
func (service *Service) ValidateName(value string) (string, int32) {
	name, code := petidentity.ValidateName(value)
	if code == petidentity.NameApproved && service.filter != nil {
		if _, censored := service.filter.Censor(name); censored {
			code = petidentity.NameCensored
		}
	}
	return name, code
}

// SendNameApproval validates and sends one protocol result.
func (service *Service) SendNameApproval(ctx context.Context, target netconn.Context, value string) error {
	name, code := service.ValidateName(value)
	packet, err := outname.Encode(code, name)
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// SendBreeds sends enabled palettes for one product code.
func (service *Service) SendBreeds(ctx context.Context, target netconn.Context, productCode string) error {
	typeID, err := typeFromProductCode(productCode)
	if err != nil {
		return err
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		return err
	}
	values := make([]outbreeds.Palette, 0)
	for _, breed := range references.Breeds {
		if breed.TypeID == typeID && breed.Enabled {
			values = append(values, outbreeds.Palette{TypeID: breed.TypeID, BreedID: breed.BreedID, PaletteID: breed.PaletteID, Sellable: breed.Sellable, Rare: breed.Rarity > 0})
		}
	}
	packet, err := outbreeds.Encode(productCode, values)
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// Grant creates one validated pet from a trusted reward definition.
func (service *Service) Grant(ctx context.Context, ownerID int64, typeID int32, breedID int32, paletteID int32, color string, name string, operationKey string) (petrecord.Pet, bool, error) {
	name, code := service.ValidateName(name)
	if code != petidentity.NameApproved {
		return petrecord.Pet{}, false, petrecord.ErrInvalidName
	}
	color, err := petidentity.NormalizeColor(color)
	if err != nil {
		return petrecord.Pet{}, false, err
	}
	references, err := service.references.Current(ctx)
	if err != nil || typeID < 0 || typeID >= int32(len(references.Species)) || !references.SpeciesPresent[typeID] || !references.Species[typeID].Enabled {
		return petrecord.Pet{}, false, firstError(err, petrecord.ErrInvalidAppearance)
	}
	breed, found := references.Breeds[petreference.BreedKey{TypeID: typeID, BreedID: breedID, PaletteID: paletteID}]
	if !found || !breed.Enabled || !breed.Sellable {
		return petrecord.Pet{}, false, petrecord.ErrInvalidAppearance
	}
	return service.store.Grant(ctx, petrecord.GrantParams{OwnerPlayerID: ownerID, Name: name, TypeID: typeID, BreedID: breedID, PaletteID: paletteID, Color: color, OperationKey: operationKey})
}

// GrantCatalog creates one typed pet reward inside the catalog transaction.
func (service *Service) GrantCatalog(ctx context.Context, params catalogservice.PetGrantParams) (catalogservice.PetReward, error) {
	name, paletteID, color, err := parsePurchaseData(params.ExtraData)
	if err != nil {
		return catalogservice.PetReward{}, err
	}
	productTypeID, err := typeFromProductCode(params.ProductCode)
	if err != nil || productTypeID != params.TypeID {
		return catalogservice.PetReward{}, petrecord.ErrInvalidAppearance
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		return catalogservice.PetReward{}, err
	}
	breed, found := sellablePalette(references, params.TypeID, paletteID)
	if !found {
		return catalogservice.PetReward{}, petrecord.ErrInvalidAppearance
	}
	pet, _, err := service.Grant(ctx, params.OwnerPlayerID, params.TypeID, breed.BreedID, paletteID, color, name, params.OperationKey)
	if err != nil {
		return catalogservice.PetReward{}, err
	}
	return catalogservice.PetReward{ID: pet.ID, OwnerPlayerID: pet.OwnerPlayerID}, nil
}

// ProjectCatalog sends the committed pet reward to an online inventory owner.
func (service *Service) ProjectCatalog(ctx context.Context, reward catalogservice.PetReward) {
	pet, found, err := service.store.Find(ctx, reward.ID)
	if err != nil || !found || pet.OwnerPlayerID != reward.OwnerPlayerID {
		return
	}
	service.runtime.SendInventoryAdd(ctx, reward.OwnerPlayerID, pet)
	service.runtime.SendInventoryReceived(ctx, reward.OwnerPlayerID, pet)
	service.runtime.Publish(ctx, petcreated.Name, petcreated.Payload{PetID: pet.ID, OwnerPlayerID: pet.OwnerPlayerID, TypeID: pet.TypeID})
}

// parsePurchaseData validates Nitro's three-line pet purchase payload.
func parsePurchaseData(value string) (string, int32, string, error) {
	parts := strings.Split(value, "\n")
	if len(parts) != 3 {
		return "", 0, "", petrecord.ErrInvalidAppearance
	}
	palette, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil || palette < 0 {
		return "", 0, "", petrecord.ErrInvalidAppearance
	}
	color, err := petidentity.NormalizeColor(parts[2])
	if err != nil {
		return "", 0, "", err
	}
	return parts[0], int32(palette), color, nil
}

// sellablePalette selects a stable sellable breed for one palette.
func sellablePalette(references *petreference.Snapshot, typeID int32, paletteID int32) (petrecord.Breed, bool) {
	selected := petrecord.Breed{}
	found := false
	for _, breed := range references.Breeds {
		if breed.TypeID != typeID || breed.PaletteID != paletteID || !breed.Enabled || !breed.Sellable {
			continue
		}
		if !found || breed.BreedID < selected.BreedID {
			selected, found = breed, true
		}
	}
	return selected, found
}

// typeFromProductCode parses the trailing protocol species identifier.
func typeFromProductCode(value string) (int32, error) {
	value = strings.TrimSpace(value)
	index := len(value)
	for index > 0 && value[index-1] >= '0' && value[index-1] <= '9' {
		index--
	}
	if index == len(value) {
		return 0, petrecord.ErrInvalidAppearance
	}
	parsed, err := strconv.ParseInt(value[index:], 10, 32)
	if err != nil || parsed < 0 || parsed > 35 {
		return 0, petrecord.ErrInvalidAppearance
	}
	return int32(parsed), nil
}

// firstError chooses infrastructure failures over domain fallbacks.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

// PackageOperationKey returns the idempotency key for one package item.
func PackageOperationKey(itemID int64) string { return fmt.Sprintf("package:%d", itemID) }
