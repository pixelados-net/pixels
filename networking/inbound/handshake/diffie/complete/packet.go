// Package complete contains the HANDSHAKE_COMPLETE_DIFFIE inbound packet.
package complete

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HANDSHAKE_COMPLETE_DIFFIE packet identifier.
	Header uint16 = 773
)

// Payload contains the unpacked HANDSHAKE_COMPLETE_DIFFIE fields.
type Payload struct {
	// EncryptedPublicKey is the encryptedPublicKey protocol field.
	EncryptedPublicKey string
}

// Definition describes the HANDSHAKE_COMPLETE_DIFFIE payload fields.
var Definition = codec.Definition{
	codec.Named("encryptedPublicKey", codec.StringField),
}

// Decode unpacks a HANDSHAKE_COMPLETE_DIFFIE packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return payloadFromValues(values), nil
}

// payloadFromValues returns a typed payload from decoded values.
func payloadFromValues(values []codec.Value) Payload {
	var payload Payload
	payload.EncryptedPublicKey = values[0].String

	return payload
}
