// Package load encodes LOADGAME responses.
package load

import (
	"sort"

	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies LOADGAME.
const Header uint16 = 3654

// Data describes one parameterized external game launch.
type Data struct {
	// GameTypeID identifies the game type.
	GameTypeID int32
	// GameClientID identifies the launched client instance.
	GameClientID string
	// URL stores the external game URL.
	URL string
	// Quality stores the requested renderer quality.
	Quality string
	// ScaleMode stores the requested renderer scale mode.
	ScaleMode string
	// FrameRate stores the requested frame rate.
	FrameRate int32
	// MinMajorVersion stores the legacy minimum runtime major version.
	MinMajorVersion int32
	// MinMinorVersion stores the legacy minimum runtime minor version.
	MinMinorVersion int32
	// Parameters stores launch parameters.
	Parameters map[string]string
}

// Encode creates one deterministic parameterized launch response.
func Encode(data Data) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(data.GameTypeID), codec.String(data.GameClientID), codec.String(data.URL), codec.String(data.Quality), codec.String(data.ScaleMode), codec.Int32(data.FrameRate), codec.Int32(data.MinMajorVersion), codec.Int32(data.MinMinorVersion), codec.Int32(int32(len(data.Parameters))))
	keys := make([]string, 0, len(data.Parameters))
	for key := range data.Parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField}, codec.String(key), codec.String(data.Parameters[key]))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
