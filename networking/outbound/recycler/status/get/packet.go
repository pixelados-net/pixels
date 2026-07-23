// Package get contains the RECYCLER_STATUS outbound packet.
package get

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies RECYCLER_STATUS.
	Header uint16 = 3433
	// Enabled identifies an available recycler.
	Enabled int32 = 1
	// Disabled identifies a closed recycler.
	Disabled int32 = 2
)

// Encode creates one recycler availability packet.
func Encode(status int32, timeoutSeconds int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(status), codec.Int32(timeoutSeconds))
}
