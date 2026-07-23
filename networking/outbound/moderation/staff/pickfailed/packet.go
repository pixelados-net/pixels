// Package pickfailed contains the moderation pickfailed outbound packet.
package pickfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation pickfailed packet.
const Header uint16 = 3150

// Definition describes moderation pickfailed fields.
var Definition = codec.Definition{
	codec.Named("issueID", codec.Int32Field),
	codec.Named("state", codec.Int32Field),
	codec.Named("moderatorID", codec.Int32Field),
	codec.Named("moderatorName", codec.StringField),
	codec.Named("retry", codec.BooleanField),
	codec.Named("retryCount", codec.Int32Field),
}

// Encode creates a moderation pickfailed packet.
func Encode(issueID int32, state int32, moderatorID int32, moderatorName string, retry bool, retryCount int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(issueID), codec.Int32(state), codec.Int32(moderatorID), codec.String(moderatorName), codec.Bool(retry), codec.Int32(retryCount))
}
