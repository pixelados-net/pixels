package service

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/player/repository"
)

// Service validates and coordinates player persistence behavior.
type Service struct {
	// store reads and writes player persistence records.
	store repository.Store
	// clubs writes derived club entitlement when the store supports it.
	clubs repository.ClubWriter
	// admin writes complete administrative player mutations.
	admin repository.AdminWriter
	// permissions assigns the default permission group.
	permissions permissionservice.DefaultAssigner
}

// New creates a player service.
func New(store repository.Store, assigners ...permissionservice.DefaultAssigner) *Service {
	service := &Service{store: store}
	service.clubs, _ = store.(repository.ClubWriter)
	service.admin, _ = store.(repository.AdminWriter)
	if len(assigners) > 0 {
		service.permissions = assigners[0]
	}

	return service
}
