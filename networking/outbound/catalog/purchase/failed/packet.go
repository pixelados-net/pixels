// Package failed contains the PURCHASE_ERROR outbound packet.
package failed

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PURCHASE_ERROR packet identifier.
	Header uint16 = 1404
	// CodeServer identifies a server-side purchase failure.
	CodeServer int32 = 0
	// CodeAlreadyOwned identifies a unique product already owned by the buyer.
	CodeAlreadyOwned int32 = 1
	// CodeRoomLimit identifies a room bundle blocked by the ownership limit.
	CodeRoomLimit int32 = 2
)

// Definition describes the PURCHASE_ERROR payload fields.
var Definition = codec.Definition{codec.Named("code", codec.Int32Field)}

// Encode creates a PURCHASE_ERROR packet.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code))
}
