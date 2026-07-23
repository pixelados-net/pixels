package record

import "errors"

var (
	// ErrPetNotFound reports a missing or deleted pet.
	ErrPetNotFound = errors.New("pet not found")
	// ErrNoRights reports unauthorized pet mutation.
	ErrNoRights = errors.New("pet mutation not allowed")
	// ErrInventoryLimit reports a full owner inventory.
	ErrInventoryLimit = errors.New("pet inventory limit reached")
	// ErrRoomLimit reports a full room pet capacity.
	ErrRoomLimit = errors.New("room pet limit reached")
	// ErrTileNotFree reports an invalid placement destination.
	ErrTileNotFree = errors.New("pet tile not free")
	// ErrPetsDisabled reports room pet admission disabled.
	ErrPetsDisabled = errors.New("pets disabled in room")
	// ErrFeedingDisabled reports room pet feeding disabled.
	ErrFeedingDisabled = errors.New("pet feeding disabled in room")
	// ErrInvalidName reports a rejected pet name.
	ErrInvalidName = errors.New("invalid pet name")
	// ErrInvalidAppearance reports malformed or disabled appearance data.
	ErrInvalidAppearance = errors.New("invalid pet appearance")
	// ErrInvalidProduct reports unsupported product use.
	ErrInvalidProduct = errors.New("invalid pet product")
	// ErrInvalidState reports an invalid pet lifecycle transition.
	ErrInvalidState = errors.New("invalid pet state")
	// ErrConflict reports an optimistic concurrency conflict.
	ErrConflict = errors.New("pet changed concurrently")
	// ErrRespectQuota reports duplicate daily respect.
	ErrRespectQuota = errors.New("pet respect quota exhausted")
)
