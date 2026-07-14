// Package list encodes USER_EFFECTS snapshots.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_EFFECTS.
const Header uint16 = 340

// Effect describes one effect inventory record.
type Effect struct {
	// Type identifies the avatar effect.
	Type int32
	// SubType identifies the effect variant.
	SubType int32
	// Duration stores one charge duration in seconds.
	Duration int32
	// InactiveEffectsInInventory stores charges not currently active.
	InactiveEffectsInInventory int32
	// SecondsLeftIfActive stores remaining active duration.
	SecondsLeftIfActive int32
	// Permanent reports a non-expiring effect.
	Permanent bool
}

// Encode creates a USER_EFFECTS packet.
func Encode(effects []Effect) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(effects))))
	for _, effect := range effects {
		if err != nil {
			break
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField},
			codec.Int32(effect.Type), codec.Int32(effect.SubType), codec.Int32(effect.Duration), codec.Int32(effect.InactiveEffectsInInventory), codec.Int32(effect.SecondsLeftIfActive), codec.Bool(effect.Permanent))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
