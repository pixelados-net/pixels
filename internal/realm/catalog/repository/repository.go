package repository

import (
	"context"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

// PageReader reads catalog pages.
type PageReader interface {
	ListPages(context.Context) ([]catalogmodel.Page, error)
	FindPageByID(context.Context, int64) (catalogmodel.Page, bool, error)
}

// PageWriter writes catalog pages.
type PageWriter interface {
	CreatePage(context.Context, catalogmodel.Page) (catalogmodel.Page, error)
	UpdatePage(context.Context, catalogmodel.Page) (catalogmodel.Page, bool, error)
}

// ItemReader reads catalog offers.
type ItemReader interface {
	ListItems(context.Context, *int64) ([]catalogmodel.Item, error)
	FindItemByID(context.Context, int64) (catalogmodel.Item, bool, error)
	SanitizeList(context.Context) ([]furnituremodel.Definition, error)
	CountEnabledDefinitionsWithoutOffer(context.Context) (int64, error)
}

// ItemWriter writes catalog offers.
type ItemWriter interface {
	CreateItem(context.Context, catalogmodel.Item) (catalogmodel.Item, error)
	UpdateItem(context.Context, catalogmodel.Item) (catalogmodel.Item, bool, error)
	SoftDeleteItem(context.Context, int64, int64) (bool, error)
}

// LimitedWriter manages numbered LTD allocations.
type LimitedWriter interface {
	CreateLimitedUnits(context.Context, int64, int32) error
	SyncLimitedUnits(context.Context, int64, int32) error
	ReserveLimitedUnit(context.Context, int64, int64) (int32, bool, error)
	CompleteLimitedUnit(context.Context, int64, int32, int64, int64) (bool, error)
}

// CommerceStore persists extended catalog commerce state.
type CommerceStore interface {
	ListItemProducts(context.Context, int64) ([]catalogmodel.Product, error)
	ListProducts(context.Context) ([]catalogmodel.Product, error)
	FindVoucherByCode(context.Context, string) (catalogmodel.Voucher, bool, error)
	CountVoucherRedemptions(context.Context, int64) (int32, error)
	InsertVoucherRedemption(context.Context, int64, int64) error
	MarkNewAdditionsSeen(context.Context, int64) error
	NewAdditionsAvailable(context.Context, int64) (bool, error)
	LogPurchase(context.Context, int64, catalogmodel.Item, int32, int64, int64, []int64) error
	CreditsSpentSince(context.Context, int64, time.Time) (int64, error)
	// CreditsSpentBetween sums kickback-eligible purchases in one period.
	CreditsSpentBetween(context.Context, int64, time.Time, time.Time) (int64, error)
}

// VoucherAdminStore manages voucher administration records.
type VoucherAdminStore interface {
	ListVouchers(context.Context) ([]catalogmodel.Voucher, error)
	UpsertVoucher(context.Context, catalogmodel.Voucher) (catalogmodel.Voucher, error)
	ListVoucherRedemptions(context.Context, int64) ([]catalogmodel.VoucherRedemption, error)
}

// Store reads and mutates catalog persistence.
type Store interface {
	PageReader
	PageWriter
	ItemReader
	ItemWriter
	LimitedWriter
	WithinTransaction(context.Context, func(context.Context) error) error
}

// transactionRunner runs catalog work in one transaction.
type transactionRunner func(context.Context, func(context.Context) error) error

// Repository reads and writes catalog records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor

	// withinTx runs atomic catalog work.
	withinTx transactionRunner
}

// New creates a catalog repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{
		executor: pool,
		withinTx: func(ctx context.Context, work func(context.Context) error) error {
			return postgres.WithinScope(ctx, pool, work)
		},
	}
}

// newRepository creates a repository around a test executor.
func newRepository(executor postgres.Executor) *Repository {
	return &Repository{executor: executor, withinTx: func(ctx context.Context, work func(context.Context) error) error {
		return work(ctx)
	}}
}

// executorFor returns the active transaction or repository executor.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}

// WithinTransaction runs catalog purchase work atomically.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, ok := postgres.ScopedExecutor(ctx); ok {
		return work(ctx)
	}

	return repository.withinTx(ctx, work)
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
