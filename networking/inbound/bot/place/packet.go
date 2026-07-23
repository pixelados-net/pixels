// Package place decodes BOT_PLACE requests.
package place

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_PLACE.
const Header uint16 = 1592

// Definition describes BOT_PLACE fields.
var Definition = codec.Definition{codec.Named("botId", codec.Int32Field), codec.Named("x", codec.Int32Field), codec.Named("y", codec.Int32Field)}

// Payload contains one placement request.
type Payload struct {
	// BotID identifies the inventory bot.
	BotID int64
	// X stores the requested tile coordinate.
	X int
	// Y stores the requested tile coordinate.
	Y int
}

// Decode decodes BOT_PLACE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{BotID: int64(values[0].Int32), X: int(values[1].Int32), Y: int(values[2].Int32)}, nil
}
