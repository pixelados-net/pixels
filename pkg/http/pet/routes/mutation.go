package routes

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	petadmin "github.com/niflaot/pixels/internal/realm/pet/admin"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

const idempotencyHeader = "Idempotency-Key"

// create grants one pet idempotently.
func (dependencies Dependencies) create(ctx *fiber.Ctx) error {
	request := CreateRequest{}
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, petpolicy.ManageAny); err != nil {
		return err
	}
	pet, created, err := dependencies.Pets.Create(ctx.Context(), petadmin.CreateParams{OwnerPlayerID: request.OwnerPlayerID, Name: request.Name, TypeID: request.TypeID, BreedID: request.BreedID, PaletteID: request.PaletteID, Color: request.Color, OperationKey: strings.TrimSpace(ctx.Get(idempotencyHeader)), Audit: audit(request.AuditRequest)})
	if err != nil {
		return petError(err)
	}
	response := petResponse(pet)
	status := fiber.StatusOK
	if created {
		status = fiber.StatusCreated
	}
	return ctx.Status(status).JSON(MutationResponse{Pet: &response, Created: created})
}

// update applies one optimistic pet patch.
func (dependencies Dependencies) update(ctx *fiber.Ctx) error {
	id, err := pathID(ctx)
	if err != nil {
		return err
	}
	request := UpdateRequest{}
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, petpolicy.ManageAny); err != nil {
		return err
	}
	pet, err := dependencies.Pets.Update(ctx.Context(), id, petrecord.AdminPatch{Name: request.Name, BreedID: request.BreedID, PaletteID: request.PaletteID, Color: request.Color, PublicRide: request.PublicRide, PublicBreed: request.PublicBreed, Version: request.Version}, audit(request.AuditRequest))
	return mutationPet(ctx, pet, err)
}

// delete soft-deletes one pet.
func (dependencies Dependencies) delete(ctx *fiber.Ctx) error {
	id, err := pathID(ctx)
	if err != nil {
		return err
	}
	request := DeleteRequest{}
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, petpolicy.ManageAny); err != nil {
		return err
	}
	if err = dependencies.Pets.Delete(ctx.Context(), id, request.Version, audit(request.AuditRequest)); err != nil {
		return petError(err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// transferOwner moves one inventory pet to another owner.
func (dependencies Dependencies) transferOwner(ctx *fiber.Ctx) error {
	id, err := pathID(ctx)
	if err != nil {
		return err
	}
	request := OwnerRequest{}
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, petpolicy.ManageAny); err != nil {
		return err
	}
	pet, err := dependencies.Pets.TransferOwner(ctx.Context(), id, request.OwnerPlayerID, request.Version, audit(request.AuditRequest))
	return mutationPet(ctx, pet, err)
}

// setLocation places or picks one pet through the room world.
func (dependencies Dependencies) setLocation(ctx *fiber.Ctx) error {
	id, err := pathID(ctx)
	if err != nil {
		return err
	}
	request := LocationRequest{}
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, petpolicy.ManageAny); err != nil {
		return err
	}
	var point *grid.Point
	if request.RoomID != nil {
		if request.X == nil || request.Y == nil {
			return fiber.NewError(fiber.StatusBadRequest, "room placement requires x and y")
		}
		value, valid := grid.NewPoint(*request.X, *request.Y)
		if !valid {
			return fiber.NewError(fiber.StatusBadRequest, "invalid room placement tile")
		}
		point = &value
	}
	pet, err := dependencies.Pets.SetLocation(ctx.Context(), id, request.RoomID, point, audit(request.AuditRequest))
	return mutationPet(ctx, pet, err)
}

// updateStats applies bounded stat deltas.
func (dependencies Dependencies) updateStats(ctx *fiber.Ctx) error {
	id, err := pathID(ctx)
	if err != nil {
		return err
	}
	request := StatsRequest{}
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, petpolicy.LifecycleManage); err != nil {
		return err
	}
	pet, err := dependencies.Pets.UpdateStats(ctx.Context(), id, request.EnergyDelta, request.HappinessDelta, request.ExperienceDelta, request.Version, audit(request.AuditRequest))
	return mutationPet(ctx, pet, err)
}

// refreshReference validates and publishes one reference generation.
func (dependencies Dependencies) refreshReference(ctx *fiber.Ctx) error {
	request := AuditRequest{}
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, petpolicy.ManageAny); err != nil {
		return err
	}
	if err := dependencies.Pets.RefreshReference(ctx.Context(), audit(request)); err != nil {
		return petError(err)
	}
	return ctx.JSON(ReferenceRefreshResponse{Refreshed: true})
}

// parseBody parses one required JSON request body.
func parseBody(ctx *fiber.Ctx, target any) error {
	if err := ctx.BodyParser(target); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid pet administration request body")
	}
	return nil
}

// audit maps one transport request to domain attribution.
func audit(request AuditRequest) petadmin.Audit {
	return petadmin.Audit{ActorPlayerID: request.ActorPlayerID, Reason: request.Reason}
}

// mutationPet writes one common mutation response.
func mutationPet(ctx *fiber.Ctx, pet petrecord.Pet, err error) error {
	if err != nil {
		return petError(err)
	}
	response := petResponse(pet)
	return ctx.JSON(MutationResponse{Pet: &response})
}

// petError maps expected domain failures to meaningful HTTP responses.
func petError(err error) error {
	switch {
	case errors.Is(err, petrecord.ErrPetNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, petrecord.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, petrecord.ErrNoRights):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, petrecord.ErrInvalidState), errors.Is(err, petrecord.ErrInvalidName), errors.Is(err, petrecord.ErrInvalidAppearance), errors.Is(err, petrecord.ErrInvalidProduct), errors.Is(err, petrecord.ErrTileNotFree), errors.Is(err, petrecord.ErrRoomLimit), errors.Is(err, petrecord.ErrInventoryLimit), errors.Is(err, petrecord.ErrPetsDisabled):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	default:
		return err
	}
}
