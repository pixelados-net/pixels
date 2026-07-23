package messenger

import "github.com/niflaot/pixels/internal/permission"

var (
	// FriendsUnlimited bypasses normal and club friend-list limits.
	FriendsUnlimited = permission.RegisterNode("messenger.friends.unlimited", "")
	// FollowAny bypasses a target player's follow privacy setting.
	FollowAny = permission.RegisterNode("messenger.follow.any", "")
)
