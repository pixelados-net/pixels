// Package merge contains the HOTEL_MERGE_NAME_CHANGE outbound packet.
package merge

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HOTEL_MERGE_NAME_CHANGE packet identifier.
	Header uint16 = 1663
)

// Definition describes the HOTEL_MERGE_NAME_CHANGE payload fields.
var Definition = codec.Definition{}

// Encode creates a HOTEL_MERGE_NAME_CHANGE packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
