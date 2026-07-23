// Package settings contains the NAVIGATOR_SETTINGS_SAVE inbound packet.
package settings

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NAVIGATOR_SETTINGS_SAVE.
const Header uint16 = 3159

// Definition describes Navigator window preference fields.
var Definition = codec.Definition{
	codec.Named("windowX", codec.Int32Field), codec.Named("windowY", codec.Int32Field),
	codec.Named("windowWidth", codec.Int32Field), codec.Named("windowHeight", codec.Int32Field),
	codec.Named("leftPanelHidden", codec.BooleanField), codec.Named("resultsMode", codec.Int32Field),
}

// Payload contains decoded Navigator window preferences.
type Payload struct {
	// WindowX stores the window x coordinate.
	WindowX int32
	// WindowY stores the window y coordinate.
	WindowY int32
	// WindowWidth stores the window width.
	WindowWidth int32
	// WindowHeight stores the window height.
	WindowHeight int32
	// LeftPanelHidden reports whether the left panel is hidden.
	LeftPanelHidden bool
	// ResultsMode stores the default list display mode.
	ResultsMode int32
}

// Decode decodes NAVIGATOR_SETTINGS_SAVE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{WindowX: values[0].Int32, WindowY: values[1].Int32, WindowWidth: values[2].Int32,
		WindowHeight: values[3].Int32, LeftPanelHidden: values[4].Boolean, ResultsMode: values[5].Int32}, nil
}
