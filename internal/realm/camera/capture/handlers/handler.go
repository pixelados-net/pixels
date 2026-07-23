// Package handlers adapts camera render packets to capture workflows.
package handlers

import (
	"context"
	"errors"

	cameracapture "github.com/niflaot/pixels/internal/realm/camera/capture"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inthumbnail "github.com/niflaot/pixels/networking/inbound/camera/thumbnail"
	inrender "github.com/niflaot/pixels/networking/inbound/session/render"
	outstorage "github.com/niflaot/pixels/networking/outbound/camera/storageurl"
	outthumbnail "github.com/niflaot/pixels/networking/outbound/camera/thumbnailstatus"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler handles camera photo and thumbnail uploads.
type Handler struct {
	// Service executes capture workflows.
	Service *cameracapture.Service
	// Players resolves live players.
	Players *playerlive.Registry
	// Bindings resolves authenticated connections.
	Bindings *binding.Registry
	// Translations resolves hotel-facing errors.
	Translations i18n.Translator
}

// New creates a grouped camera capture packet handler.
func New(service *cameracapture.Service, players *playerlive.Registry, bindings *binding.Registry, translations i18n.Translator) *Handler {
	return &Handler{Service: service, Players: players, Bindings: bindings, Translations: translations}
}

// Handle decodes and executes one camera upload.
func (handler *Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return handler.reject(connection, camerarecord.ErrNoPermission)
	}
	ctx := context.Background()
	switch packet.Header {
	case inrender.Header:
		payload, decodeErr := inrender.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		capture, captureErr := handler.Service.Photo(ctx, player.ID(), roomID, payload.PNG)
		if captureErr != nil {
			return handler.reject(connection, captureErr)
		}
		response, encodeErr := outstorage.Encode(capture.URL)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	case inthumbnail.Header:
		payload, decodeErr := inthumbnail.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		_, captureErr := handler.Service.Thumbnail(ctx, player.ID(), roomID, payload.PNG)
		response, encodeErr := outthumbnail.Encode(captureErr == nil, false)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// reject sends one localized camera capture error.
func (handler *Handler) reject(connection netconn.Context, cause error) error {
	key := i18n.Key("camera.error.unavailable")
	if errors.Is(cause, camerarecord.ErrDisabled) {
		key = "camera.error.disabled"
	}
	if errors.Is(cause, camerarecord.ErrNoPermission) {
		key = "camera.error.no_permission"
	}
	if errors.Is(cause, camerarecord.ErrTooLarge) {
		key = "camera.error.render_too_large"
	}
	if errors.Is(cause, camerarecord.ErrCooldown) {
		key = "camera.error.capture_cooldown"
	}
	if errors.Is(cause, camerarecord.ErrNotRoomOwner) {
		key = "camera.error.not_room_owner"
	}
	if errors.Is(cause, camerarecord.ErrInvalidPhoto) {
		key = "camera.error.invalid_photo"
	}
	message := string(key)
	if handler.Translations != nil {
		message = handler.Translations.Default(key)
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), packet)
}

// Register adds camera capture headers to a connection registry.
func Register(registry *netconn.HandlerRegistry, handler *Handler) {
	for _, header := range []uint16{inrender.Header, inthumbnail.Header} {
		_ = registry.Register(header, handler.Handle)
	}
}
