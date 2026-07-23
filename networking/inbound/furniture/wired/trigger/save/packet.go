// Package save decodes the WIRED_TRIGGER_SAVE inbound packet.
package save

import (
	"github.com/niflaot/pixels/networking/codec"
	wiredcommon "github.com/niflaot/pixels/networking/inbound/furniture/wired/common"
)

// Header is the WIRED_TRIGGER_SAVE packet identifier.
const Header uint16 = 1520

// Payload stores one trigger configuration request.
type Payload = wiredcommon.Settings

// Decode decodes one WIRED_TRIGGER_SAVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	return wiredcommon.DecodeSettings(packet.Payload, false)
}
