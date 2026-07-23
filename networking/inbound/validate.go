// Package inbound provides shared inbound packet validation.
package inbound

import "github.com/niflaot/pixels/networking/codec"

// ValidateHeader rejects packets delivered under an unexpected header.
func ValidateHeader(packet codec.Packet, expected uint16) error {
	if packet.Header != expected {
		return codec.ErrUnexpectedHeader
	}
	return nil
}
