// Package requesterror contains MESSENGER_REQUEST_ERROR.
package requesterror

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_REQUEST_ERROR.
const Header uint16 = 892

// Encode creates MESSENGER_REQUEST_ERROR.
func Encode(clientMessageID int32, errorCode int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(clientMessageID), codec.Int32(errorCode))
}
