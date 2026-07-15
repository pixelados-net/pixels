package service

import (
	"context"
	"fmt"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// Service implements catalog browsing and purchase behavior.
type Service struct {
	// store persists catalog records and transaction boundaries.
	store catalogrepo.Store
	// commerce persists optional extended store capabilities.
	commerce catalogrepo.CommerceStore

	// currencies charges player balances.
	currencies currencyservice.Granter

	// furniture grants purchased inventory instances.
	furniture furnitureservice.DefinitionGranter

	// teleportPairs pairs teleport instances granted by one purchase.
	teleportPairs furnitureservice.TeleportPairer

	// cache stores one immutable catalog generation.
	cache *catalogCache

	// events publishes completed purchase facts.
	events bus.Publisher

	// log records post-commit projection failures.
	log *zap.Logger
	// permissions resolves optional catalog page requirements.
	permissions permissionservice.Checker

	// players resolves gift recipients.
	players playerservice.Finder

	// roomBundles clones and previews room template offers.
	roomBundles roombundle.Manager

	// effects grants catalog effect charges.
	effects playereffect.Manager

	// trophies formats immutable buyer inscriptions.
	trophies TrophyFormatter
}

// WithTrophies configures trophy inscription formatting.
func (service *Service) WithTrophies(formatter TrophyFormatter) *Service {
	service.trophies = formatter
	return service
}

// WithEffects configures transactional catalog effect grants.
func (service *Service) WithEffects(effects playereffect.Manager) *Service {
	service.effects = effects
	return service
}

// WithRoomBundles configures room bundle catalog behavior.
func (service *Service) WithRoomBundles(roomBundles roombundle.Manager) *Service {
	service.roomBundles = roomBundles
	return service
}

// WithPlayers configures player lookup for catalog gifts.
func (service *Service) WithPlayers(players playerservice.Finder) *Service {
	service.players = players
	return service
}

// New creates a catalog service.
func New(store catalogrepo.Store, currencies currencyservice.Granter, furniture furnitureservice.DefinitionGranter, events bus.Publisher, log *zap.Logger, checkers ...permissionservice.Checker) *Service {
	if log == nil {
		log = zap.NewNop()
	}

	service := &Service{store: store, currencies: currencies, furniture: furniture, cache: newCache(), events: events, log: log}
	service.commerce, _ = store.(catalogrepo.CommerceStore)
	if len(checkers) > 0 {
		service.permissions = checkers[0]
	}

	return service
}

// WithTeleportPairer configures transactional pairing for teleport offers.
func (service *Service) WithTeleportPairer(pairer furnitureservice.TeleportPairer) *Service {
	service.teleportPairs = pairer

	return service
}

// Refresh reloads the complete catalog cache.
func (service *Service) Refresh(ctx context.Context) error {
	pages, err := service.store.ListPages(ctx)
	if err != nil {
		return fmt.Errorf("refresh catalog pages: %w", err)
	}
	items, err := service.store.ListItems(ctx, nil)
	if err != nil {
		return fmt.Errorf("refresh catalog items: %w", err)
	}
	definitions, err := service.furniture.ListDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("refresh catalog furniture definitions: %w", err)
	}
	var products []catalogmodel.Product
	if service.commerce != nil {
		products, err = service.commerce.ListProducts(ctx)
		if err != nil {
			return fmt.Errorf("refresh catalog products: %w", err)
		}
	}

	service.cache.replace(pages, items, definitions, products)

	return nil
}

// Products returns preview products for one offer.
func (service *Service) Products(ctx context.Context, catalogItemID int64) []catalogmodel.Product {
	item, found := service.cache.item(catalogItemID)
	if found && item.IsRoomBundle() && service.roomBundles != nil {
		products, err := service.roomBundleProducts(ctx, item)
		if err != nil {
			service.log.Warn("room bundle preview failed", zap.Int64("catalog_item_id", catalogItemID), zap.Error(err))
			return nil
		}
		return products
	}
	return service.cache.products(catalogItemID)
}

// roomBundleProducts resolves a room bundle preview into catalog products.
func (service *Service) roomBundleProducts(ctx context.Context, item catalogmodel.Item) ([]catalogmodel.Product, error) {
	products, err := service.roomBundles.Preview(ctx, *item.RoomBundleTemplateRoomID)
	if err != nil {
		return nil, err
	}
	mapped := make([]catalogmodel.Product, len(products))
	for index := range products {
		mapped[index] = catalogmodel.Product{CatalogItemID: item.ID, DefinitionID: products[index].DefinitionID, Quantity: products[index].Quantity, OrderNum: int32(index)}
	}

	return mapped, nil
}

// Definition returns cached furniture metadata for one catalog offer.
func (service *Service) Definition(_ context.Context, definitionID int64) (furnituremodel.Definition, bool, error) {
	definition, found := service.cache.definition(definitionID)

	return definition, found, nil
}

// Pages returns pages visible to one player capability set.
func (service *Service) Pages(ctx context.Context, playerID int64, hasClub bool) ([]catalogmodel.Page, error) {
	pages := service.cache.pages()
	visible := make([]catalogmodel.Page, 0, len(pages))
	for _, page := range pages {
		accessible, err := service.pageAccessible(ctx, page, playerID, hasClub)
		if err != nil {
			return nil, err
		}
		if accessible {
			visible = append(visible, page)
		}
	}

	return visible, nil
}

// Page returns one visible page and its enabled offers.
func (service *Service) Page(ctx context.Context, pageID int64, playerID int64, hasClub bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	page, found := service.cache.page(pageID)
	if !found {
		return catalogmodel.Page{}, nil, ErrPageNotFound
	}
	accessible, err := service.pageAccessible(ctx, page, playerID, hasClub)
	if err != nil {
		return catalogmodel.Page{}, nil, err
	}
	if !accessible {
		return catalogmodel.Page{}, nil, ErrOfferNotVisible
	}

	items := service.cache.pageItems(pageID)
	visible := make([]catalogmodel.Item, 0, len(items))
	for _, item := range items {
		if item.Enabled && (!item.ClubOnly || hasClub) {
			visible = append(visible, item)
		}
	}

	return page, visible, nil
}

// SanitizeList returns definitions without an enabled active offer.
func (service *Service) SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error) {
	return service.store.SanitizeList(ctx)
}

// MarkNewAdditionsSeen records catalog novelty acknowledgement.
func (service *Service) MarkNewAdditionsSeen(ctx context.Context, playerID int64) error {
	if service.commerce == nil {
		return ErrCommerceUnavailable
	}
	return service.commerce.MarkNewAdditionsSeen(ctx, playerID)
}

// NewAdditionsAvailable reports whether one player has unseen novelty offers.
func (service *Service) NewAdditionsAvailable(ctx context.Context, playerID int64) (bool, error) {
	if service.commerce == nil {
		return false, nil
	}
	return service.commerce.NewAdditionsAvailable(ctx, playerID)
}

// CreditsSpentSince sums kickback-eligible catalog spending.
func (service *Service) CreditsSpentSince(ctx context.Context, playerID int64, since time.Time) (int64, error) {
	if service.commerce == nil {
		return 0, ErrCommerceUnavailable
	}
	return service.commerce.CreditsSpentSince(ctx, playerID, since)
}

// CreditsSpentBetween sums kickback-eligible spending inside one payday period.
func (service *Service) CreditsSpentBetween(ctx context.Context, playerID int64, after time.Time, through time.Time) (int64, error) {
	if service.commerce == nil {
		return 0, ErrCommerceUnavailable
	}
	return service.commerce.CreditsSpentBetween(ctx, playerID, after, through)
}
