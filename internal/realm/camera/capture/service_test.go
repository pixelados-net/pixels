package capture

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	cameraconfig "github.com/niflaot/pixels/internal/realm/camera/config"
	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// pngForTest stores a minimally recognizable bounded PNG payload.
var pngForTest = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 1}

// TestPhotoStoresPendingCapture verifies authorization, storage, and receipt creation.
func TestPhotoStoresPendingCapture(t *testing.T) {
	store := &captureStore{}
	objects := &captureStorage{}
	metrics := camerametrics.New()
	service := New(cameraConfigForTest(), store, objects, capturePermission(true), captureRooms{ownerID: 7}, captureRights(false), nil, metrics)
	service.uuid = func() string { return "67445470-f8a2-4671-a049-3249c360f14c" }
	service.now = func() time.Time { return time.Unix(100, 0) }
	capture, err := service.Photo(context.Background(), 7, 40, pngForTest)
	if err != nil {
		t.Fatalf("capture photo: %v", err)
	}
	if capture.StorageKey != "photos/7/67445470-f8a2-4671-a049-3249c360f14c.png" || capture.URL == "" || objects.puts != 2 || store.created.Kind != camerarecord.KindPhoto {
		t.Fatalf("unexpected capture=%+v storage=%+v", capture, objects)
	}
	if len(objects.keys) != 2 || objects.keys[1] != "photos/7/67445470-f8a2-4671-a049-3249c360f14c_small.png" {
		t.Fatalf("unexpected storage keys: %+v", objects.keys)
	}
	if metrics.Snapshot().Photos != 1 || metrics.Snapshot().UploadedBytes != uint64(len(pngForTest)) {
		t.Fatalf("unexpected metrics: %+v", metrics.Snapshot())
	}
}

// TestPhotoRejectsInvalidCooldownAndSize verifies every pre-storage guard.
func TestPhotoRejectsInvalidCooldownAndSize(t *testing.T) {
	now := time.Unix(100, 0)
	store := &captureStore{latest: now.Add(-time.Second), hasLatest: true}
	objects := &captureStorage{}
	service := New(cameraConfigForTest(), store, objects, capturePermission(true), captureRooms{ownerID: 7}, captureRights(false), nil, camerametrics.New())
	service.now = func() time.Time { return now }
	if _, err := service.Photo(context.Background(), 7, 40, pngForTest); !errors.Is(err, camerarecord.ErrCooldown) {
		t.Fatalf("expected cooldown, got %v", err)
	}
	oversized := append(append([]byte(nil), pngForTest...), make([]byte, 100)...)
	service.config.MaxPhotoBytes = 8
	if _, err := service.Photo(context.Background(), 7, 40, oversized); !errors.Is(err, camerarecord.ErrTooLarge) {
		t.Fatalf("expected too large, got %v", err)
	}
	if _, err := service.Photo(context.Background(), 7, 40, []byte("not png")); !errors.Is(err, camerarecord.ErrInvalidPhoto) {
		t.Fatalf("expected invalid png, got %v", err)
	}
	if objects.puts != 0 {
		t.Fatalf("pre-storage validation performed %d uploads", objects.puts)
	}
}

// TestThumbnailRequiresOwnerOrRights verifies authoritative room authorization.
func TestThumbnailRequiresOwnerOrRights(t *testing.T) {
	store := &captureStore{}
	objects := &captureStorage{}
	service := New(cameraConfigForTest(), store, objects, capturePermission(true), captureRooms{ownerID: 8}, captureRights(false), nil, camerametrics.New())
	if _, err := service.Thumbnail(context.Background(), 7, 40, pngForTest); !errors.Is(err, camerarecord.ErrNotRoomOwner) {
		t.Fatalf("expected rights rejection, got %v", err)
	}
	service.rights = captureRights(true)
	capture, err := service.Thumbnail(context.Background(), 7, 40, pngForTest)
	if err != nil {
		t.Fatalf("capture thumbnail: %v", err)
	}
	if capture.StorageKey != "rooms/40/thumbnail.png" || capture.ConsumedAt == nil {
		t.Fatalf("unexpected thumbnail: %+v", capture)
	}
}

// cameraConfigForTest returns deterministic permissive upload limits.
func cameraConfigForTest() cameraconfig.Config {
	return cameraconfig.Config{Enabled: true, CaptureCooldown: 3 * time.Second, MaxPhotoBytes: 100, MaxThumbnailBytes: 100}
}

// captureStorage stores bounded upload observations.
type captureStorage struct {
	// puts counts upload attempts.
	puts int
	// err stores an injected upload failure.
	err error
	// failKey limits the injected upload failure to one object key.
	failKey string
	// keys stores uploaded object keys.
	keys []string
	// deleted stores cleanup object keys.
	deleted []string
}

// Put records one uploaded body.
func (storage *captureStorage) Put(_ context.Context, key string, body io.Reader, size int64, _ string) (string, error) {
	storage.puts++
	storage.keys = append(storage.keys, key)
	if storage.err != nil && (storage.failKey == "" || storage.failKey == key) {
		return "", storage.err
	}
	_, _ = io.Copy(io.Discard, body)
	return "https://storage/" + key, nil
}

// Delete records cleanup requests.
func (storage *captureStorage) Delete(_ context.Context, key string) error {
	storage.deleted = append(storage.deleted, key)
	return nil
}

// capturePermission stores one permission decision.
type capturePermission bool

// HasPermission returns the configured decision.
func (allowed capturePermission) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return bool(allowed), nil
}

// captureRooms stores one room owner.
type captureRooms struct {
	// ownerID identifies the room owner.
	ownerID int64
}

// FindByID returns one room fixture.
func (rooms captureRooms) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 40}}, OwnerPlayerID: rooms.ownerID}, true, nil
}

// captureRights stores one explicit rights decision.
type captureRights bool

// HasRights returns the configured decision.
func (rights captureRights) HasRights(context.Context, int64, int64) (bool, error) {
	return bool(rights), nil
}

// captureStore stores focused capture state.
type captureStore struct {
	// created stores the persisted capture.
	created camerarecord.Capture
	// latest stores the latest capture timestamp.
	latest time.Time
	// hasLatest reports whether a latest capture exists.
	hasLatest bool
	// createErr stores an injected receipt persistence failure.
	createErr error
}

// WithinTransaction executes work immediately.
func (store *captureStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// CreateCapture stores and returns one capture.
func (store *captureStore) CreateCapture(_ context.Context, value camerarecord.Capture) (camerarecord.Capture, error) {
	if store.createErr != nil {
		return camerarecord.Capture{}, store.createErr
	}
	value.ID, value.CreatedAt = 1, time.Unix(100, 0)
	store.created = value
	return value, nil
}

// LatestCaptureAt returns configured cooldown state.
func (store *captureStore) LatestCaptureAt(context.Context, int64, camerarecord.Kind) (time.Time, bool, error) {
	return store.latest, store.hasLatest, nil
}

// ActiveCapture returns no capture.
func (*captureStore) ActiveCapture(context.Context, int64) (camerarecord.Capture, bool, error) {
	return camerarecord.Capture{}, false, nil
}

// AttachPurchase accepts no purchase link.
func (*captureStore) AttachPurchase(context.Context, int64, int64) error { return nil }

// Settings returns empty settings.
func (*captureStore) Settings(context.Context) (camerarecord.Settings, error) {
	return camerarecord.Settings{}, nil
}

// UpdateSettings reports no mutation.
func (*captureStore) UpdateSettings(context.Context, camerarecord.Settings, int64) (camerarecord.Settings, bool, error) {
	return camerarecord.Settings{}, false, nil
}

// PublishCooldown returns no cooldown.
func (*captureStore) PublishCooldown(context.Context, int64) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

// SetPublishCooldown accepts one cooldown.
func (*captureStore) SetPublishCooldown(context.Context, int64, time.Time) error { return nil }

// CreatePublication returns an empty publication.
func (*captureStore) CreatePublication(context.Context, camerarecord.Capture) (camerarecord.Publication, error) {
	return camerarecord.Publication{}, nil
}

// PublicationByCapture returns no publication.
func (*captureStore) PublicationByCapture(context.Context, int64) (camerarecord.Publication, bool, error) {
	return camerarecord.Publication{}, false, nil
}

// Publications returns no publications.
func (*captureStore) Publications(context.Context, int, int, bool) ([]camerarecord.Publication, error) {
	return nil, nil
}

// RemovePublication reports no mutation.
func (*captureStore) RemovePublication(context.Context, int64, string) (bool, error) {
	return false, nil
}

// Captures returns no captures.
func (*captureStore) Captures(context.Context, int64, int) ([]camerarecord.Capture, error) {
	return nil, nil
}

// ClaimCleanup returns no cleanup work.
func (*captureStore) ClaimCleanup(context.Context, time.Time, time.Time, time.Time, int) ([]camerarecord.CleanupCandidate, error) {
	return nil, nil
}

// MarkDeleted reports no cleanup mutation.
func (*captureStore) MarkDeleted(context.Context, int64, time.Time) (bool, error) { return false, nil }

// InsertAudit accepts one audit.
func (*captureStore) InsertAudit(context.Context, camerarecord.Audit) error { return nil }
