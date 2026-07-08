package service

import furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"

// validateActor validates common item/actor identifiers.
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
	if placement.X < 0 || placement.Y < 0 || placement.Z < 0 {
		return ErrInvalidPlacement
	}
	if !placement.Rotation.Valid() {
		return ErrInvalidPlacement
	}

	return nil
}
