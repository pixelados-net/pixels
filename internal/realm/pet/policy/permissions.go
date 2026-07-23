package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// ManageAny allows administrative pet mutation.
	ManageAny = permission.RegisterNode("pet.manage.any", "")
	// PlaceAny allows placement in any manageable room.
	PlaceAny = permission.RegisterNode("pet.place.any", "")
	// RoomLimitBypass allows bypassing room pet limits.
	RoomLimitBypass = permission.RegisterNode("pet.room.limit.bypass", "")
	// InventoryLimitBypass allows bypassing inventory pet limits.
	InventoryLimitBypass = permission.RegisterNode("pet.inventory.limit.bypass", "")
	// RespectLimitBypass allows bypassing the ordinary respect quota.
	RespectLimitBypass = permission.RegisterNode("pet.respect.limit.bypass", "")
	// LifecycleManage allows administrative lifecycle and stat changes.
	LifecycleManage = permission.RegisterNode("pet.lifecycle.manage", "")
	// MoveAny allows directed movement of pets owned by others.
	MoveAny = permission.RegisterNode("pet.move.any", "")
)
