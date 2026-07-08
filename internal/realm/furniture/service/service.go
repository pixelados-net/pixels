// Package service contains furniture persistence rules.
package service

import "github.com/niflaot/pixels/internal/realm/furniture/repository"

// Service validates and coordinates furniture persistence behavior.
type Service struct {
	// store reads and writes furniture persistence records.
	store repository.Store
}

// New creates a furniture service.
func New(store repository.Store) *Service {
	return &Service{store: store}
}
