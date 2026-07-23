package record

import "errors"

var (
	// ErrDisabled reports that camera operations are disabled.
	ErrDisabled = errors.New("camera is disabled")
	// ErrTooLarge reports an image above its configured byte limit.
	ErrTooLarge = errors.New("camera image is too large")
	// ErrCooldown reports an operation attempted before its cooldown elapsed.
	ErrCooldown = errors.New("camera cooldown is active")
	// ErrNoPendingCapture reports a missing purchasable or publishable photo.
	ErrNoPendingCapture = errors.New("camera has no active capture")
	// ErrNoPermission reports a missing camera permission.
	ErrNoPermission = errors.New("camera permission denied")
	// ErrNotRoomOwner reports missing thumbnail room rights.
	ErrNotRoomOwner = errors.New("camera room rights required")
	// ErrInvalidPhoto reports a furniture item that is not a placed photo.
	ErrInvalidPhoto = errors.New("camera photo is invalid")
	// ErrInsufficientCredits reports insufficient credits for purchase.
	ErrInsufficientCredits = errors.New("insufficient camera purchase credits")
	// ErrInsufficientPoints reports insufficient points for a camera operation.
	ErrInsufficientPoints = errors.New("insufficient camera points")
)
