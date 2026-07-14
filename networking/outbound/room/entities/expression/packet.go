// Package expression encodes UNIT_EXPRESSION broadcasts.
package expression

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_EXPRESSION.
const Header uint16 = 1631

// Definition describes UNIT_EXPRESSION fields.
var Definition = codec.Definition{codec.Named("roomIndex", codec.Int32Field), codec.Named("expressionId", codec.Int32Field)}

// Encode creates a UNIT_EXPRESSION packet.
func Encode(roomIndex int64, expressionID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(roomIndex)), codec.Int32(expressionID))
}
