package promotion

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancel "github.com/niflaot/pixels/networking/outbound/room/promotion/cancel"
	outevent "github.com/niflaot/pixels/networking/outbound/room/promotion/event"
)

var (
	// ErrNotFound reports a missing promotion or room.
	ErrNotFound = errors.New("room promotion not found")
	// ErrNoRights reports an actor who cannot manage the requested room promotion.
	ErrNoRights = errors.New("room promotion rights required")
	// ErrInvalidCopy reports empty or oversized promotion text.
	ErrInvalidCopy = errors.New("invalid room promotion copy")
	// ErrInvalidOffer reports a catalog offer outside the Room Ads layout.
	ErrInvalidOffer = errors.New("invalid room promotion catalog offer")
)

// Service coordinates catalog charging, persistence, and live projection.
type Service struct {
	store       Store
	rooms       roomservice.Manager
	catalog     catalogservice.Manager
	permissions permissionservice.Checker
	manageAny   permission.Node
	runtime     *roomlive.Registry
	connections *netconn.Registry
	config      Config
	now         func() time.Time
}

// New creates room promotion behavior.
func New(config Config, store Store, rooms roomservice.Manager, catalog catalogservice.Manager, permissions permissionservice.Checker, manageAny permission.Node, runtime *roomlive.Registry, connections *netconn.Registry) *Service {
	return &Service{store: store, rooms: rooms, catalog: catalog, permissions: permissions, manageAny: manageAny, runtime: runtime, connections: connections, config: config.Normalize(), now: time.Now}
}

// Purchase atomically charges one catalog offer and creates or extends its room event.
func (service *Service) Purchase(ctx context.Context, params PurchaseParams) (Promotion, error) {
	room, err := service.authorizedRoom(ctx, params.RoomID, params.PlayerID)
	if err != nil {
		return Promotion{}, err
	}
	if err = validateCopy(params.Title, params.Description); err != nil {
		return Promotion{}, err
	}
	page, items, err := service.catalog.Page(ctx, params.PageID, params.PlayerID, params.HasClub)
	if err != nil || page.Layout != "room_ads" || !containsOffer(items, params.OfferID) {
		return Promotion{}, ErrInvalidOffer
	}
	var result Promotion
	var complete func(context.Context)
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		purchaser, ok := service.catalog.(catalogservice.TransactionalPurchaser)
		if !ok {
			return catalogservice.ErrCommerceUnavailable
		}
		_, postCommit, purchaseErr := purchaser.PurchaseWithin(txCtx, catalogservice.PurchaseParams{PlayerID: params.PlayerID, CatalogItemID: params.OfferID, HasClub: params.HasClub, Amount: 1})
		if purchaseErr != nil {
			return purchaseErr
		}
		complete = postCommit
		var upsertErr error
		result, upsertErr = service.store.Upsert(txCtx, params, service.config)
		return upsertErr
	})
	if err != nil {
		return Promotion{}, err
	}
	if complete != nil {
		complete(ctx)
	}
	_ = room
	return result, service.Broadcast(ctx, result)
}

// Edit changes active promotion copy without charging again.
func (service *Service) Edit(ctx context.Context, params EditParams) (Promotion, error) {
	if err := validateCopy(params.Title, params.Description); err != nil {
		return Promotion{}, err
	}
	current, found, err := service.store.FindByID(ctx, params.PromotionID)
	if err != nil {
		return Promotion{}, err
	}
	if !found || !current.ActiveAt(service.now()) {
		return Promotion{}, ErrNotFound
	}
	if _, err = service.authorizedRoom(ctx, current.RoomID, params.PlayerID); err != nil {
		return Promotion{}, err
	}
	updated, changed, err := service.store.UpdateCopy(ctx, params)
	if err != nil {
		return Promotion{}, err
	}
	if !changed {
		return Promotion{}, ErrNotFound
	}
	return updated, service.Broadcast(ctx, updated)
}

// Active returns one current room promotion.
func (service *Service) Active(ctx context.Context, roomID int64) (Promotion, bool, error) {
	return service.store.FindActiveByRoom(ctx, roomID)
}

// Cancel force-removes one room promotion and closes active client banners.
func (service *Service) Cancel(ctx context.Context, roomID int64) (bool, error) {
	deleted, err := service.store.DeleteByRoom(ctx, roomID)
	if err != nil || !deleted || service.runtime == nil || service.connections == nil {
		return deleted, err
	}
	active, found := service.runtime.Find(roomID)
	if !found {
		return true, nil
	}
	packet, err := outcancel.Encode()
	if err != nil {
		return false, err
	}
	for _, occupant := range active.Occupants() {
		if connection, exists := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID); exists {
			_ = connection.Send(ctx, packet)
		}
	}
	return true, nil
}

// EligibleRooms lists owned rooms and whether each already has an active promotion.
func (service *Service) EligibleRooms(ctx context.Context, playerID int64) ([]roommodel.Room, map[int64]struct{}, error) {
	rooms, err := service.rooms.ListByOwner(ctx, playerID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]int64, len(rooms))
	for index := range rooms {
		ids[index] = rooms[index].ID
	}
	active, err := service.store.ActiveRoomIDs(ctx, ids)
	return rooms, active, err
}

// SendActive sends one room's active event to a joining connection.
func (service *Service) SendActive(ctx context.Context, roomID int64, connection netconn.Context) error {
	promotion, found, err := service.Active(ctx, roomID)
	if err != nil || !found {
		return err
	}
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil || !found {
		return err
	}
	packet, err := service.packet(promotion, room)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// Broadcast projects a committed promotion to current room occupants.
func (service *Service) Broadcast(ctx context.Context, promotion Promotion) error {
	if service.runtime == nil || service.connections == nil {
		return nil
	}
	room, found, err := service.rooms.FindByID(ctx, promotion.RoomID)
	if err != nil || !found {
		return err
	}
	active, found := service.runtime.Find(promotion.RoomID)
	if !found {
		return nil
	}
	packet, err := service.packet(promotion, room)
	if err != nil {
		return err
	}
	for _, occupant := range active.Occupants() {
		if connection, exists := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID); exists {
			_ = connection.Send(ctx, packet)
		}
	}
	return nil
}

func (service *Service) authorizedRoom(ctx context.Context, roomID int64, playerID int64) (roommodel.Room, error) {
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !found {
		return roommodel.Room{}, ErrNotFound
	}
	if room.OwnerPlayerID == playerID {
		return room, nil
	}
	allowed := false
	if service.permissions != nil && service.manageAny.Concrete() {
		allowed, err = service.permissions.HasPermission(ctx, playerID, service.manageAny)
	}
	if err != nil {
		return roommodel.Room{}, err
	}
	if !allowed {
		return roommodel.Room{}, ErrNoRights
	}
	return room, nil
}

func (service *Service) packet(value Promotion, room roommodel.Room) (codec.Packet, error) {
	now := service.now()
	since := int32(now.Sub(value.StartsAt) / time.Minute)
	until := int32(value.EndsAt.Sub(now) / time.Minute)
	if since < 0 {
		since = 0
	}
	if until < 0 {
		until = 0
	}
	return outevent.Encode(outevent.Data{AdID: int32(value.ID), OwnerAvatarID: int32(value.CreatedBy), OwnerAvatarName: room.OwnerName, RoomID: int32(value.RoomID), EventType: value.CategoryID, Name: value.Title, Description: value.Description, MinutesSinceCreation: since, MinutesUntilExpiration: until, CategoryID: value.CategoryID})
}

func validateCopy(title string, description string) error {
	title = strings.TrimSpace(title)
	if title == "" || len([]rune(title)) > 64 || len([]rune(description)) > 256 {
		return ErrInvalidCopy
	}
	return nil
}
func containsOffer(items []catalogmodel.Item, offerID int64) bool {
	for _, item := range items {
		if item.ID == offerID {
			return true
		}
	}
	return false
}
