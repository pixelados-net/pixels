// Package privateerror contains MESSENGER_INSTANCE_MESSAGE_ERROR.
package privateerror

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_INSTANCE_MESSAGE_ERROR.
const Header uint16 = 3359

// Encode creates a private-message failure.
func Encode(errorCode int32, playerID int64, message string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField}, codec.Int32(errorCode), codec.Int32(int32(playerID)), codec.String(message))
}
