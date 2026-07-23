// Package admin owns protected camera support and moderation workflows.
package admin

import (
	"context"
	"errors"
	"strings"

	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfloorremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outwallremove "github.com/niflaot/pixels/networking/outbound/room/furniture/wallremove"
)

// ErrConflict reports an optimistic settings conflict.
var ErrConflict = errors.New("camera settings version conflict")

// Service coordinates protected camera operations.
type Service struct {
	// store persists camera state and audit.
	store camerarecord.Store
	// furniture manages photo instances.
	furniture *furnitureservice.Service
	// rooms locates active room projections.
	rooms *roomlive.Registry
	// connections sends room furniture deltas.
	connections *netconn.Registry
}

// New creates a camera administration service.
func New(store camerarecord.Store, furniture *furnitureservice.Service, rooms *roomlive.Registry, connections *netconn.Registry) *Service {
	return &Service{store: store, furniture: furniture, rooms: rooms, connections: connections}
}

// Settings returns current camera pricing and policy.
func (service *Service) Settings(ctx context.Context) (camerarecord.Settings, error) {
	return service.store.Settings(ctx)
}

// UpdateSettings atomically replaces policy and appends audit attribution.
func (service *Service) UpdateSettings(ctx context.Context, settings camerarecord.Settings, expectedVersion int64, audit camerarecord.Audit) (camerarecord.Settings, error) {
	var updated camerarecord.Settings
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		value, found, err := service.store.UpdateSettings(txCtx, settings, expectedVersion)
		if err != nil {
			return err
		}
		if !found {
			return ErrConflict
		}
		updated = value
		audit.Action = "camera.settings.updated"
		return service.store.InsertAudit(txCtx, audit)
	})
	return updated, err
}

// Captures lists recent captures for one player.
func (service *Service) Captures(ctx context.Context, playerID int64, limit int) ([]camerarecord.Capture, error) {
	return service.store.Captures(ctx, playerID, boundedLimit(limit))
}

// Publications lists gallery entries with bounded pagination.
func (service *Service) Publications(ctx context.Context, limit int, offset int, includeRemoved bool) ([]camerarecord.Publication, error) {
	if offset < 0 {
		offset = 0
	}
	return service.store.Publications(ctx, boundedLimit(limit), offset, includeRemoved)
}

// RemovePublication soft-removes one gallery entry with audit attribution.
func (service *Service) RemovePublication(ctx context.Context, publicationID int64, audit camerarecord.Audit) (bool, error) {
	removed := false
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		removed, err = service.store.RemovePublication(txCtx, publicationID, audit.Reason)
		if err != nil || !removed {
			return err
		}
		audit.Action, audit.EntityID = "camera.publication.removed", publicationID
		return service.store.InsertAudit(txCtx, audit)
	})
	return removed, err
}

// DeletePhoto removes one external-image photo from inventory or a room.
func (service *Service) DeletePhoto(ctx context.Context, itemID int64, audit camerarecord.Audit) (bool, error) {
	var item furnituremodel.Item
	removed := false
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		foundItem, found, err := service.furniture.FindItemByID(txCtx, itemID)
		if err != nil || !found {
			return err
		}
		definition, found, err := service.furniture.FindDefinitionByID(txCtx, foundItem.DefinitionID)
		if err != nil || !found || definition.InteractionType != "external_image" {
			return err
		}
		item = foundItem
		if item.RoomID != nil {
			if _, err = service.furniture.Pickup(txCtx, furnitureservice.PickupParams{ItemID: item.ID, ActorPlayerID: audit.ActorPlayerID, RoomID: *item.RoomID, AllowForeign: true}); err != nil {
				return err
			}
		}
		if err = service.furniture.DeleteInventoryItem(txCtx, item.ID, item.OwnerPlayerID); err != nil {
			return err
		}
		audit.Action, audit.EntityID = "camera.photo.deleted", item.ID
		if err = service.store.InsertAudit(txCtx, audit); err != nil {
			return err
		}
		removed = true
		return nil
	})
	if err == nil && removed && item.RoomID != nil {
		service.projectRemoval(ctx, item)
	}
	return removed, err
}

// projectRemoval removes one deleted photo from active room clients.
func (service *Service) projectRemoval(ctx context.Context, item furnituremodel.Item) {
	active, found := service.rooms.Find(*item.RoomID)
	if !found {
		return
	}
	_, _ = active.ReloadFurniture(item.ID, nil)
	definition, found, err := service.furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found {
		return
	}
	if definition.Kind == furnituremodel.KindWall {
		packet, encodeErr := outwallremove.Encode(item.ID, item.OwnerPlayerID)
		if encodeErr == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
		return
	}
	packet, encodeErr := outfloorremove.Encode(item.ID, item.OwnerPlayerID)
	if encodeErr == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// ValidAudit reports whether required administrative attribution is present.
func ValidAudit(audit camerarecord.Audit) bool {
	return audit.ActorPlayerID > 0 && strings.TrimSpace(audit.Reason) != ""
}

// boundedLimit normalizes support query limits.
func boundedLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}
