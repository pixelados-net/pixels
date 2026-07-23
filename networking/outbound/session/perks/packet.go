// Package perks contains the USER_PERKS outbound packet.
package perks

import (
	"encoding/binary"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the USER_PERKS packet identifier.
	Header uint16 = 2586
)

// Entry contains one Nitro perk allowance.
type Entry struct {
	// Code identifies the client-known perk.
	Code string
	// Error stores the client requirement message for denied perks.
	Error string
	// Allowed reports whether the perk is enabled.
	Allowed bool
}

// Definition describes the USER_PERKS collection header.
var Definition = codec.Definition{codec.Named("perkCount", codec.Int32Field)}

// EntryDefinition describes one USER_PERKS entry.
var EntryDefinition = codec.Definition{
	codec.Named("code", codec.StringField),
	codec.Named("error", codec.StringField),
	codec.Named("allowed", codec.BooleanField),
}

// Encode creates a USER_PERKS packet.
func Encode(entries []Entry) (codec.Packet, error) {
	size := 4
	for _, entry := range entries {
		size += 5 + len(entry.Code) + len(entry.Error)
	}
	payload := make([]byte, 0, size)
	payload = binary.BigEndian.AppendUint32(payload, uint32(len(entries)))
	for _, entry := range entries {
		var err error
		payload, err = codec.AppendPayload(payload, EntryDefinition,
			codec.String(entry.Code),
			codec.String(entry.Error),
			codec.Bool(entry.Allowed),
		)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
