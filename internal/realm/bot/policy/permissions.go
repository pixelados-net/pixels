package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// AnyRoomOwner allows managing bots in any room.
	AnyRoomOwner = permission.RegisterNode("bot.any_room_owner", "")
	// PlaceAnywhere allows bot placement wherever furniture may be managed.
	PlaceAnywhere = permission.RegisterNode("bot.place_anywhere", "")
	// Unlimited allows bypassing room and inventory bot limits.
	Unlimited = permission.RegisterNode("bot.unlimited", "")
)
