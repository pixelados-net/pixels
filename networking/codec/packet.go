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
	Header  uint16
	Payload []byte
}

// Size returns the packet size covered by the frame length prefix.
func (packet Packet) Size() int {
	return HeaderSize + len(packet.Payload)
}
