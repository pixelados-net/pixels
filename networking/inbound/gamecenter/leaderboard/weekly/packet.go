// Package weekly decodes GAME2GETWEEKLYLEADERBOARD requests.
package weekly

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GAME2GETWEEKLYLEADERBOARD.
const Header uint16 = 2565

// Payload describes one weekly leaderboard request.
type Payload struct {
	// GameTypeID stores the protocol gametypeid value.
	GameTypeID int32
	// Year stores the protocol year value.
	Year int32
	// Week stores the protocol week value.
	Week int32
	// Offset stores the protocol offset value.
	Offset int32
	// Limit stores the protocol limit value.
	Limit int32
	// Unknown stores the protocol unknown value.
	Unknown int32
}

// Decode returns one validated weekly leaderboard request.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GameTypeID: values[0].Int32, Year: values[1].Int32, Week: values[2].Int32, Offset: values[3].Int32, Limit: values[4].Int32, Unknown: values[5].Int32}, nil
}
