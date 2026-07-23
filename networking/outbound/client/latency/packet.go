// Package latency contains the CLIENT_LATENCY outbound packet.
package latency

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_LATENCY packet identifier.
	Header uint16 = 10
)

// Definition describes the CLIENT_LATENCY payload fields.
var Definition = codec.Definition{
	codec.Named("requestId", codec.Int32Field),
}

// Encode creates a CLIENT_LATENCY packet.
func Encode(requestID int32) (codec.Packet, error) {
	values := make([]codec.Value, 0, 1)
	values = append(values, codec.Int32(requestID))

	return codec.NewPacket(Header, Definition, values...)
}
