// Package save decodes the WIRED_ACTION_SAVE inbound packet.
package save

import (
	"github.com/niflaot/pixels/networking/codec"
	wiredcommon "github.com/niflaot/pixels/networking/inbound/furniture/wired/common"
)

// Header is the WIRED_ACTION_SAVE packet identifier.
const Header uint16 = 2281

// Payload stores one action configuration request.
type Payload = wiredcommon.Settings

// Decode decodes one WIRED_ACTION_SAVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	return wiredcommon.DecodeSettings(packet.Payload, true)
}
