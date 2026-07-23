// Package guildbasesearch decodes GUILD_BASE_SEARCH requests.
package guildbasesearch

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GUILD_BASE_SEARCH.
const Header uint16 = 2930

// Definition describes the groupId filter carried by the legacy composer.
var Definition = codec.Definition{codec.Named("groupId", codec.Int32Field)}

// Decode returns the requested groupId value.
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
