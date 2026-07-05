// Package machine contains the SECURITY_MACHINE inbound packet.
package machine

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the SECURITY_MACHINE packet identifier.
	Header uint16 = 2490
)

// Payload contains the unpacked SECURITY_MACHINE fields.
type Payload struct {
	// MachineID is the machineId protocol field.
	MachineID string
	// Fingerprint is the fingerprint protocol field.
	Fingerprint string
	// Capabilities is the capabilities protocol field.
	Capabilities string
}

// Definition describes the SECURITY_MACHINE payload fields.
var Definition = codec.Definition{
	codec.Named("machineId", codec.StringField),
	codec.Named("fingerprint", codec.StringField),
	codec.Named("capabilities", codec.StringField),
}

// Decode unpacks a SECURITY_MACHINE packet payload.
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
	payload.MachineID = values[0].String
	payload.Fingerprint = values[1].String
	payload.Capabilities = values[2].String

	return payload
}
