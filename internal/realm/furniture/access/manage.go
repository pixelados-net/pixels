// Package access resolves furniture management authorization.
package access

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

var (
	// ManageAny allows staff to manage furniture in any room.
	ManageAny = permission.RegisterNode("room.furniture.any.manage", "")
)

// CanManage reports whether a player has local build rights or global furniture authority.
func CanManage(ctx context.Context, checker permissionservice.Checker, room *roomlive.Room, playerID int64) (bool, error) {
	if room != nil && room.CanManageFurniture(playerID) {
		return true, nil
	}
	if checker == nil {
		return false, nil
	}

	return checker.HasPermission(ctx, playerID, ManageAny)
}
