// Package status contains ACCOUNT_SAFETY_LOCK_STATUS_CHANGE.
package status

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ACCOUNT_SAFETY_LOCK_STATUS_CHANGE.
	Header uint16 = 1243
	// Locked is Nitro's safety-lock value.
	Locked int32 = 0
	// Unlocked is Nitro's normal account value.
	Unlocked int32 = 1
)

// Definition describes the safety-lock status.
var Definition = codec.Definition{codec.Named("status", codec.Int32Field)}

// Encode creates an ACCOUNT_SAFETY_LOCK_STATUS_CHANGE packet.
func Encode(status int32) (codec.Packet, error) {
	if status != Locked && status != Unlocked {
		return codec.Packet{}, codec.ErrInvalidField
	}
	return codec.NewPacket(Header, Definition, codec.Int32(status))
}
