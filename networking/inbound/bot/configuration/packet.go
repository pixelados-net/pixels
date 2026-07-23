// Package configuration decodes BOT_CONFIGURATION requests.
package configuration

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_CONFIGURATION.
const Header uint16 = 1986

// Definition describes BOT_CONFIGURATION fields.
var Definition = codec.Definition{codec.Named("botId", codec.Int32Field), codec.Named("skillId", codec.Int32Field)}

// Payload contains one requested bot skill configuration.
type Payload struct {
	// BotID identifies the placed bot.
	BotID int64
	// SkillID identifies requested configuration data.
	SkillID int32
}

// Decode decodes BOT_CONFIGURATION.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	id := int64(values[0].Int32)
	if id < 0 {
		id = -id
	}
	return Payload{BotID: id, SkillID: values[1].Int32}, nil
}
