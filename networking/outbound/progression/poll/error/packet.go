// Package error encodes POLL_ERROR responses.
package error

import "github.com/niflaot/pixels/networking/codec"

// Header identifies POLL_ERROR.
const Header uint16 = 662

// Encode creates one header-only POLL_ERROR response.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{})
}
