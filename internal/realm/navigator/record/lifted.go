package record

import (
	"time"

	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// LiftedRoom stores a promoted navigator room entry.
type LiftedRoom struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// RoomID identifies the promoted room.
	RoomID int64

	// AreaID stores the visual area id.
	AreaID int

	// Image stores the image key or URL.
	Image string

	// Caption stores the promotion caption.
	Caption string

	// Order stores navigator display ordering.
	Order int

	// StartsAt optionally stores promotion start time.
	StartsAt *time.Time

	// EndsAt optionally stores promotion end time.
	EndsAt *time.Time
}
