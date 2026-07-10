package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	rightsgranted "github.com/niflaot/pixels/internal/realm/room/events/rightsgranted"
	rightsrevoked "github.com/niflaot/pixels/internal/realm/room/events/rightsrevoked"
	rightsmodel "github.com/niflaot/pixels/internal/realm/room/rights/model"
	rightsrepo "github.com/niflaot/pixels/internal/realm/room/rights/repository"
	"github.com/niflaot/pixels/pkg/bus"
)

// Nodes stores room rights administration permission nodes.
type Nodes struct {
	// OwnGrant allows owners to grant rights in their rooms.
	OwnGrant permission.Node
	// OwnRevoke allows owners to revoke rights in their rooms.
	OwnRevoke permission.Node
	// AnyGrant allows staff to grant rights in any room.
	AnyGrant permission.Node
	// AnyRevoke allows staff to revoke rights in any room.
	AnyRevoke permission.Node
}

// Service coordinates persistent room rights.
type Service struct {
	// store persists rights membership.
	store rightsrepo.Store
	// rooms resolves durable room ownership.
	rooms RoomFinder
	// permissions resolves global capability nodes.
	permissions permissionservice.Checker
	// events publishes committed rights changes.
	events bus.Publisher
	// nodes stores rights capability nodes.
	nodes Nodes
}

// New creates a room rights service.
func New(store rightsrepo.Store, rooms RoomFinder, permissions permissionservice.Checker, events bus.Publisher, nodes Nodes) *Service {
	return &Service{store: store, rooms: rooms, permissions: permissions, events: events, nodes: nodes}
}

// GrantRights grants build rights.
func (service *Service) GrantRights(ctx context.Context, roomID int64, actorID int64, playerID int64) error {
	room, err := service.authorize(ctx, roomID, actorID, service.nodes.OwnGrant, service.nodes.AnyGrant)
	if err != nil {
		return err
	}
	if playerID <= 0 {
		return ErrInvalidIdentity
	}
	if playerID == room.OwnerPlayerID {
		return ErrOwnerTarget
	}

	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		created, err := service.store.Grant(txCtx, roomID, playerID, actorID)
		if err != nil || !created {
			return err
		}

		return service.publish(txCtx, rightsgranted.Name, rightsgranted.Payload{RoomID: roomID, PlayerID: playerID, ActorID: actorID})
	})
}

// RevokeRights revokes one player's rights.
func (service *Service) RevokeRights(ctx context.Context, roomID int64, actorID int64, playerID int64) error {
	if _, err := service.authorize(ctx, roomID, actorID, service.nodes.OwnRevoke, service.nodes.AnyRevoke); err != nil {
		return err
	}
	if playerID <= 0 {
		return ErrInvalidIdentity
	}

	return service.revoke(ctx, roomID, actorID, playerID, rightsrevoked.ActionExplicit)
}

// RevokeAllRights revokes every rights holder and returns the count.
func (service *Service) RevokeAllRights(ctx context.Context, roomID int64, actorID int64) (int, error) {
	if _, err := service.authorize(ctx, roomID, actorID, service.nodes.OwnRevoke, service.nodes.AnyRevoke); err != nil {
		return 0, err
	}
	count := 0
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		rights, err := service.store.RevokeAll(txCtx, roomID)
		if err != nil {
			return err
		}
		count = len(rights)
		for _, right := range rights {
			payload := rightsrevoked.Payload{RoomID: roomID, PlayerID: right.PlayerID, ActorID: actorID, Action: rightsrevoked.ActionAll}
			if err := service.publish(txCtx, rightsrevoked.Name, payload); err != nil {
				return err
			}
		}

		return nil
	})

	return count, err
}

// RelinquishRights lets a player drop their own rights.
func (service *Service) RelinquishRights(ctx context.Context, roomID int64, playerID int64) error {
	if roomID <= 0 || playerID <= 0 {
		return ErrInvalidIdentity
	}

	return service.revoke(ctx, roomID, playerID, playerID, rightsrevoked.ActionRelinquished)
}

// ListRights lists current room rights holders.
func (service *Service) ListRights(ctx context.Context, roomID int64) ([]rightsmodel.Right, error) {
	if roomID <= 0 {
		return nil, ErrInvalidIdentity
	}

	return service.store.List(ctx, roomID)
}

// HasRights reports whether a player holds explicit room rights.
func (service *Service) HasRights(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	if roomID <= 0 || playerID <= 0 {
		return false, nil
	}

	return service.store.Exists(ctx, roomID, playerID)
}
