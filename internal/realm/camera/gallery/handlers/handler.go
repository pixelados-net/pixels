// Package handlers adapts camera gallery packets to domain workflows.
package handlers

import (
	"context"
	"errors"
	"math"
	"time"

	cameragallery "github.com/niflaot/pixels/internal/realm/camera/gallery"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnitureprojection "github.com/niflaot/pixels/internal/realm/furniture/projection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inconfiguration "github.com/niflaot/pixels/networking/inbound/camera/configuration"
	inpublish "github.com/niflaot/pixels/networking/inbound/camera/publish"
	inpurchase "github.com/niflaot/pixels/networking/inbound/camera/purchase"
	outinit "github.com/niflaot/pixels/networking/outbound/camera/init"
	outpublish "github.com/niflaot/pixels/networking/outbound/camera/publishstatus"
	outpurchase "github.com/niflaot/pixels/networking/outbound/camera/purchaseok"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler handles camera configuration, purchase, and publication packets.
type Handler struct {
	// Service executes gallery workflows.
	Service *cameragallery.Service
	// Players resolves live players.
	Players *playerlive.Registry
	// Bindings resolves authenticated connections.
	Bindings *binding.Registry
	// Translations resolves hotel-facing errors.
	Translations i18n.Translator
}

// New creates a grouped camera gallery packet handler.
func New(service *cameragallery.Service, players *playerlive.Registry, bindings *binding.Registry, translations i18n.Translator) *Handler {
	return &Handler{Service: service, Players: players, Bindings: bindings, Translations: translations}
}

// Handle decodes and executes one camera gallery request.
func (handler *Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	ctx := context.Background()
	switch packet.Header {
	case inconfiguration.Header:
		if err = inconfiguration.Decode(packet); err != nil {
			return err
		}
		settings, settingsErr := handler.Service.Configuration(ctx)
		if settingsErr != nil {
			return settingsErr
		}
		if settings.CreditsPrice > math.MaxInt32 || settings.PointsPrice > math.MaxInt32 || settings.PublishPointsPrice > math.MaxInt32 {
			return codec.ErrInvalidField
		}
		response, encodeErr := outinit.Encode(int32(settings.CreditsPrice), int32(settings.PointsPrice), int32(settings.PublishPointsPrice))
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	case inpurchase.Header:
		if err = inpurchase.Decode(packet); err != nil {
			return err
		}
		result, purchaseErr := handler.Service.Purchase(ctx, player.ID())
		if purchaseErr != nil {
			return handler.reject(connection, purchaseErr)
		}
		if err = furnitureprojection.Inventory(ctx, connection, nil, result.Item, result.Definition); err != nil {
			return err
		}
		response, encodeErr := outpurchase.Encode()
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	case inpublish.Header:
		if err = inpublish.Decode(packet); err != nil {
			return err
		}
		publication, remaining, publishErr := handler.Service.Publish(ctx, player.ID())
		seconds := cooldownSeconds(remaining)
		if publishErr != nil {
			if !errors.Is(publishErr, camerarecord.ErrCooldown) {
				if rejectErr := handler.reject(connection, publishErr); rejectErr != nil {
					return rejectErr
				}
			}
			response, encodeErr := outpublish.Encode(false, seconds)
			if encodeErr != nil {
				return encodeErr
			}
			return connection.Send(ctx, response)
		}
		response, encodeErr := outpublish.Encode(true, 0, outpublish.WithURL(publication.URL))
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// cooldownSeconds rounds a positive duration up to whole seconds.
func cooldownSeconds(duration time.Duration) int32 {
	if duration <= 0 {
		return 0
	}
	seconds := (duration + time.Second - 1) / time.Second
	if seconds > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(seconds)
}

// reject sends one localized camera purchase error.
func (handler *Handler) reject(connection netconn.Context, cause error) error {
	key := i18n.Key("camera.error.unavailable")
	if errors.Is(cause, camerarecord.ErrDisabled) {
		key = "camera.error.disabled"
	}
	if errors.Is(cause, camerarecord.ErrNoPermission) {
		key = "camera.error.no_permission"
	}
	if errors.Is(cause, camerarecord.ErrNoPendingCapture) {
		key = "camera.error.no_pending_capture"
	}
	if errors.Is(cause, camerarecord.ErrInsufficientCredits) {
		key = "camera.error.insufficient_credits"
	}
	if errors.Is(cause, camerarecord.ErrInsufficientPoints) {
		key = "camera.error.insufficient_points"
	}
	message := string(key)
	if handler.Translations != nil {
		message = handler.Translations.Default(key)
	}
	response, err := outalert.Encode(message)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// Register adds camera gallery headers to a connection registry.
func Register(registry *netconn.HandlerRegistry, handler *Handler) {
	for _, header := range []uint16{inconfiguration.Header, inpurchase.Header, inpublish.Header} {
		_ = registry.Register(header, handler.Handle)
	}
}
