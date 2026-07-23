// Package list encodes GAME_CENTER_GAME_LIST responses.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_GAME_LIST.
const Header uint16 = 222

// Game describes one externally launchable game.
type Game struct {
	// ID identifies the game type.
	ID int32
	// Name stores the client-visible localization key or name.
	Name string
	// BackgroundColor stores a six-digit hexadecimal RGB value.
	BackgroundColor string
	// TextColor stores a six-digit hexadecimal RGB value.
	TextColor string
	// AssetURL stores the lobby artwork URL.
	AssetURL string
	// SupportURL stores the support page URL.
	SupportURL string
}

// Encode creates one ordered game list.
func Encode(games []Game) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(games))))
	for _, game := range games {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.StringField}, codec.Int32(game.ID), codec.String(game.Name), codec.String(game.BackgroundColor), codec.String(game.TextColor), codec.String(game.AssetURL), codec.String(game.SupportURL))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
