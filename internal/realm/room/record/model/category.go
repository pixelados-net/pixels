package model

import sharedmodel "github.com/niflaot/pixels/pkg/model"

// Category describes a room category visible to navigator.
type Category struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// Caption stores the visible category caption.
	Caption string

	// CaptionKey stores the localization-safe caption key.
	CaptionKey string

	// Visible reports whether the category is exposed to clients.
	Visible bool

	// Automatic reports whether the category is system-managed.
	Automatic bool

	// AutomaticKey stores an optional automatic category key.
	AutomaticKey string

	// GlobalKey stores an optional global localization key.
	GlobalKey string

	// StaffOnly reports whether category use is restricted.
	StaffOnly bool

	// Order stores navigator display ordering.
	Order int
}

// Tag stores a normalized room tag.
type Tag struct {
	// RoomID identifies the tagged room.
	RoomID int64

	// Value stores the normalized tag value.
	Value string
}
