package service

import "github.com/niflaot/pixels/internal/realm/navigator/repository"

// Service validates and coordinates navigator persistence behavior.
type Service struct {
	// store reads and writes navigator persistence records.
	store repository.Store
}

// New creates a navigator service.
func New(store repository.Store) *Service {
	return &Service{store: store}
}
