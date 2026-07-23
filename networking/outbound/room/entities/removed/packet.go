// Package removed contains the UNIT_REMOVE outbound packet.
package removed

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the UNIT_REMOVE packet identifier.
	Header uint16 = 2661
)

// Definition describes the UNIT_REMOVE payload fields.
var Definition = codec.Definition{codec.Named("roomIndex", codec.StringField)}

// Encode creates a UNIT_REMOVE packet.
func Encode(roomIndex int64) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(strconv.FormatInt(roomIndex, 10)))
}
