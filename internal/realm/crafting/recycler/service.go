package recycler

import (
	"context"
	"sync"

	craftingconfig "github.com/niflaot/pixels/internal/realm/crafting/config"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	recycledevent "github.com/niflaot/pixels/internal/realm/crafting/recycler/events/recycled"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/pkg/bus"
)

// Service coordinates exact-batch recycler transactions.
type Service struct {
	// mutex protects mutable runtime configuration.
	mutex sync.RWMutex
	// config stores the active recycler policy.
	config craftingconfig.Config
	// store persists prize catalog state.
	store craftingrecord.Store
	// furniture validates and consumes inventory.
	furniture furnitureservice.TradingManager
	// granter creates the prize in the shared transaction.
	granter furnitureservice.Granter
	// rng chooses deterministic injected prizes.
	rng RNG
	// events publishes committed recycler outcomes.
	events bus.Publisher
}

// Result stores one committed recycle projection.
type Result struct {
	// Removed stores consumed item identifiers.
	Removed []int64
	// Granted stores the created prize instance.
	Granted furnituremodel.Item
	// Definition stores the prize furniture definition.
	Definition furnituremodel.Definition
	// Prize stores the selected prize metadata.
	Prize craftingrecord.Prize
}

// New creates a recycler service.
func New(config craftingconfig.Config, store craftingrecord.Store, furniture furnitureservice.TradingManager, granter furnitureservice.Granter, events bus.Publisher) *Service {
	return &Service{config: config, store: store, furniture: furniture, granter: granter, rng: Random(), events: events}
}

// NewWithRNG creates a recycler service with a deterministic source.
func NewWithRNG(config craftingconfig.Config, store craftingrecord.Store, furniture furnitureservice.TradingManager, granter furnitureservice.Granter, rng RNG) *Service {
	service := New(config, store, furniture, granter, nil)
	service.rng = rng
	return service
}

// Config returns an isolated runtime configuration snapshot.
func (service *Service) Config() craftingconfig.Config {
	service.mutex.RLock()
	defer service.mutex.RUnlock()
	return copyConfig(service.config)
}

// UpdateConfig atomically replaces validated runtime recycler policy.
func (service *Service) UpdateConfig(config craftingconfig.Config) error {
	if config.RecyclerBatchSize <= 0 || len(config.RecyclerRarityChance) == 0 {
		return craftingrecord.ErrInvalid
	}
	for tier, chance := range config.RecyclerRarityChance {
		if tier < 2 || tier > 5 || chance <= 0 {
			return craftingrecord.ErrInvalid
		}
	}
	service.mutex.Lock()
	service.config = copyConfig(config)
	service.mutex.Unlock()
	return nil
}

// Recycle validates, destroys, and replaces one exact furniture batch.
func (service *Service) Recycle(ctx context.Context, playerID int64, itemIDs []int64) (Result, error) {
	config := service.Config()
	if !config.RecyclerEnabled {
		return Result{}, craftingrecord.ErrRecyclerClosed
	}
	if len(itemIDs) != config.RecyclerBatchSize {
		return Result{}, craftingrecord.ErrRecyclerBatch
	}
	if duplicateIDs(itemIDs) {
		return Result{}, craftingrecord.ErrRecyclerBatch
	}
	result := Result{}
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		inventory, err := service.furniture.ListInventory(txCtx, playerID)
		if err != nil {
			return err
		}
		indexed := make(map[int64]furnituremodel.Item, len(inventory))
		for _, item := range inventory {
			indexed[item.ID] = item
		}
		definitions := make(map[int64]furnituremodel.Definition, len(itemIDs))
		for _, itemID := range itemIDs {
			item, found := indexed[itemID]
			if !found || !item.InInventory() || item.MarketplaceReserved {
				return craftingrecord.ErrItemUnavailable
			}
			definition, known := definitions[item.DefinitionID]
			if !known {
				definition, found, err = service.furniture.FindDefinitionByID(txCtx, item.DefinitionID)
				if err != nil {
					return err
				}
				if !found {
					return craftingrecord.ErrItemUnavailable
				}
				definitions[item.DefinitionID] = definition
			}
			if !definition.AllowRecycle {
				return craftingrecord.ErrItemUnavailable
			}
		}
		prizes, err := service.store.Prizes(txCtx)
		if err != nil {
			return err
		}
		prize, found := Draw(prizes, config.RecyclerRarityChance, service.rng)
		if !found {
			return craftingrecord.ErrRecyclerPrize
		}
		result.Prize = prize
		for _, itemID := range itemIDs {
			if err = service.furniture.DeleteInventoryItem(txCtx, itemID, playerID); err != nil {
				return craftingrecord.ErrItemUnavailable
			}
		}
		granted, err := service.granter.Grant(txCtx, furnitureservice.GrantParams{DefinitionID: prize.RewardDefinitionID, OwnerPlayerID: playerID, Quantity: 1})
		if err != nil {
			return err
		}
		if len(granted) != 1 {
			return craftingrecord.ErrRecyclerPrize
		}
		result.Removed = append([]int64(nil), itemIDs...)
		result.Granted = granted[0]
		result.Definition, found, err = service.furniture.FindDefinitionByID(txCtx, prize.RewardDefinitionID)
		if err != nil {
			return err
		}
		if !found {
			return craftingrecord.ErrRecyclerPrize
		}
		return nil
	})
	if err == nil && service.events != nil {
		_ = service.events.Publish(context.Background(), bus.Event{Name: recycledevent.Name, Payload: recycledevent.Payload{PlayerID: playerID, PrizeDefinitionID: result.Prize.RewardDefinitionID, ItemCount: int32(len(itemIDs))}})
	}
	return result, err
}

func duplicateIDs(ids []int64) bool {
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if _, found := seen[id]; found {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}
func copyConfig(config craftingconfig.Config) craftingconfig.Config {
	copied := config
	copied.RecyclerRarityChance = make(map[int32]int, len(config.RecyclerRarityChance))
	for tier, chance := range config.RecyclerRarityChance {
		copied.RecyclerRarityChance[tier] = chance
	}
	return copied
}
