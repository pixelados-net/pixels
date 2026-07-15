// Package number encodes a transient number above one room unit.
package number

import "github.com/niflaot/pixels/networking/codec"

// Header is the UNIT_NUMBER identifier.
const Header uint16 = 2324

// Encode creates a UNIT_NUMBER packet.
func Encode(unitID int64, value int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(unitID)), codec.Int32(value))
}
