package routes

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// list returns one filtered keyset page.
func (dependencies Dependencies) list(ctx *fiber.Ctx) error {
	if err := dependencies.authorizeRead(ctx, petpolicy.ManageAny); err != nil {
		return err
	}
	filter, err := listFilter(ctx)
	if err != nil {
		return err
	}
	pets, err := dependencies.Pets.List(ctx.Context(), filter)
	if err != nil {
		return err
	}
	items := make([]PetResponse, len(pets))
	next := int64(0)
	for index, pet := range pets {
		items[index] = petResponse(pet)
		next = pet.ID
	}
	if len(pets) < filter.Limit {
		next = 0
	}
	return ctx.JSON(PetListResponse{Items: items, NextCursor: next})
}

// read returns one complete non-deleted pet.
func (dependencies Dependencies) read(ctx *fiber.Ctx) error {
	if err := dependencies.authorizeRead(ctx, petpolicy.ManageAny); err != nil {
		return err
	}
	id, err := pathID(ctx)
	if err != nil {
		return err
	}
	pet, found, err := dependencies.Pets.Find(ctx.Context(), id)
	if err != nil {
		return err
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, petrecord.ErrPetNotFound.Error())
	}
	return ctx.JSON(petResponse(pet))
}

// metrics returns process-wide low-cardinality pet telemetry.
func (dependencies Dependencies) metrics(ctx *fiber.Ctx) error {
	if err := dependencies.authorizeRead(ctx, petpolicy.ManageAny); err != nil {
		return err
	}
	return ctx.JSON(dependencies.Pets.Metrics())
}

// species returns the immutable species manifest.
func (dependencies Dependencies) species(ctx *fiber.Ctx) error {
	if err := dependencies.authorizeRead(ctx, petpolicy.ManageAny); err != nil {
		return err
	}
	values, err := dependencies.Pets.Species(ctx.Context())
	if err != nil {
		return err
	}
	items := make([]SpeciesResponse, len(values))
	for index, value := range values {
		items[index] = SpeciesResponse{TypeID: value.TypeID, Slug: value.Slug, DisplayKey: value.DisplayKey, BehaviorKind: value.BehaviorKind, MaximumLevel: value.MaximumLevel, Rideable: value.Rideable, Breedable: value.Breedable, Plant: value.Plant, Enabled: value.Enabled}
	}
	return ctx.JSON(items)
}

// breeds returns ordered appearance options.
func (dependencies Dependencies) breeds(ctx *fiber.Ctx) error {
	if err := dependencies.authorizeRead(ctx, petpolicy.ManageAny); err != nil {
		return err
	}
	typeID, err := optionalInt32(ctx.Query("typeId"))
	if err != nil {
		return err
	}
	values, err := dependencies.Pets.Breeds(ctx.Context(), typeID, ctx.QueryBool("sellable", false))
	if err != nil {
		return err
	}
	items := make([]BreedResponse, len(values))
	for index, value := range values {
		items[index] = BreedResponse{TypeID: value.TypeID, BreedID: value.BreedID, PaletteID: value.PaletteID, Color: value.Color, Sellable: value.Sellable, Rarity: value.Rarity, Enabled: value.Enabled}
	}
	return ctx.JSON(items)
}

// commands returns the command registry and per-species mappings.
func (dependencies Dependencies) commands(ctx *fiber.Ctx) error {
	if err := dependencies.authorizeRead(ctx, petpolicy.ManageAny); err != nil {
		return err
	}
	values, species, err := dependencies.Pets.Commands(ctx.Context())
	if err != nil {
		return err
	}
	items := make([]CommandResponse, len(values))
	for index, value := range values {
		items[index] = CommandResponse{ID: value.ID, NameKey: value.NameKey, RequiredLevel: value.RequiredLevel, Family: value.Family, EnergyCost: value.EnergyCost, HappinessCost: value.HappinessCost, ExperienceReward: value.ExperienceReward, DurationMilliseconds: value.Duration.Milliseconds(), CooldownMilliseconds: value.Cooldown.Milliseconds(), Enabled: value.Enabled}
	}
	return ctx.JSON(CommandsResponse{Items: items, Species: species})
}

// listFilter parses bounded pet list query fields.
func listFilter(ctx *fiber.Ctx) (petrecord.AdminFilter, error) {
	ownerID, err := optionalInt64(ctx.Query("ownerPlayerId"))
	if err != nil {
		return petrecord.AdminFilter{}, err
	}
	typeID, err := optionalInt32(ctx.Query("typeId"))
	if err != nil {
		return petrecord.AdminFilter{}, err
	}
	roomID, err := optionalInt64(ctx.Query("roomId"))
	if err != nil {
		return petrecord.AdminFilter{}, err
	}
	cursor, err := nonNegativeInt64(ctx.Query("cursor"))
	if err != nil {
		return petrecord.AdminFilter{}, err
	}
	limit, err := boundedLimit(ctx.Query("limit"))
	if err != nil {
		return petrecord.AdminFilter{}, err
	}
	state := strings.TrimSpace(ctx.Query("state"))
	if state != "" && state != petrecord.StateInventory && state != petrecord.StateRoom && state != petrecord.StateBreedingReserved && state != petrecord.StateHarvested && state != petrecord.StateComposted {
		return petrecord.AdminFilter{}, fiber.NewError(fiber.StatusBadRequest, "invalid pet state filter")
	}
	return petrecord.AdminFilter{OwnerPlayerID: ownerID, Name: strings.TrimSpace(ctx.Query("name")), TypeID: typeID, RoomID: roomID, State: state, IncludeDeleted: ctx.QueryBool("deleted", false), Cursor: cursor, Limit: limit}, nil
}

// pathID parses one positive pet identifier.
func pathID(ctx *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid pet identifier")
	}
	return id, nil
}

// optionalInt64 parses one optional positive identifier.
func optionalInt64(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid identifier filter")
	}
	return &parsed, nil
}

// optionalInt32 parses one optional non-negative protocol identifier.
func optionalInt32(value string) (*int32, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed < 0 || parsed > 35 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid pet type filter")
	}
	result := int32(parsed)
	return &result, nil
}

// nonNegativeInt64 parses one optional cursor.
func nonNegativeInt64(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
	}
	return parsed, nil
}

// boundedLimit parses one bounded result limit.
func boundedLimit(value string) (int, error) {
	if value == "" {
		return 50, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 || parsed > 200 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid result limit")
	}
	return parsed, nil
}
