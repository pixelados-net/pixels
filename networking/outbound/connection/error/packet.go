// Package error contains the CONNECTION_ERROR outbound packet.
package error

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CONNECTION_ERROR packet identifier.
	Header uint16 = 1004
)

// Definition describes the CONNECTION_ERROR payload fields.
var Definition = codec.Definition{
	codec.Named("messageId", codec.Int32Field),
	codec.Named("errorCode", codec.Int32Field),
	codec.Named("timestamp", codec.StringField),
}

// Encode creates a CONNECTION_ERROR packet.
func Encode(messageID int32, errorCode int32, timestamp string) (codec.Packet, error) {
	values := make([]codec.Value, 0, 3)
	values = append(values, codec.Int32(messageID))
	values = append(values, codec.Int32(errorCode))
	values = append(values, codec.String(timestamp))

	return codec.NewPacket(Header, Definition, values...)
}
