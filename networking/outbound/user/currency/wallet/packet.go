// Package wallet contains the USER_CURRENCY outbound packet.
package wallet

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the USER_CURRENCY packet identifier.
	Header uint16 = 2018
)

var (
	// ErrAmountOutOfRange reports a balance that cannot fit the protocol field.
	ErrAmountOutOfRange = errors.New("currency amount exceeds int32 protocol range")
)

// Entry contains one seasonal currency balance.
type Entry struct {
	// Type identifies the protocol currency.
	Type int32

	// Amount stores the absolute balance.
	Amount int64
}

// Definition describes the USER_CURRENCY collection header.
var Definition = codec.Definition{codec.Named("currencyCount", codec.Int32Field)}

// EntryDefinition describes one USER_CURRENCY entry.
var EntryDefinition = codec.Definition{
	codec.Named("currencyType", codec.Int32Field),
	codec.Named("amount", codec.Int32Field),
}

// Encode creates a USER_CURRENCY packet.
func Encode(entries []Entry) (codec.Packet, error) {
	payload := make([]byte, 0, 4+(len(entries)*8))
	payload = binary.BigEndian.AppendUint32(payload, uint32(len(entries)))
	for _, entry := range entries {
		if entry.Amount < 0 || entry.Amount > math.MaxInt32 {
			return codec.Packet{}, ErrAmountOutOfRange
		}
		payload = binary.BigEndian.AppendUint32(payload, uint32(entry.Type))
		payload = binary.BigEndian.AppendUint32(payload, uint32(entry.Amount))
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
