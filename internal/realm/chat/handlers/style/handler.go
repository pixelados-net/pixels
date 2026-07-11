// Package style adapts USER_SETTINGS_CHAT_STYLE to a style command.
package style

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/realm/chat/bubble"
	stylecmd "github.com/niflaot/pixels/internal/realm/chat/commands/style"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	instyle "github.com/niflaot/pixels/networking/inbound/chat/style"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// New creates a USER_SETTINGS_CHAT_STYLE packet handler.
func New(handler stylecmd.Handler, translations i18n.Translator, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		styleID, err := instyle.Decode(packet)
		if err != nil {
			return err
		}
		err = dispatcher.Dispatch(context.Background(), command.Envelope[stylecmd.Command]{
			Command:  stylecmd.Command{Handler: connection, StyleID: styleID},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
		if !errors.Is(err, bubble.ErrBubbleLocked) && !errors.Is(err, bubble.ErrInvalidBubble) {
			return err
		}
		message := "chat.error.bubble_locked"
		if translations != nil {
			message = translations.Default(i18n.Key("chat.error.bubble_locked"))
		}
		alert, encodeErr := outalert.Encode(message)
		if encodeErr != nil {
			return encodeErr
		}

		return connection.Send(context.Background(), alert)
	}
}

// Register adds USER_SETTINGS_CHAT_STYLE to a handler registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(instyle.Header, handler)
}
