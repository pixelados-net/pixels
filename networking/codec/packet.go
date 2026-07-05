// Package codec encodes and decodes pixel-protocol frames and payloads.
package codec

// HeaderSize is the number of bytes in the packet header segment.
const HeaderSize = 2

// LengthSize is the number of bytes in the frame length prefix.
const LengthSize = 4

// FrameOverhead is the number of bytes outside the payload in an encoded frame.
const FrameOverhead = LengthSize + HeaderSize

// Packet is a decoded pixel-protocol packet.
type Packet struct {
	// Header stores the packet identifier.
	Header uint16
	// Payload stores the encoded packet body.
	Payload []byte
}

// NewPacket creates a packet by encoding values with a payload definition.
func NewPacket(header uint16, definition Definition, values ...Value) (Packet, error) {
	payload, err := AppendPayload(nil, definition, values...)
	if err != nil {
		return Packet{}, err
	}

	return Packet{Header: header, Payload: payload}, nil
}

// DecodePacket decodes packet payload values with a payload definition.
func DecodePacket(packet Packet, definition Definition) ([]Value, []byte, error) {
	return DecodePayload(nil, definition, packet.Payload)
}

// DecodePacketExact decodes packet payload values and rejects remaining bytes.
func DecodePacketExact(packet Packet, definition Definition) ([]Value, error) {
	values, rest, err := DecodePacket(packet, definition)
	if err != nil {
		return nil, err
	}

	if len(rest) != 0 {
		return nil, ErrUnexpectedPayload
	}

	return values, nil
}

// Size returns the packet size covered by the frame length prefix.
func (packet Packet) Size() int {
	return HeaderSize + len(packet.Payload)
}
