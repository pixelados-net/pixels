package breeding

import (
	"context"
	"errors"
	"fmt"

	petcompleted "github.com/niflaot/pixels/internal/realm/pet/breeding/events/completed"
	petidentity "github.com/niflaot/pixels/internal/realm/pet/identity"
	petcreated "github.com/niflaot/pixels/internal/realm/pet/identity/events/created"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	netconn "github.com/niflaot/pixels/networking/connection"
	outreceived "github.com/niflaot/pixels/networking/outbound/inventory/pet/received"
	outconfirm "github.com/niflaot/pixels/networking/outbound/room/pet/breeding/confirmresult"
	outresult "github.com/niflaot/pixels/networking/outbound/room/pet/breeding/result"
	outsuccess "github.com/niflaot/pixels/networking/outbound/room/pet/breeding/success"
)

// Confirm validates one confirmed session and grants an idempotent offspring.
func (service *Service) Confirm(ctx context.Context, target netconn.Context, roomID int64, actorID int64, nestID int64, name string, firstID int64, secondID int64) (err error) {
	result := petobservability.ResultSuccess
	defer func() { service.runtime.Metrics().RecordBreeding(petobservability.BreedingConfirm, result) }()
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}
	name, code := service.validateName(name)
	if code != petidentity.NameApproved {
		result = petobservability.ResultRejected
		return service.sendConfirm(ctx, target, nestID, code)
	}
	session, found, err := service.store.FindBreedingSession(ctx, nestID)
	if err != nil || !found || session.RoomID != roomID || session.State != "confirmed" || session.ParentOneID != firstID || session.ParentTwoID != secondID || !service.runtime.Now().Before(session.ExpiresAt) {
		result = petobservability.Classify(err, err == nil)
		return firstError(err, service.sendConfirm(ctx, target, nestID, 5))
	}
	first, second, err := service.parents(ctx, roomID, actorID, firstID, secondID)
	if err != nil {
		result = petobservability.ResultRejected
		return service.sendConfirm(ctx, target, nestID, 5)
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		result = petobservability.ResultFailed
		return err
	}
	resultTypeID, found := references.BreedingResult(first.TypeID, second.TypeID)
	if !found {
		result = petobservability.ResultRejected
		return service.sendConfirm(ctx, target, nestID, 5)
	}
	seed := uint64(session.NestItemID) ^ uint64(session.ParentOneID)<<17 ^ uint64(session.ParentTwoID)<<33 ^ uint64(session.Version)
	breed, found := selectOffspringBreed(references, resultTypeID, seed)
	if !found {
		result = petobservability.ResultRejected
		return service.sendConfirm(ctx, target, nestID, 5)
	}
	ownerID := first.OwnerPlayerID
	if actorID == second.OwnerPlayerID {
		ownerID = actorID
	}
	params := petrecord.GrantParams{OwnerPlayerID: ownerID, Name: name, TypeID: resultTypeID, BreedID: breed.BreedID, PaletteID: breed.PaletteID, Color: breed.Color, OperationKey: fmt.Sprintf("breeding:%d:%d", nestID, session.Version)}
	if references.Species[resultTypeID].Plant {
		params.Parts = petidentity.MonsterPlantOffspringAppearance(first.Parts, second.Parts, seed)
	}
	offspring := petrecord.Pet{}
	firstAfter, secondAfter := petrecord.Pet{}, petrecord.Pet{}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		completed, completeErr := service.store.SetBreedingSessionState(txCtx, nestID, "confirmed", "completed", session.Version)
		if completeErr != nil || !completed {
			return firstError(completeErr, petrecord.ErrConflict)
		}
		var updated bool
		firstAfter, updated, completeErr = service.store.SetBreedingEligibility(txCtx, first.ID, false, first.Version)
		if completeErr != nil || !updated {
			return firstError(completeErr, petrecord.ErrConflict)
		}
		secondAfter, updated, completeErr = service.store.SetBreedingEligibility(txCtx, second.ID, false, second.Version)
		if completeErr != nil || !updated {
			return firstError(completeErr, petrecord.ErrConflict)
		}
		var granted bool
		offspring, granted, completeErr = service.store.Grant(txCtx, params)
		if completeErr == nil && !granted && offspring.ID == 0 {
			return petrecord.ErrConflict
		}
		return completeErr
	})
	if err != nil {
		result = petobservability.Classify(err, errors.Is(err, petrecord.ErrConflict))
		return err
	}
	service.runtime.ReplacePlaced(firstAfter)
	service.runtime.ReplacePlaced(secondAfter)
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		service.runtime.ProjectFigure(ctx, active, firstAfter)
		service.runtime.ProjectFigure(ctx, active, secondAfter)
	}
	service.runtime.SendInventoryAdd(ctx, ownerID, offspring)
	service.runtime.Publish(ctx, petcreated.Name, petcreated.Payload{PetID: offspring.ID, OwnerPlayerID: offspring.OwnerPlayerID, TypeID: offspring.TypeID})
	service.runtime.Publish(ctx, petcompleted.Name, petcompleted.Payload{NestItemID: nestID, ParentOneID: first.ID, ParentTwoID: second.ID, OffspringID: offspring.ID})
	if packet, encodeErr := outreceived.Encode(false, petruntime.InventoryPet(offspring)); encodeErr == nil {
		service.sendOwner(ctx, ownerID, packet)
	}
	if err = service.sendConfirm(ctx, target, nestID, 0); err != nil {
		result = petobservability.ResultFailed
		return err
	}
	if packet, encodeErr := outsuccess.Encode(offspring.ID, offspring.Rarity); encodeErr == nil {
		_ = target.Send(ctx, packet)
	}
	firstResult := breedingResult(offspring, first.OwnerName)
	secondResult := breedingResult(offspring, second.OwnerName)
	if packet, encodeErr := outresult.Encode(firstResult, secondResult); encodeErr == nil {
		_ = target.Send(ctx, packet)
	}
	return nil
}

// Cancel releases one requested or confirmed breeding session.
func (service *Service) Cancel(ctx context.Context, nestID int64, roomID int64) (err error) {
	defer func() {
		expected := errors.Is(err, petrecord.ErrInvalidState) || errors.Is(err, petrecord.ErrConflict)
		service.runtime.Metrics().RecordBreeding(petobservability.BreedingCancel, petobservability.Classify(err, expected))
	}()
	session, found, err := service.store.FindBreedingSession(ctx, nestID)
	if err != nil || !found || session.RoomID != roomID {
		return firstError(err, petrecord.ErrInvalidState)
	}
	if session.State != "requested" && session.State != "confirmed" {
		return nil
	}
	updated, err := service.store.SetBreedingSessionState(ctx, nestID, session.State, "cancelled", session.Version)
	if err != nil || !updated {
		return firstError(err, petrecord.ErrConflict)
	}
	return nil
}

// CloseRoom releases every active session owned by a closing room.
func (service *Service) CloseRoom(roomID int64) {
	_ = service.store.CancelBreedingRoom(context.Background(), roomID)
}

// sendConfirm sends one native breeding result code.
func (service *Service) sendConfirm(ctx context.Context, target netconn.Context, nestID int64, code int32) error {
	packet, err := outconfirm.Encode(nestID, code)
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// breedingResult maps one offspring to Nitro's two-result schema.
func breedingResult(pet petrecord.Pet, ownerName string) outresult.Result {
	return outresult.Result{StuffID: pet.ID, ClassID: pet.TypeID, ProductCode: fmt.Sprintf("pet%d", pet.TypeID), UserID: pet.OwnerPlayerID, UserName: ownerName, RarityLevel: pet.Rarity}
}
