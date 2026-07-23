// Package thumbnailstatus contains THUMBNAIL_STATUS.
package thumbnailstatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies THUMBNAIL_STATUS.
const Header uint16 = 3595

// Definition describes thumbnail render status.
var Definition = codec.Definition{codec.BooleanField, codec.BooleanField}

// Encode creates a thumbnail render result.
func Encode(ok bool, renderLimitHit bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(ok), codec.Bool(renderLimitHit))
}
