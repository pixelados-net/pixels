// Package info contains the USER_INFO outbound packet.
package info

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the USER_INFO packet identifier.
	Header uint16 = 2725
)

// Definition describes the USER_INFO payload fields.
var Definition = codec.Definition{
	codec.Named("userId", codec.Int32Field),
	codec.Named("username", codec.StringField),
	codec.Named("figure", codec.StringField),
	codec.Named("gender", codec.StringField),
	codec.Named("motto", codec.StringField),
	codec.Named("realName", codec.StringField),
	codec.Named("directMail", codec.BooleanField),
	codec.Named("respectsReceived", codec.Int32Field),
	codec.Named("respectsRemaining", codec.Int32Field),
	codec.Named("respectsPetRemaining", codec.Int32Field),
	codec.Named("streamPublishingAllowed", codec.BooleanField),
	codec.Named("lastAccessDate", codec.StringField),
	codec.Named("canChangeName", codec.BooleanField),
	codec.Named("safetyLocked", codec.BooleanField),
}

// Params contains USER_INFO packet data.
type Params struct {
	// UserID identifies the player.
	UserID int32

	// Username stores the visible player name.
	Username string

	// Figure stores the avatar figure string.
	Figure string

	// Gender stores the avatar gender code.
	Gender string

	// Motto stores the public profile motto.
	Motto string

	// RealName stores the optional real name.
	RealName string

	// CanChangeName reports whether username changes are allowed.
	CanChangeName bool

	// DirectMail reports whether hotel direct mail is enabled.
	DirectMail bool

	// RespectsReceived stores the durable respect total.
	RespectsReceived int32

	// RespectsRemaining stores the remaining daily user-respect allowance.
	RespectsRemaining int32

	// RespectsPetRemaining stores the remaining daily pet-respect allowance.
	RespectsPetRemaining int32

	// StreamPublishingAllowed reports whether stream publishing is enabled.
	StreamPublishingAllowed bool

	// LastAccessDate stores the previous login date expected by Nitro.
	LastAccessDate string

	// SafetyLocked reports the hotel-managed account safety state.
	SafetyLocked bool
}

// Encode creates a USER_INFO packet.
func Encode(params Params) (codec.Packet, error) {
	values := make([]codec.Value, 0, len(Definition))
	values = append(values, codec.Int32(params.UserID))
	values = append(values, codec.String(params.Username))
	values = append(values, codec.String(params.Figure))
	values = append(values, codec.String(params.Gender))
	values = append(values, codec.String(params.Motto))
	values = append(values, codec.String(params.RealName))
	values = append(values, codec.Bool(params.DirectMail))
	values = append(values, codec.Int32(params.RespectsReceived))
	values = append(values, codec.Int32(params.RespectsRemaining))
	values = append(values, codec.Int32(params.RespectsPetRemaining))
	values = append(values, codec.Bool(params.StreamPublishingAllowed))
	values = append(values, codec.String(params.LastAccessDate))
	values = append(values, codec.Bool(params.CanChangeName))
	values = append(values, codec.Bool(params.SafetyLocked))

	return codec.NewPacket(Header, Definition, values...)
}
