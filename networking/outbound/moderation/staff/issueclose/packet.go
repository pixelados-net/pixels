// Package issueclose encodes ISSUE_CLOSE_NOTIFICATION.
package issueclose

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ISSUE_CLOSE_NOTIFICATION.
const Header uint16 = 934

// Definition describes the close reason and localized message.
var Definition = codec.Definition{codec.Named("closeReason", codec.Int32Field), codec.Named("messageText", codec.StringField)}

// Encode creates one issue-close notification.
func Encode(closeReason int32, messageText string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(closeReason), codec.String(messageText))
}
