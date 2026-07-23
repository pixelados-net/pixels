// Package whisper contains the UNIT_CHAT_WHISPER inbound packet.
package whisper

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies UNIT_CHAT_WHISPER.
	Header uint16 = 1543
)

// Definition describes Nitro's combined recipient-message field and bubble style.
var Definition = codec.Definition{codec.Named("recipientAndMessage", codec.StringField), codec.Named("styleId", codec.Int32Field)}

// Payload contains a decoded whisper request.
type Payload struct {
	// RecipientAndMessage stores the recipient followed by one space and the message.
	RecipientAndMessage string
	// StyleID stores the requested bubble style.
	StyleID int32
}

// Decode decodes a UNIT_CHAT_WHISPER packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{RecipientAndMessage: values[0].String, StyleID: values[1].Int32}, nil
}
