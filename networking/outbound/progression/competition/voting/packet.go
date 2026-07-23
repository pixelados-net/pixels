// Package voting encodes COMPETITION_VOTING_INFO responses.
package voting

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMPETITION_VOTING_INFO.
const Header uint16 = 3506

// Encode creates one competition voting status.
func Encode(goalID int32, goalCode string, result int32, votesRemaining int32) (codec.Packet, error) {
	definition := codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field}
	return codec.NewPacket(Header, definition, codec.Int32(goalID), codec.String(goalCode), codec.Int32(result), codec.Int32(votesRemaining))
}
