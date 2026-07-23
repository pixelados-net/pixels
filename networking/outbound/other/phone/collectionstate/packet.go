// Package collectionstate encodes the retired PHONE_COLLECTION_STATE packet.
package collectionstate

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PHONE_COLLECTION_STATE.
const Header uint16 = 2890

// Definition describes the three legacy state counters.
var Definition = codec.Definition{codec.Named("state", codec.Int32Field), codec.Named("attempts", codec.Int32Field), codec.Named("remaining", codec.Int32Field)}

// Encode creates one compatibility packet.
func Encode(state int32, attempts int32, remaining int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(state), codec.Int32(attempts), codec.Int32(remaining))
}
