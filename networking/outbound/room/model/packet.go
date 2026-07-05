// Package model contains the ROOM_MODEL outbound heightmap packet.
package model

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_MODEL packet identifier.
	Header uint16 = 1301
)

// Definition describes the ROOM_MODEL payload fields.
var Definition = codec.Definition{
	codec.Named("scale", codec.BooleanField),
	codec.Named("wallHeight", codec.Int32Field),
	codec.Named("heightmap", codec.StringField),
}

// Encode creates a ROOM_MODEL packet.
func Encode(scale bool, wallHeight int32, heightmap string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(scale), codec.Int32(wallHeight), codec.String(heightmap))
}
