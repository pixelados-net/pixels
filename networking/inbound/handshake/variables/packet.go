// Package variables contains the CLIENT_VARIABLES inbound packet.
package variables

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_VARIABLES packet identifier.
	Header uint16 = 1053
)

// Payload contains the unpacked CLIENT_VARIABLES fields.
type Payload struct {
	// ClientID is the clientId protocol field.
	ClientID int32
	// ClientURL is the clientUrl protocol field.
	ClientURL string
	// ExternalVariablesURL is the externalVariablesUrl protocol field.
	ExternalVariablesURL string
}

// Definition describes the CLIENT_VARIABLES payload fields.
var Definition = codec.Definition{
	codec.Named("clientId", codec.Int32Field),
	codec.Named("clientUrl", codec.StringField),
	codec.Named("externalVariablesUrl", codec.StringField),
}

// Decode unpacks a CLIENT_VARIABLES packet payload.
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
	payload.ClientID = values[0].Int32
	payload.ClientURL = values[1].String
	payload.ExternalVariablesURL = values[2].String

	return payload
}
