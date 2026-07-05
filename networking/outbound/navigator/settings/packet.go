// Package settings contains the NAVIGATOR_SETTINGS outbound packet.
package settings

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_SETTINGS packet identifier.
	Header uint16 = 518
)

// Params contains NAVIGATOR_SETTINGS packet data.
type Params struct {
	// WindowX stores the navigator window x position.
	WindowX int32
	// WindowY stores the navigator window y position.
	WindowY int32
	// WindowWidth stores the navigator window width.
	WindowWidth int32
	// WindowHeight stores the navigator window height.
	WindowHeight int32
	// LeftPanelHidden reports whether the left panel is hidden.
	LeftPanelHidden bool
	// ResultsMode stores the default results display mode.
	ResultsMode int32
}

// Definition describes the NAVIGATOR_SETTINGS payload fields.
var Definition = codec.Definition{
	codec.Named("windowX", codec.Int32Field),
	codec.Named("windowY", codec.Int32Field),
	codec.Named("windowWidth", codec.Int32Field),
	codec.Named("windowHeight", codec.Int32Field),
	codec.Named("leftPanelHidden", codec.BooleanField),
	codec.Named("resultsMode", codec.Int32Field),
}

// Encode creates a NAVIGATOR_SETTINGS packet.
func Encode(params Params) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.Int32(params.WindowX),
		codec.Int32(params.WindowY),
		codec.Int32(params.WindowWidth),
		codec.Int32(params.WindowHeight),
		codec.Bool(params.LeftPanelHidden),
		codec.Int32(params.ResultsMode),
	)
}
