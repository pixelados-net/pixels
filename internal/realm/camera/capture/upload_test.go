package capture

import (
	"context"
	"errors"
	"reflect"
	"testing"

	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
)

// TestPhotoCompanionFailureRemovesCanonicalObject verifies partial uploads do not leak.
func TestPhotoCompanionFailureRemovesCanonicalObject(t *testing.T) {
	canonical := "photos/7/67445470-f8a2-4671-a049-3249c360f14c.png"
	companion := "photos/7/67445470-f8a2-4671-a049-3249c360f14c_small.png"
	store := &captureStore{}
	objects := &captureStorage{err: errors.New("companion failed"), failKey: companion}
	service := New(cameraConfigForTest(), store, objects, capturePermission(true), captureRooms{ownerID: 7}, captureRights(false), nil, camerametrics.New())
	service.uuid = func() string { return "67445470-f8a2-4671-a049-3249c360f14c" }
	if _, err := service.Photo(context.Background(), 7, 40, pngForTest); err == nil {
		t.Fatal("expected companion upload failure")
	}
	if !reflect.DeepEqual(objects.deleted, []string{canonical}) {
		t.Fatalf("unexpected rollback keys: %+v", objects.deleted)
	}
	if store.created.ID != 0 {
		t.Fatalf("unexpected persisted capture: %+v", store.created)
	}
}

// TestPhotoReceiptFailureRemovesBothObjects verifies untracked uploads do not leak.
func TestPhotoReceiptFailureRemovesBothObjects(t *testing.T) {
	canonical := "photos/7/67445470-f8a2-4671-a049-3249c360f14c.png"
	companion := "photos/7/67445470-f8a2-4671-a049-3249c360f14c_small.png"
	store := &captureStore{createErr: errors.New("receipt failed")}
	objects := &captureStorage{}
	service := New(cameraConfigForTest(), store, objects, capturePermission(true), captureRooms{ownerID: 7}, captureRights(false), nil, camerametrics.New())
	service.uuid = func() string { return "67445470-f8a2-4671-a049-3249c360f14c" }
	if _, err := service.Photo(context.Background(), 7, 40, pngForTest); err == nil {
		t.Fatal("expected receipt failure")
	}
	if !reflect.DeepEqual(objects.deleted, []string{companion, canonical}) {
		t.Fatalf("unexpected rollback keys: %+v", objects.deleted)
	}
}
