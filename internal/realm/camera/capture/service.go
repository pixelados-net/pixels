// Package capture owns camera uploads and room thumbnail authorization.
package capture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	camcaptured "github.com/niflaot/pixels/internal/realm/camera/capture/events/captured"
	cameraconfig "github.com/niflaot/pixels/internal/realm/camera/config"
	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerapolicy "github.com/niflaot/pixels/internal/realm/camera/policy"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Storage uploads and deletes durable camera objects.
type Storage interface {
	// Put uploads one object and returns its permanent public URL.
	Put(context.Context, string, io.Reader, int64, string) (string, error)
	// Delete removes one object.
	Delete(context.Context, string) error
}

// RoomFinder reads durable room ownership.
type RoomFinder interface {
	// FindByID finds one room.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// RightsChecker reads explicit room rights.
type RightsChecker interface {
	// HasRights reports whether a player holds explicit rights.
	HasRights(context.Context, int64, int64) (bool, error)
}

// Service validates and stores camera uploads.
type Service struct {
	// config stores immutable upload limits.
	config cameraconfig.Config
	// store persists capture receipts.
	store camerarecord.Store
	// storage uploads camera objects.
	storage Storage
	// permissions authorizes camera use.
	permissions permissionservice.Checker
	// rooms reads room ownership.
	rooms RoomFinder
	// rights reads explicit room rights.
	rights RightsChecker
	// events publishes committed camera facts.
	events bus.Publisher
	// metrics records bounded camera outcomes.
	metrics *camerametrics.Metrics
	// now supplies deterministic time.
	now func() time.Time
	// uuid supplies deterministic identifiers.
	uuid func() string
}

// New creates a camera capture service.
func New(config cameraconfig.Config, store camerarecord.Store, storage Storage, permissions permissionservice.Checker, rooms RoomFinder, rights RightsChecker, events bus.Publisher, metrics *camerametrics.Metrics) *Service {
	return &Service{config: config, store: store, storage: storage, permissions: permissions, rooms: rooms, rights: rights, events: events, metrics: metrics, now: time.Now, uuid: func() string { return uuid.NewString() }}
}

// Photo uploads one pending room photo.
func (service *Service) Photo(ctx context.Context, playerID int64, roomID int64, png []byte) (camerarecord.Capture, error) {
	if !validPNG(png) {
		return camerarecord.Capture{}, camerarecord.ErrInvalidPhoto
	}
	if err := service.authorize(ctx, playerID, len(png), service.config.MaxPhotoBytes); err != nil {
		return camerarecord.Capture{}, err
	}
	latest, found, err := service.store.LatestCaptureAt(ctx, playerID, camerarecord.KindPhoto)
	if err != nil {
		return camerarecord.Capture{}, err
	}
	if found && service.now().Sub(latest) < service.config.CaptureCooldown {
		return camerarecord.Capture{}, camerarecord.ErrCooldown
	}
	identifier := service.uuid()
	key := fmt.Sprintf("photos/%d/%s.png", playerID, identifier)
	return service.upload(ctx, camerarecord.Capture{UUID: identifier, PlayerID: playerID, RoomID: roomID, Kind: camerarecord.KindPhoto, StorageKey: key}, png, false, true)
}

// Thumbnail uploads one deterministic room thumbnail.
func (service *Service) Thumbnail(ctx context.Context, playerID int64, roomID int64, png []byte) (camerarecord.Capture, error) {
	if !validPNG(png) {
		return camerarecord.Capture{}, camerarecord.ErrInvalidPhoto
	}
	if err := service.authorize(ctx, playerID, len(png), service.config.MaxThumbnailBytes); err != nil {
		return camerarecord.Capture{}, err
	}
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil {
		return camerarecord.Capture{}, err
	}
	allowed := found && room.OwnerPlayerID == playerID
	if found && !allowed {
		allowed, err = service.rights.HasRights(ctx, roomID, playerID)
	}
	if err != nil {
		return camerarecord.Capture{}, err
	}
	if !allowed {
		return camerarecord.Capture{}, camerarecord.ErrNotRoomOwner
	}
	key := fmt.Sprintf("rooms/%d/thumbnail.png", roomID)
	return service.upload(ctx, camerarecord.Capture{UUID: service.uuid(), PlayerID: playerID, RoomID: roomID, Kind: camerarecord.KindThumbnail, StorageKey: key}, png, true, false)
}

// validPNG verifies the fixed PNG signature before storage access.
func validPNG(value []byte) bool {
	return len(value) >= 8 && bytes.Equal(value[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})
}

// authorize validates common upload policy before storage access.
func (service *Service) authorize(ctx context.Context, playerID int64, size int, maximum int) error {
	if !service.config.Enabled {
		return camerarecord.ErrDisabled
	}
	if playerID <= 0 || size == 0 || maximum <= 0 {
		return camerarecord.ErrNoPermission
	}
	if size > maximum {
		service.metrics.UploadFailed(camerametrics.UploadFailureTooLarge)
		return camerarecord.ErrTooLarge
	}
	allowed, err := service.permissions.HasPermission(ctx, playerID, camerapolicy.CaptureUse)
	if err != nil {
		return err
	}
	if !allowed {
		return camerarecord.ErrNoPermission
	}
	return nil
}

// upload stores bytes and their durable receipt.
func (service *Service) upload(ctx context.Context, capture camerarecord.Capture, png []byte, consumed bool, companion bool) (camerarecord.Capture, error) {
	url, err := service.storage.Put(ctx, capture.StorageKey, bytes.NewReader(png), int64(len(png)), "image/png")
	if err != nil {
		service.uploadFailed(err)
		return camerarecord.Capture{}, err
	}
	companionKey, hasCompanion := camerarecord.PhotoCompanionKey(capture.StorageKey)
	if companion && hasCompanion {
		if _, err = service.storage.Put(ctx, companionKey, bytes.NewReader(png), int64(len(png)), "image/png"); err != nil {
			service.uploadFailed(err)
			_ = service.storage.Delete(context.Background(), capture.StorageKey)
			return camerarecord.Capture{}, err
		}
	}
	capture.URL = url
	if consumed {
		now := service.now()
		capture.ConsumedAt = &now
	}
	created, err := service.store.CreateCapture(ctx, capture)
	if err != nil {
		service.metrics.UploadFailed(camerametrics.UploadFailureReceipt)
		if !consumed {
			if companion && hasCompanion {
				_ = service.storage.Delete(context.Background(), companionKey)
			}
			_ = service.storage.Delete(context.Background(), capture.StorageKey)
		}
		return camerarecord.Capture{}, err
	}
	service.publish(ctx, created)
	service.metrics.Capture(consumed, len(png))
	return created, nil
}

// uploadFailed records one bounded object-storage failure class.
func (service *Service) uploadFailed(err error) {
	reason := camerametrics.UploadFailureStorage
	if errors.Is(err, context.DeadlineExceeded) {
		reason = camerametrics.UploadFailureTimeout
	}
	service.metrics.UploadFailed(reason)
}

// publish emits one capture event after transaction commit when scoped.
func (service *Service) publish(ctx context.Context, capture camerarecord.Capture) {
	emit := func(eventCtx context.Context) {
		if service.events != nil {
			_ = service.events.Publish(eventCtx, bus.Event{Name: camcaptured.Name, Payload: camcaptured.Payload{CaptureID: capture.ID, PlayerID: capture.PlayerID, RoomID: capture.RoomID, Kind: string(capture.Kind)}})
		}
	}
	if !postgres.AfterCommit(ctx, emit) {
		emit(ctx)
	}
}
