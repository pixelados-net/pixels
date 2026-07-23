// Package prize encodes GOTMYSTERYBOXPRIZEMESSAGE responses.
package prize

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GOTMYSTERYBOXPRIZEMESSAGE.
const Header uint16 = 3712

// Definition describes a mystery-box prize.
var Definition = codec.Definition{codec.Named("contentType", codec.StringField), codec.Named("classId", codec.Int32Field)}

// Encode creates one mystery-box prize response.
func Encode(contentType string, classID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(contentType), codec.Int32(classID))
}
