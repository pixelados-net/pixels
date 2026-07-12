// Package service contains furniture persistence rules.
package service

import (
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

// Service validates and coordinates furniture persistence behavior.
type Service struct {
	// store reads and writes furniture persistence records.
	store repository.Store
}

// New creates a furniture service.
func New(store repository.Store) *Service {
	return &Service{store: store}
}

// validateActor validates common item and actor identifiers.
func validateActor(itemID int64, actorPlayerID int64) error {
	if itemID <= 0 {
		return ErrInvalidItemID
	}
	if actorPlayerID <= 0 {
		return ErrInvalidPlayerID
	}

	return nil
}

// validatePlacement validates floor placement input.
func validatePlacement(placement furnituremodel.Placement) error {
	if placement.X < 0 || placement.Y < 0 || placement.Z < 0 || !placement.Rotation.Valid() {
		return ErrInvalidPlacement
	}

	return nil
}
