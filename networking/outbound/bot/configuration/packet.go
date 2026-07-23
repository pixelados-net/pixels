// Package configuration encodes BOT_COMMAND_CONFIGURATION.
package configuration

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_COMMAND_CONFIGURATION.
const Header uint16 = 1618

// Definition describes BOT_COMMAND_CONFIGURATION fields.
var Definition = codec.Definition{codec.Named("botId", codec.Int32Field), codec.Named("skillId", codec.Int32Field), codec.Named("data", codec.StringField)}

// Encode creates BOT_COMMAND_CONFIGURATION.
func Encode(botID int64, skillID int32, data string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(-botID)), codec.Int32(skillID), codec.String(data))
}
