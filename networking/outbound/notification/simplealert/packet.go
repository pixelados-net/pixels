// Package simplealert encodes NOTIFICATION_SIMPLE_ALERT.
package simplealert

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NOTIFICATION_SIMPLE_ALERT.
const Header uint16 = 5100

// Definition describes the localized body and title.
var Definition = codec.Definition{codec.Named("message", codec.StringField), codec.Named("title", codec.StringField)}

// Encode creates one titled alert.
func Encode(message string, title string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(message), codec.String(title))
}
