// Package privatechat contains MESSENGER_CHAT.
package privatechat

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_CHAT.
const Header uint16 = 1587

// Option configures optional MESSENGER_CHAT fields.
type Option func(*options)

// options stores optional MESSENGER_CHAT values.
type options struct {
	// extraData stores optional client metadata.
	extraData *string
}

// WithExtraData includes Nitro's optional extra-data field.
func WithExtraData(value string) Option {
	return func(configured *options) {
		configured.extraData = &value
	}
}

// Encode creates a live private message.
func Encode(senderID int64, message string, secondsSinceSent int32, optionFunctions ...Option) (codec.Packet, error) {
	configured := options{}
	for _, option := range optionFunctions {
		option(&configured)
	}
	definition := codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field}
	values := []codec.Value{codec.Int32(int32(senderID)), codec.String(message), codec.Int32(secondsSinceSent)}
	if configured.extraData != nil {
		definition = append(definition, codec.StringField)
		values = append(values, codec.String(*configured.extraData))
	}
	return codec.NewPacket(Header, definition, values...)
}
