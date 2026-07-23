// Package care owns pet respect, training, and stat changes.
package care

import (
	"context"
	"errors"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	petrespected "github.com/niflaot/pixels/internal/realm/pet/care/events/respected"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// Service coordinates pet care mutations.
type Service struct {
	// config stores respect policy.
	config petpolicy.Config
	// store persists stat mutations.
	store petrecord.Store
	// permissions resolves respect quota bypasses.
	permissions permissionservice.Checker
	// rooms resolves active room generations.
	rooms *roomlive.Registry
	// runtime owns room projections.
	runtime *petruntime.Service
}

// New creates pet care behavior.
func New(config petpolicy.Config, store petrecord.Store, permissions permissionservice.Checker, rooms *roomlive.Registry, runtime *petruntime.Service) *Service {
	return &Service{config: config.Normalize(), store: store, permissions: permissions, rooms: rooms, runtime: runtime}
}

// RespectResult stores one respect attempt outcome.
type RespectResult struct {
	// Pet stores the resulting pet state.
	Pet petrecord.Pet
	// Applied reports a newly consumed daily respect.
	Applied bool
	// TooYoung reports the minimum age gate failed.
	TooYoung bool
	// AgeDays stores the current whole-day age.
	AgeDays int32
	// RequiredAgeDays stores the minimum whole-day age.
	RequiredAgeDays int32
}

// Respect validates and applies one daily respect.
func (service *Service) Respect(ctx context.Context, roomID int64, petID int64, actorID int64) (respect RespectResult, err error) {
	result := petobservability.ResultSuccess
	defer func() {
		expected := errors.Is(err, petrecord.ErrPetNotFound) || errors.Is(err, petrecord.ErrNoRights) || errors.Is(err, petrecord.ErrRespectQuota)
		if err != nil {
			result = petobservability.Classify(err, expected)
		}
		service.runtime.Metrics().RecordOperation(petobservability.OperationRespect, result)
	}()
	pet, found := service.runtime.Snapshot(roomID, petID)
	if !found {
		return RespectResult{}, petrecord.ErrPetNotFound
	}
	age := service.runtime.Now().Sub(pet.CreatedAt)
	if age < 0 {
		age = 0
	}
	respect = RespectResult{Pet: pet, AgeDays: int32(age / (24 * time.Hour)), RequiredAgeDays: int32(service.config.RespectMinimumAge / (24 * time.Hour))}
	if age < service.config.RespectMinimumAge {
		respect.TooYoung = true
		result = petobservability.ResultRejected
		return respect, nil
	}
	if pet.OwnerPlayerID == actorID && !service.config.AllowRespectOwn {
		return respect, petrecord.ErrNoRights
	}
	dailyLimit := service.config.RespectDailyLimit
	if service.permissions != nil {
		bypass, permissionErr := service.permissions.HasPermission(ctx, actorID, petpolicy.RespectLimitBypass)
		if permissionErr != nil {
			return RespectResult{}, permissionErr
		}
		if bypass {
			dailyLimit = 0
		}
	}
	saved, applied, err := service.store.Respect(ctx, petID, actorID, service.config.RespectExperience, dailyLimit)
	if err != nil {
		return RespectResult{}, err
	}
	respect.Pet, respect.Applied = saved, applied
	if !applied {
		return respect, petrecord.ErrRespectQuota
	}
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		service.runtime.ProjectStatChange(ctx, active, pet, saved, service.config.RespectExperience, true)
	}
	service.runtime.Publish(ctx, petrespected.Name, petrespected.Payload{PetID: saved.ID, OwnerPlayerID: saved.OwnerPlayerID, ActorPlayerID: actorID, Respect: saved.Respect})
	return respect, nil
}

// Training sends one visible pet's learned-command panel.
func (service *Service) Training(ctx context.Context, target netconn.Context, roomID int64, petID int64) error {
	pet, found := service.runtime.Snapshot(roomID, petID)
	if !found {
		return petrecord.ErrPetNotFound
	}
	return service.runtime.SendTraining(ctx, target, pet)
}
