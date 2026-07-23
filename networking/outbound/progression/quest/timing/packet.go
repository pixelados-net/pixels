// Package timing encodes CURRENT_TIMING_CODE responses.
package timing

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CURRENT_TIMING_CODE.
const Header uint16 = 1745

// Encode creates the renderer-compatible two-string timing response.
func Encode(schedulingCode string, timingCode string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.StringField}, codec.String(schedulingCode), codec.String(timingCode))
}
