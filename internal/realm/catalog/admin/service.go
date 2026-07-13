package admin

import (
	"context"
	"strings"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
)

// Service implements catalog administration behavior.
type Service struct {
	// store persists catalog records.
	store catalogrepo.Store
	// vouchers persists voucher administration records.
	vouchers catalogrepo.VoucherAdminStore
	// catalog refreshes player-facing catalog data.
	catalog catalogservice.Manager
	// definitions validates furniture references.
	definitions furnitureservice.DefinitionGranter
	// roomBundles validates room template references.
	roomBundles roombundle.Manager
}

// WithRoomBundles configures room template validation.
func (service *Service) WithRoomBundles(roomBundles roombundle.Manager) *Service {
	service.roomBundles = roomBundles
	return service
}

// New creates a catalog administration service.
func New(store catalogrepo.Store, catalog catalogservice.Manager, definitions furnitureservice.DefinitionGranter) *Service {
	vouchers, _ := store.(catalogrepo.VoucherAdminStore)
	return &Service{store: store, vouchers: vouchers, catalog: catalog, definitions: definitions}
}

// Refresh reloads the player-facing catalog cache.
func (service *Service) Refresh(ctx context.Context) error {
	return service.catalog.Refresh(ctx)
}

// Pages lists all active catalog pages without player access filtering.
func (service *Service) Pages(ctx context.Context) ([]catalogmodel.Page, error) {
	return service.store.ListPages(ctx)
}

// Items lists active offers with an optional page filter.
func (service *Service) Items(ctx context.Context, pageID *int64) ([]catalogmodel.Item, error) {
	return service.store.ListItems(ctx, pageID)
}

// SanitizeList lists furniture definitions without active offers.
func (service *Service) SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error) {
	return service.store.SanitizeList(ctx)
}

// refresh reloads player-facing catalog data after a mutation.
func (service *Service) refresh(ctx context.Context) error {
	return service.catalog.Refresh(ctx)
}

// Vouchers lists every voucher.
func (service *Service) Vouchers(ctx context.Context) ([]catalogmodel.Voucher, error) {
	if service.vouchers == nil {
		return nil, ErrInvalidItem
	}
	return service.vouchers.ListVouchers(ctx)
}

// SaveVoucher validates and stores one voucher.
func (service *Service) SaveVoucher(ctx context.Context, voucher catalogmodel.Voucher) (catalogmodel.Voucher, error) {
	voucher.Code = strings.ToUpper(strings.TrimSpace(voucher.Code))
	if service.vouchers == nil || len(voucher.Code) < 4 || len(voucher.Code) > 32 || voucher.CostCredits < 0 ||
		voucher.CostPoints < 0 || voucher.PerPlayerCap <= 0 || voucher.RedemptionCap != nil && *voucher.RedemptionCap <= 0 {
		return catalogmodel.Voucher{}, ErrInvalidItem
	}
	return service.vouchers.UpsertVoucher(ctx, voucher)
}

// VoucherRedemptions lists voucher redemption history.
func (service *Service) VoucherRedemptions(ctx context.Context, voucherID int64) ([]catalogmodel.VoucherRedemption, error) {
	if service.vouchers == nil || voucherID <= 0 {
		return nil, ErrInvalidItem
	}
	return service.vouchers.ListVoucherRedemptions(ctx, voucherID)
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
