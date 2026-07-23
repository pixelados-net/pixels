package report

import (
	"context"

	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inreport "github.com/niflaot/pixels/networking/inbound/camera/report"
	outresult "github.com/niflaot/pixels/networking/outbound/moderation/cfh/result"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler adapts photo report packets to moderation intake.
type Handler struct {
	// Service validates photo evidence.
	Service *Service
	// Players resolves live reporters.
	Players *playerlive.Registry
	// Bindings resolves authenticated connections.
	Bindings *binding.Registry
	// Translations resolves hotel-facing messages.
	Translations i18n.Translator
}

// NewHandler creates a camera photo report handler.
func NewHandler(service *Service, players *playerlive.Registry, bindings *binding.Registry, translations i18n.Translator) *Handler {
	return &Handler{Service: service, Players: players, Bindings: bindings, Translations: translations}
}

// Handle validates and submits one photo report.
func (handler *Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	payload, err := inreport.Decode(packet)
	if err != nil {
		return err
	}
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	_, reportErr := handler.Service.Submit(context.Background(), player.ID(), int64(payload.ItemID), int64(payload.RoomID), int64(payload.TopicID), payload.ExtraDataID)
	code, message := int32(0), handler.text("moderation.report.received")
	if reportErr != nil {
		code, message = 3, handler.text("moderation.report.failed")
	}
	response, err := outresult.Encode(code, message)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// text resolves one translation key with a stable fallback.
func (handler *Handler) text(key i18n.Key) string {
	if handler.Translations == nil {
		return string(key)
	}
	return handler.Translations.Default(key)
}

// Register adds the photo report header to a connection registry.
func Register(registry *netconn.HandlerRegistry, handler *Handler) {
	_ = registry.Register(inreport.Header, handler.Handle)
}
