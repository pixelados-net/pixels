package gallery

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestPurchaseCreatesPhotoAndChargesAtomically verifies the complete purchase mapping.
func TestPurchaseCreatesPhotoAndChargesAtomically(t *testing.T) {
	store := galleryStoreForTest()
	furniture := &galleryFurniture{}
	currencies := &galleryCurrencies{}
	metrics := camerametrics.New()
	service := New(store, furniture, currencies, galleryPlayers{}, nil, metrics)
	result, err := service.Purchase(context.Background(), 7)
	if err != nil {
		t.Fatalf("purchase photo: %v", err)
	}
	if store.purchases != 1 || result.Item.ID != 99 || len(currencies.mutations) != 2 || currencies.mutations[0].Amount != -2 || currencies.mutations[1].Amount != -3 {
		t.Fatalf("unexpected purchase result=%+v mutations=%+v", result, currencies.mutations)
	}
	want := `{"t":100,"u":"capture-uuid","s":40,"w":"https://storage/photo.png","m":"","n":"demo","o":"demo","oi":7}`
	if furniture.extraData != want {
		t.Fatalf("unexpected furniture photo data: %s", furniture.extraData)
	}
	if metrics.Snapshot().Purchases != 1 {
		t.Fatalf("purchase metric missing: %+v", metrics.Snapshot())
	}
}

// TestPurchaseRollsBackOnInsufficientBalance verifies expected debit mapping.
func TestPurchaseRollsBackOnInsufficientBalance(t *testing.T) {
	store := galleryStoreForTest()
	currencies := &galleryCurrencies{failType: CreditsType}
	service := New(store, &galleryFurniture{}, currencies, galleryPlayers{}, nil, camerametrics.New())
	if _, err := service.Purchase(context.Background(), 7); !errors.Is(err, camerarecord.ErrInsufficientCredits) {
		t.Fatalf("expected credits error, got %v", err)
	}
	if store.purchases != 0 {
		t.Fatal("capture purchase was linked after failed debit")
	}
}

// TestPublishEnforcesCooldownAndCreatesPublication verifies publication policy.
func TestPublishEnforcesCooldownAndCreatesPublication(t *testing.T) {
	now := time.Unix(200, 0)
	store := galleryStoreForTest()
	store.lastPublished, store.hasPublished = now.Add(-5*time.Second), true
	service := New(store, &galleryFurniture{}, &galleryCurrencies{}, galleryPlayers{}, nil, camerametrics.New())
	service.now = func() time.Time { return now }
	if _, remaining, err := service.Publish(context.Background(), 7); !errors.Is(err, camerarecord.ErrCooldown) || remaining != 5*time.Second {
		t.Fatalf("expected five second cooldown, remaining=%s err=%v", remaining, err)
	}
	store.hasPublished = false
	publication, remaining, err := service.Publish(context.Background(), 7)
	if err != nil || remaining != 0 || publication.URL != store.active.URL || !store.hasPublication || !store.hasPublished {
		t.Fatalf("unexpected publication=%+v remaining=%s err=%v", publication, remaining, err)
	}
}

// galleryStore stores focused camera transaction state.
type galleryStore struct {
	// transactionMu serializes the focused transaction fixture.
	transactionMu sync.Mutex
	// settings stores operational camera policy.
	settings camerarecord.Settings
	// active stores the reusable capture.
	active camerarecord.Capture
	// purchases counts linked furniture copies.
	purchases int
	// publication stores the idempotent gallery entry.
	publication camerarecord.Publication
	// hasPublication reports whether the capture was published.
	hasPublication bool
	// lastPublished stores the latest publication time.
	lastPublished time.Time
	// hasPublished reports whether a cooldown exists.
	hasPublished bool
}

// galleryStoreForTest returns a complete pending photo fixture.
func galleryStoreForTest() *galleryStore {
	return &galleryStore{settings: camerarecord.Settings{Enabled: true, CreditsPrice: 2, PointsPrice: 3, PointsType: 5, PublishPointsPrice: 10, PublishPointsType: 5, PublishCooldown: 10 * time.Second}, active: camerarecord.Capture{ID: 1, UUID: "capture-uuid", PlayerID: 7, RoomID: 40, Kind: camerarecord.KindPhoto, State: camerarecord.StatePending, URL: "https://storage/photo.png", CreatedAt: time.Unix(100, 0)}}
}

// WithinTransaction executes work immediately for focused tests.
func (store *galleryStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	store.transactionMu.Lock()
	defer store.transactionMu.Unlock()
	return work(ctx)
}

// CreateCapture returns the input capture.
func (*galleryStore) CreateCapture(_ context.Context, value camerarecord.Capture) (camerarecord.Capture, error) {
	return value, nil
}

// ActiveCapture returns the configured reusable photo.
func (store *galleryStore) ActiveCapture(context.Context, int64) (camerarecord.Capture, bool, error) {
	return store.active, store.active.ID > 0, nil
}

// LatestCaptureAt returns no capture time.
func (*galleryStore) LatestCaptureAt(context.Context, int64, camerarecord.Kind) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

// AttachPurchase records one linked furniture copy.
func (store *galleryStore) AttachPurchase(context.Context, int64, int64) error {
	store.purchases++
	return nil
}

// Settings returns configured policy.
func (store *galleryStore) Settings(context.Context) (camerarecord.Settings, error) {
	return store.settings, nil
}

// UpdateSettings reports no mutation.
func (*galleryStore) UpdateSettings(context.Context, camerarecord.Settings, int64) (camerarecord.Settings, bool, error) {
	return camerarecord.Settings{}, false, nil
}

// PublishCooldown returns configured publication state.
func (store *galleryStore) PublishCooldown(context.Context, int64) (time.Time, bool, error) {
	return store.lastPublished, store.hasPublished, nil
}

// SetPublishCooldown stores one publication time.
func (store *galleryStore) SetPublishCooldown(_ context.Context, _ int64, value time.Time) error {
	store.lastPublished, store.hasPublished = value, true
	return nil
}

// CreatePublication maps one capture to a gallery record.
func (store *galleryStore) CreatePublication(_ context.Context, capture camerarecord.Capture) (camerarecord.Publication, error) {
	store.publication = camerarecord.Publication{ID: 2, CaptureID: capture.ID, PlayerID: capture.PlayerID, RoomID: capture.RoomID, URL: capture.URL}
	store.hasPublication = true
	return store.publication, nil
}

// PublicationByCapture returns the configured idempotent publication.
func (store *galleryStore) PublicationByCapture(context.Context, int64) (camerarecord.Publication, bool, error) {
	return store.publication, store.hasPublication, nil
}

// Publications returns no publications.
func (*galleryStore) Publications(context.Context, int, int, bool) ([]camerarecord.Publication, error) {
	return nil, nil
}

// RemovePublication reports no mutation.
func (*galleryStore) RemovePublication(context.Context, int64, string) (bool, error) {
	return false, nil
}

// Captures returns no captures.
func (*galleryStore) Captures(context.Context, int64, int) ([]camerarecord.Capture, error) {
	return nil, nil
}

// ClaimCleanup returns no cleanup work.
func (*galleryStore) ClaimCleanup(context.Context, time.Time, time.Time, time.Time, int) ([]camerarecord.CleanupCandidate, error) {
	return nil, nil
}

// MarkDeleted reports no cleanup mutation.
func (*galleryStore) MarkDeleted(context.Context, int64, time.Time) (bool, error) { return false, nil }

// InsertAudit accepts one audit.
func (*galleryStore) InsertAudit(context.Context, camerarecord.Audit) error { return nil }

// galleryFurniture stores the granted photo payload.
type galleryFurniture struct {
	// extraData stores the granted Nitro photo JSON.
	extraData string
	// grants counts generated furniture identifiers.
	grants int64
}

// Grant returns one photo furniture fixture.
func (furniture *galleryFurniture) Grant(_ context.Context, params furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	furniture.extraData = params.ExtraData
	furniture.grants++
	return []furnituremodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 98 + furniture.grants}}, DefinitionID: PhotoDefinitionID, OwnerPlayerID: params.OwnerPlayerID, ExtraData: params.ExtraData}}, nil
}

// FindDefinitionByID returns the photo definition.
func (*galleryFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: PhotoDefinitionID}}, Kind: furnituremodel.KindWall, SpriteID: 4597}, true, nil
}

// galleryCurrencies records player-originated debits.
type galleryCurrencies struct {
	// mutations stores observed wallet changes.
	mutations []currencyservice.GrantParams
	// failType identifies one rejected currency type.
	failType int32
}

// Grant records or rejects one currency mutation.
func (currencies *galleryCurrencies) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currencies.mutations = append(currencies.mutations, params)
	if params.CurrencyType == currencies.failType {
		return 0, currencyservice.ErrInsufficientBalance
	}
	return 100, nil
}

// galleryPlayers resolves the buyer fixture.
type galleryPlayers struct{}

// FindByID returns the buyer fixture.
func (galleryPlayers) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, Username: "demo"}}, true, nil
}

// FindByUsername returns no fixture.
func (galleryPlayers) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}
