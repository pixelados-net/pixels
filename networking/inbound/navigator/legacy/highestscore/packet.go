// Package highestscore decodes ROOMS_WITH_HIGHEST_SCORE_SEARCH requests.
package highestscore

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOMS_WITH_HIGHEST_SCORE_SEARCH.
const Header uint16 = 2939

// Definition describes the pageIndex filter carried by the legacy composer.
var Definition = codec.Definition{codec.Named("pageIndex", codec.Int32Field)}

// Decode returns the requested pageIndex value.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
