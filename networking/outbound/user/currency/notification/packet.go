// Package notification contains the ACTIVITY_POINT_NOTIFICATION outbound packet.
package notification

import (
	"errors"
	"math"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the ACTIVITY_POINT_NOTIFICATION packet identifier.
	Header uint16 = 2275
)

var (
	// ErrAmountOutOfRange reports values that cannot fit protocol fields.
	ErrAmountOutOfRange = errors.New("currency notification amount exceeds int32 protocol range")
)

// Definition describes the ACTIVITY_POINT_NOTIFICATION payload fields.
var Definition = codec.Definition{
	codec.Named("amount", codec.Int32Field),
	codec.Named("amountChanged", codec.Int32Field),
	codec.Named("currencyType", codec.Int32Field),
}

// Encode creates an ACTIVITY_POINT_NOTIFICATION packet.
func Encode(amount int64, changed int64, currencyType int32) (codec.Packet, error) {
	if amount < 0 || amount > math.MaxInt32 || changed < math.MinInt32 || changed > math.MaxInt32 {
		return codec.Packet{}, ErrAmountOutOfRange
	}

	return codec.NewPacket(Header, Definition,
		codec.Int32(int32(amount)),
		codec.Int32(int32(changed)),
		codec.Int32(currencyType),
	)
}
