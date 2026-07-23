// Package ticket contains the SECURITY_TICKET inbound packet.
package ticket

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the SECURITY_TICKET packet identifier.
	Header uint16 = 2419
)

// Payload contains the unpacked SECURITY_TICKET fields.
type Payload struct {
	// Ticket is the ticket protocol field.
	Ticket string
	// Timestamp is the timestamp protocol field.
	Timestamp *int32
}

// Definition describes the SECURITY_TICKET payload fields.
var Definition = codec.Definition{
	codec.Named("ticket", codec.StringField),
	codec.Optional(codec.Named("timestamp", codec.Int32Field)),
}

// Decode unpacks a SECURITY_TICKET packet payload.
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
	payload.Ticket = values[0].String
	if len(values) > 1 {
		value := values[1].Int32
		payload.Timestamp = &value
	}

	return payload
}
