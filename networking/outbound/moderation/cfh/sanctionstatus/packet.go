// Package sanctionstatus contains USER_SANCTION_STATUS projection.
package sanctionstatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_SANCTION_STATUS.
const Header uint16 = 3679

// Params stores the exact Nitro sanction panel fields.
type Params struct {
	IsNew              bool
	Active             bool
	Name               string
	LengthHours        int32
	Reason             string
	CreatedAt          string
	ProbationHoursLeft int32
	NextName           string
	NextLengthHours    int32
	HasCustomMute      bool
	TradeLockExpiry    string
}

// Encode creates a sanction status panel packet.
func Encode(params Params) (codec.Packet, error) {
	definition := codec.Definition{codec.BooleanField, codec.BooleanField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Optional(codec.StringField)}
	values := []codec.Value{codec.Bool(params.IsNew), codec.Bool(params.Active), codec.String(params.Name), codec.Int32(params.LengthHours), codec.Int32(0), codec.String(params.Reason), codec.String(params.CreatedAt), codec.Int32(params.ProbationHoursLeft), codec.String(params.NextName), codec.Int32(params.NextLengthHours), codec.Int32(0), codec.Bool(params.HasCustomMute)}
	if params.TradeLockExpiry != "" {
		values = append(values, codec.String(params.TradeLockExpiry))
	}
	return codec.NewPacket(Header, definition, values...)
}
