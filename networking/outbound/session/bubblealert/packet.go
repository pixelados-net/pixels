// Package bubblealert contains the BUBBLE_ALERT outbound packet.
package bubblealert

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the BUBBLE_ALERT packet identifier.
	Header uint16 = 1992

	// keyCount is the number of key/value pairs sent, matching Arcturus's single "message" key usage.
	keyCount int32 = 1
)

// Definition describes the BUBBLE_ALERT payload fields.
var Definition = codec.Definition{
	codec.Named("errorKey", codec.StringField),
	codec.Named("keyCount", codec.Int32Field),
	codec.Named("key", codec.StringField),
	codec.Named("value", codec.StringField),
}

// Encode creates a BUBBLE_ALERT packet carrying a single "message" key, matching the shape every
// real Arcturus BubbleAlertComposer call site actually uses.
func Encode(errorKey string, message string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.String(errorKey),
		codec.Int32(keyCount),
		codec.String("message"),
		codec.String(message),
	)
}
