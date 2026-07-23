// Package skillsave decodes BOT_SKILL_SAVE requests.
package skillsave

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_SKILL_SAVE.
const Header uint16 = 2624

// Definition describes BOT_SKILL_SAVE fields.
var Definition = codec.Definition{codec.Named("botId", codec.Int32Field), codec.Named("skillId", codec.Int32Field), codec.Named("data", codec.StringField)}

// Payload contains one skill mutation.
type Payload struct {
	// BotID identifies the placed bot.
	BotID int64
	// SkillID identifies the mutation.
	SkillID int32
	// Data stores skill-specific free-form data.
	Data string
}

// Decode decodes BOT_SKILL_SAVE.
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
	return Payload{BotID: id, SkillID: values[1].Int32, Data: values[2].String}, nil
}
