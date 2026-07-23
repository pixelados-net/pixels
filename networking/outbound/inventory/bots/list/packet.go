// Package list encodes USER_BOTS inventory responses.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BOTS.
const Header uint16 = 3086

// Bot stores one inventory bot record.
type Bot struct {
	// ID identifies the bot.
	ID int64
	// Name stores its visible name.
	Name string
	// Motto stores its visible motto.
	Motto string
	// Gender stores its one-letter gender.
	Gender string
	// Figure stores its Nitro figure.
	Figure string
}

// Encode creates USER_BOTS.
func Encode(bots []Bot) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(bots))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, bot := range bots {
		payload, err = appendBot(payload, bot)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendBot appends one inventory bot record.
func appendBot(payload []byte, bot Bot) ([]byte, error) {
	definition := codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField}
	return codec.AppendPayload(payload, definition, codec.Int32(int32(bot.ID)), codec.String(bot.Name), codec.String(bot.Motto), codec.String(bot.Gender), codec.String(bot.Figure))
}
