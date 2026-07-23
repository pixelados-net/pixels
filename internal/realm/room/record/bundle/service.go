package bundle

import (
	"context"
	"fmt"
	"strings"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
)

// Service implements room bundle cloning and administration.
type Service struct {
	// config stores optional bundle cloning policy.
	config Config
	// store persists atomic bundle records.
	store Store
	// rooms reads ordinary room records.
	rooms roomservice.Finder
	// layouts clones custom floor plans.
	layouts layout.CustomManager
	// furniture clones and summarizes room items.
	furniture furnitureservice.RoomBundleManager
	// bots clones placed room bots and their chat.
	bots BotCloner
}

// New creates room bundle behavior.
func New(config Config, store Store, rooms *roomservice.Service, layouts *layout.Service, furniture furnitureservice.RoomBundleManager, bots BotCloner) *Service {
	return &Service{config: config, store: store, rooms: rooms, layouts: layouts, furniture: furniture, bots: bots}
}

// Clone clones room data, custom geometry, and furniture atomically.
func (service *Service) Clone(ctx context.Context, params CloneParams) (CloneResult, error) {
	params.BuyerName = strings.TrimSpace(params.BuyerName)
	if params.TemplateRoomID <= 0 || params.BuyerPlayerID <= 0 || params.CatalogItemID <= 0 || params.BuyerName == "" {
		return CloneResult{}, ErrInvalidTemplate
	}
	template, found, err := service.FindTemplate(ctx, params.TemplateRoomID)
	if err != nil || !found {
		return CloneResult{}, err
	}
	var result CloneResult
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := service.store.LockRoomOwner(txCtx, params.BuyerPlayerID); err != nil {
			return err
		}
		count, err := service.store.CountRoomsByOwner(txCtx, params.BuyerPlayerID)
		if err != nil {
			return err
		}
		if count >= roomservice.MaxRoomsPerPlayer {
			return ErrRoomLimitReached
		}
		created, err := service.store.CloneBundleRoom(txCtx, template.ID, params.BuyerPlayerID, params.BuyerName)
		if err != nil {
			return err
		}
		if custom, customFound, findErr := service.layouts.FindCustomByRoomID(txCtx, template.ID); findErr != nil {
			return findErr
		} else if customFound {
			if _, err = service.layouts.SaveCustom(txCtx, customLayout(custom, created.ID)); err != nil {
				return err
			}
		}
		furnitureCount, err := service.furniture.CloneRoom(txCtx, template.ID, created.ID, params.BuyerPlayerID)
		if err != nil {
			return err
		}
		botCount := 0
		if service.config.CloneBots && service.bots != nil {
			botCount, err = service.bots.CloneRoom(txCtx, template.ID, created.ID, params.BuyerPlayerID)
			if err != nil {
				return err
			}
		}
		result = CloneResult{Room: created, FurnitureCount: furnitureCount, BotCount: botCount}
		return service.store.RecordBundlePurchase(txCtx, PurchaseRecord{CatalogItemID: params.CatalogItemID, TemplateRoomID: template.ID, CreatedRoomID: created.ID, BuyerPlayerID: params.BuyerPlayerID, FurnitureCount: furnitureCount, BotCount: botCount})
	})
	if err != nil {
		return CloneResult{}, fmt.Errorf("clone room bundle: %w", err)
	}
	return result, nil
}

// Preview groups template furniture without loading individual rows.
func (service *Service) Preview(ctx context.Context, templateRoomID int64) ([]Product, error) {
	_, found, err := service.FindTemplate(ctx, templateRoomID)
	if err != nil || !found {
		return nil, err
	}
	products, err := service.furniture.PreviewRoom(ctx, templateRoomID)
	if err != nil {
		return nil, err
	}
	result := make([]Product, len(products))
	for index := range products {
		result[index] = Product(products[index])
	}
	return result, nil
}

// customLayout maps custom geometry to a new room.
func customLayout(source layout.Layout, roomID int64) layout.CustomSaveParams {
	return layout.CustomSaveParams{RoomID: roomID, Heightmap: source.Heightmap, DoorX: source.DoorX, DoorY: source.DoorY, DoorZ: source.DoorZ, DoorDirection: source.DoorDirection, WallThickness: source.WallThickness, FloorThickness: source.FloorThickness, WallHeight: source.WallHeight}
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
