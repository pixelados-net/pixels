// Package userinfo contains MODERATION_USER_INFO projection.
package userinfo

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MODERATION_USER_INFO.
const Header uint16 = 2866

// Params stores Nitro moderator user information.
type Params struct {
	PlayerID            int32
	Username            string
	Look                string
	RegistrationMinutes int32
	LastLoginMinutes    int32
	Online              bool
	CFHCount            int32
	AbusiveCFHCount     int32
	WarningCount        int32
	BanCount            int32
	TradeLockCount      int32
	TradeExpiry         string
	LastPurchase        string
	IdentityID          int32
	IdentityBanCount    int32
	Email               string
	Classification      string
	LastSanction        string
	SanctionAgeHours    int32
}

// Encode creates moderator user information.
func Encode(p Params) (codec.Packet, error) {
	d := codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.Int32Field}
	return codec.NewPacket(Header, d, codec.Int32(p.PlayerID), codec.String(p.Username), codec.String(p.Look), codec.Int32(p.RegistrationMinutes), codec.Int32(p.LastLoginMinutes), codec.Bool(p.Online), codec.Int32(p.CFHCount), codec.Int32(p.AbusiveCFHCount), codec.Int32(p.WarningCount), codec.Int32(p.BanCount), codec.Int32(p.TradeLockCount), codec.String(p.TradeExpiry), codec.String(p.LastPurchase), codec.Int32(p.IdentityID), codec.Int32(p.IdentityBanCount), codec.String(p.Email), codec.String(p.Classification), codec.String(p.LastSanction), codec.Int32(p.SanctionAgeHours))
}
