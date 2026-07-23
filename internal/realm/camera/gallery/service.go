// Package gallery owns camera configuration, purchases, and publication.
package gallery

import (
	"context"
	"errors"
	"time"

	camerapublished "github.com/niflaot/pixels/internal/realm/camera/gallery/events/published"
	camerapurchased "github.com/niflaot/pixels/internal/realm/camera/gallery/events/purchased"
	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// PhotoDefinitionID identifies the seeded external-image photo furniture.
	PhotoDefinitionID int64 = 940001
	// CreditsType identifies the credits wallet.
	CreditsType int32 = -1
)

// PurchaseResult contains one committed photo purchase.
type PurchaseResult struct {
	// Capture stores the reusable source capture.
	Capture camerarecord.Capture
	// Item stores the granted photo furniture.
	Item furnituremodel.Item
	// Definition stores the photo furniture definition.
	Definition furnituremodel.Definition
}

// Furniture reads definitions and grants photo inventory instances.
type Furniture interface {
	// Grant creates photo furniture.
	Grant(context.Context, furnitureservice.GrantParams) ([]furnituremodel.Item, error)
	// FindDefinitionByID finds the photo furniture definition.
	FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error)
}

// Service coordinates atomic camera gallery workflows.
type Service struct {
	// store persists camera state.
	store camerarecord.Store
	// furniture creates photo furniture.
	furniture Furniture
	// currencies applies player-originated debits.
	currencies currencyservice.Granter
	// players resolves photo owner names.
	players playerservice.Finder
	// events publishes committed camera facts.
	events bus.Publisher
	// metrics records bounded camera outcomes.
	metrics *camerametrics.Metrics
	// now supplies deterministic time.
	now func() time.Time
}

// New creates a camera gallery service.
func New(store camerarecord.Store, furniture Furniture, currencies currencyservice.Granter, players playerservice.Finder, events bus.Publisher, metrics *camerametrics.Metrics) *Service {
	return &Service{store: store, furniture: furniture, currencies: currencies, players: players, events: events, metrics: metrics, now: time.Now}
}

// Configuration returns the current camera prices and policy.
func (service *Service) Configuration(ctx context.Context) (camerarecord.Settings, error) {
	return service.store.Settings(ctx)
}

// Purchase atomically grants one reusable capture copy and charges the buyer.
func (service *Service) Purchase(ctx context.Context, playerID int64) (PurchaseResult, error) {
	var result PurchaseResult
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		settings, err := service.store.Settings(txCtx)
		if err != nil {
			return err
		}
		if !settings.Enabled {
			return camerarecord.ErrDisabled
		}
		capture, found, err := service.store.ActiveCapture(txCtx, playerID)
		if err != nil {
			return err
		}
		if !found {
			return camerarecord.ErrNoPendingCapture
		}
		player, found, err := service.players.FindByID(txCtx, playerID)
		if err != nil {
			return err
		}
		if !found {
			return camerarecord.ErrNoPermission
		}
		extraData, err := (camerarecord.PhotoData{Timestamp: capture.CreatedAt.Unix(), UUID: capture.UUID, RoomID: capture.RoomID, URL: capture.URL, Name: player.Player.Username, OwnerName: player.Player.Username, OwnerID: playerID}).JSON()
		if err != nil {
			return err
		}
		items, err := service.furniture.Grant(txCtx, furnitureservice.GrantParams{DefinitionID: PhotoDefinitionID, OwnerPlayerID: playerID, Quantity: 1, ExtraData: extraData})
		if err != nil {
			return err
		}
		if err = service.charge(txCtx, playerID, CreditsType, settings.CreditsPrice, camerarecord.ErrInsufficientCredits, "camera photo purchase"); err != nil {
			return err
		}
		if err = service.charge(txCtx, playerID, settings.PointsType, settings.PointsPrice, camerarecord.ErrInsufficientPoints, "camera photo purchase"); err != nil {
			return err
		}
		definition, found, err := service.furniture.FindDefinitionByID(txCtx, PhotoDefinitionID)
		if err != nil {
			return err
		}
		if !found || len(items) != 1 {
			return camerarecord.ErrNoPendingCapture
		}
		if err = service.store.AttachPurchase(txCtx, capture.ID, items[0].ID); err != nil {
			return err
		}
		result = PurchaseResult{Capture: capture, Item: items[0], Definition: definition}
		return nil
	})
	if err == nil {
		service.publishPurchase(ctx, result)
		service.metrics.Purchased()
	}
	return result, err
}

// Publish atomically and idempotently charges for one gallery entry.
func (service *Service) Publish(ctx context.Context, playerID int64) (camerarecord.Publication, time.Duration, error) {
	var publication camerarecord.Publication
	var remaining time.Duration
	created := false
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		settings, err := service.store.Settings(txCtx)
		if err != nil {
			return err
		}
		if !settings.Enabled {
			return camerarecord.ErrDisabled
		}
		capture, found, err := service.store.ActiveCapture(txCtx, playerID)
		if err != nil {
			return err
		}
		if !found {
			return camerarecord.ErrNoPendingCapture
		}
		publication, found, err = service.store.PublicationByCapture(txCtx, capture.ID)
		if err != nil {
			return err
		}
		if found {
			return nil
		}
		last, found, err := service.store.PublishCooldown(txCtx, playerID)
		if err != nil {
			return err
		}
		now := service.now()
		if found && now.Sub(last) < settings.PublishCooldown {
			remaining = settings.PublishCooldown - now.Sub(last)
			return camerarecord.ErrCooldown
		}
		if err = service.charge(txCtx, playerID, settings.PublishPointsType, settings.PublishPointsPrice, camerarecord.ErrInsufficientPoints, "camera photo publication"); err != nil {
			return err
		}
		publication, err = service.store.CreatePublication(txCtx, capture)
		if err != nil {
			return err
		}
		created = true
		return service.store.SetPublishCooldown(txCtx, playerID, now)
	})
	if err == nil && created {
		service.publishPublication(ctx, publication)
		service.metrics.Published()
	}
	return publication, remaining, err
}

// charge applies one optional player-originated camera debit.
func (service *Service) charge(ctx context.Context, playerID int64, currencyType int32, amount int64, insufficient error, reason string) error {
	if amount == 0 {
		return nil
	}
	actorID := playerID
	_, err := service.currencies.Grant(ctx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: currencyType, Amount: -amount, Reason: reason, ActorKind: currencyservice.ActorPlayer, ActorID: &actorID})
	if errors.Is(err, currencyservice.ErrInsufficientBalance) {
		return insufficient
	}
	return err
}

// publishPurchase emits one committed purchase fact.
func (service *Service) publishPurchase(ctx context.Context, result PurchaseResult) {
	service.emit(ctx, camerapurchased.Name, camerapurchased.Payload{CaptureID: result.Capture.ID, PlayerID: result.Capture.PlayerID, ItemID: result.Item.ID})
}

// publishPublication emits one committed publication fact.
func (service *Service) publishPublication(ctx context.Context, publication camerarecord.Publication) {
	service.emit(ctx, camerapublished.Name, camerapublished.Payload{PublicationID: publication.ID, CaptureID: publication.CaptureID, PlayerID: publication.PlayerID})
}

// emit publishes one event after commit when transaction-scoped.
func (service *Service) emit(ctx context.Context, name string, payload any) {
	work := func(eventCtx context.Context) {
		if service.events != nil {
			_ = service.events.Publish(eventCtx, bus.Event{Name: bus.Name(name), Payload: payload})
		}
	}
	if !postgres.AfterCommit(ctx, work) {
		work(ctx)
	}
}
