// Package whisper contains the UNIT_CHAT_WHISPER outbound packet.
package whisper

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies UNIT_CHAT_WHISPER.
	Header uint16 = 2704
)

// Definition describes a rendered room whisper.
var Definition = codec.Definition{
	codec.Named("unitId", codec.Int32Field), codec.Named("message", codec.StringField),
	codec.Named("gesture", codec.Int32Field), codec.Named("styleId", codec.Int32Field),
	codec.Named("urlCount", codec.Int32Field), codec.Named("messageLength", codec.Int32Field),
}

// Encode creates a UNIT_CHAT_WHISPER packet without embedded URLs.
func Encode(unitID int32, message string, gesture int32, styleID int32, messageLength int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(unitID), codec.String(message), codec.Int32(gesture), codec.Int32(styleID), codec.Int32(0), codec.Int32(messageLength))
}
