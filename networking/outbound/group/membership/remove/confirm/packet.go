// Package removeconfirm contains one Nitro social-group outbound packet.
package removeconfirm

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 1876

// Encode creates the complete packet.
func Encode(playerID int64, furnitureCount int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(playerID)), codec.Int32(int32(furnitureCount)))
}
