package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/networking/codec"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// kind returns the normalized notification kind.
func (request NotifyRequest) kind() string {
	if request.Kind == "" {
		return KindBubble
	}

	return request.Kind
}

// bubbleKey returns the requested bubble key.
func (request NotifyRequest) bubbleKey() string {
	if request.BubbleKey == "" {
		return defaultBubbleKey
	}

	return request.BubbleKey
}

// notificationPacket creates the requested notification packet.
func notificationPacket(request NotifyRequest, translations i18n.Translator) (codec.Packet, error) {
	message := localizedMessage(request, translations)
	switch request.kind() {
	case KindBubble:
		return outbubble.Encode(request.bubbleKey(), message, outbubble.WithDisplayBubble())
	case KindAlert:
		return outalert.Encode(message)
	default:
		return codec.Packet{}, fiber.NewError(fiber.StatusBadRequest, "unsupported notification kind")
	}
}

// localizedMessage resolves the request message.
func localizedMessage(request NotifyRequest, translations i18n.Translator) string {
	key := i18n.Key(request.Key)
	params := i18n.Params(request.Params)
	if translations == nil {
		return string(key)
	}
	if request.Locale != "" {
		return translations.T(i18n.Locale(request.Locale), key, params)
	}

	return translations.Default(key, params)
}
