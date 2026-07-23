// Package camera wires camera capture, gallery, reporting, and compatibility.
package camera

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	cameraadmin "github.com/niflaot/pixels/internal/realm/camera/admin"
	cameracapture "github.com/niflaot/pixels/internal/realm/camera/capture"
	capturehandlers "github.com/niflaot/pixels/internal/realm/camera/capture/handlers"
	cameracleanup "github.com/niflaot/pixels/internal/realm/camera/cleanup"
	cameracompat "github.com/niflaot/pixels/internal/realm/camera/compat/handlers"
	cameraconfig "github.com/niflaot/pixels/internal/realm/camera/config"
	cameradb "github.com/niflaot/pixels/internal/realm/camera/database"
	cameragallery "github.com/niflaot/pixels/internal/realm/camera/gallery"
	galleryhandlers "github.com/niflaot/pixels/internal/realm/camera/gallery/handlers"
	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	camerareport "github.com/niflaot/pixels/internal/realm/camera/report"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/storage"
	"go.uber.org/fx"
)

// Module provides the complete camera and photos realm.
var Module = fx.Module("realm-camera", fx.Provide(cameraconfig.Load, cameradb.New, NewStore, NewCleanupStore, NewCleanupStorage, camerametrics.New, NewCaptureService, NewGalleryService, cameraadmin.New, capturehandlers.New, galleryhandlers.New, NewReportService, camerareport.NewHandler, cameracompat.New, cameracleanup.New, cameracleanup.NewScheduler), fx.Invoke(RegisterConnectionHandlers, cameracleanup.RegisterScheduler))

// NewStore exposes PostgreSQL persistence through the camera contract.
func NewStore(repository *cameradb.Repository) camerarecord.Store { return repository }

// NewCleanupStore exposes bounded camera cleanup persistence.
func NewCleanupStore(repository *cameradb.Repository) cameracleanup.Store { return repository }

// NewCleanupStorage exposes object deletion through the narrow cleanup boundary.
func NewCleanupStorage(client *storage.Client) cameracleanup.Storage { return client }

// NewCaptureService adapts concrete shared services to camera boundaries.
func NewCaptureService(config cameraconfig.Config, store camerarecord.Store, objects *storage.Client, permissions permissionservice.Checker, rooms *roomservice.Service, rights *roomrights.Service, events bus.Publisher, metrics *camerametrics.Metrics) *cameracapture.Service {
	return cameracapture.New(config, store, objects, permissions, rooms, rights, events, metrics)
}

// NewGalleryService adapts shared services to camera gallery boundaries.
func NewGalleryService(store camerarecord.Store, furniture *furnitureservice.Service, currencies currencyservice.Granter, players playerservice.Finder, events bus.Publisher, metrics *camerametrics.Metrics) *cameragallery.Service {
	return cameragallery.New(store, furniture, currencies, players, events, metrics)
}

// NewReportService adapts furniture and moderation to photo evidence intake.
func NewReportService(furniture *furnitureservice.Service, moderation *moderationcore.Service, metrics *camerametrics.Metrics) *camerareport.Service {
	return camerareport.New(furniture, moderation, metrics)
}

// RegisterConnectionHandlers registers every camera packet adapter.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, capture *capturehandlers.Handler, gallery *galleryhandlers.Handler, reports *camerareport.Handler, compatibility *cameracompat.Handler) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	capturehandlers.Register(handlers.Inbound, capture)
	galleryhandlers.Register(handlers.Inbound, gallery)
	camerareport.Register(handlers.Inbound, reports)
	cameracompat.Register(handlers.Inbound, compatibility)
}
