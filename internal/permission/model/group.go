// Package model contains durable permission records.
package model

import (
	"github.com/niflaot/pixels/internal/permission"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// Group contains one persistent permission group.
type Group struct {
	// Base contains shared durable record fields.
	sharedmodel.Base
	// Name stores the unique group name.
	Name string
	// Weight stores group resolution priority and client security level.
	Weight int32
	// Prefix stores a future localized chat prefix.
	Prefix string
	// PrefixColor stores a future chat prefix color.
	PrefixColor string
	// RoomEffectID stores the synthetic room effect inherited from this group.
	RoomEffectID *int32
	// ParentGroupID identifies the optional inherited group.
	ParentGroupID *int64
}

// HolderID identifies the permission group.
func (group Group) HolderID() int64 {
	return group.ID
}

// HolderKind reports that Group is a group permission holder.
func (group Group) HolderKind() permission.HolderKind {
	return permission.HolderGroup
}
