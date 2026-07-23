// Package winners encodes WEEKLY_GAME_REWARD_WINNERS responses.
package winners

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WEEKLY_GAME_REWARD_WINNERS.
const Header uint16 = 3097

// Winner describes one ranked weekly winner.
type Winner struct {
	// Name stores the player name.
	Name string
	// Figure stores the avatar figure.
	Figure string
	// Gender stores the avatar gender.
	Gender string
	// Rank stores the one-based rank.
	Rank int32
	// Score stores the winning score.
	Score int32
}

// Encode creates one weekly winners response.
func Encode(gameTypeID int32, winners []Winner) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(gameTypeID), codec.Int32(int32(len(winners))))
	for _, winner := range winners {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field}, codec.String(winner.Name), codec.String(winner.Figure), codec.String(winner.Gender), codec.Int32(winner.Rank), codec.Int32(winner.Score))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
