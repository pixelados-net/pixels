package service

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/player/repository"
)

// Service validates and coordinates player persistence behavior.
type Service struct {
	// store reads and writes player persistence records.
	store repository.Store
	// permissions assigns the default permission group.
	permissions permissionservice.DefaultAssigner
}

// New creates a player service.
func New(store repository.Store, assigners ...permissionservice.DefaultAssigner) *Service {
	service := &Service{store: store}
	if len(assigners) > 0 {
		service.permissions = assigners[0]
	}

	return service
}
