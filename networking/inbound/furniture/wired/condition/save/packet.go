// Package save decodes the WIRED_CONDITION_SAVE inbound packet.
package save

import (
	"github.com/niflaot/pixels/networking/codec"
	wiredcommon "github.com/niflaot/pixels/networking/inbound/furniture/wired/common"
)

// Header is the WIRED_CONDITION_SAVE packet identifier.
const Header uint16 = 3203

// Payload stores one condition configuration request.
type Payload = wiredcommon.Settings

// Decode decodes one WIRED_CONDITION_SAVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	return wiredcommon.DecodeSettings(packet.Payload, false)
}
