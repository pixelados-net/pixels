// Package decor adapts Nitro decorator packets to room furniture commands.
package decor

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	decorcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/decor"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	dimmersave "github.com/niflaot/pixels/networking/inbound/furniture/dimmer/save"
	dimmersettings "github.com/niflaot/pixels/networking/inbound/furniture/dimmer/settings"
	dimmertoggle "github.com/niflaot/pixels/networking/inbound/furniture/dimmer/toggle"
	mannequinlook "github.com/niflaot/pixels/networking/inbound/furniture/mannequin/look"
	mannequinname "github.com/niflaot/pixels/networking/inbound/furniture/mannequin/name"
	paintapply "github.com/niflaot/pixels/networking/inbound/furniture/paint/apply"
	postitget "github.com/niflaot/pixels/networking/inbound/furniture/postit/get"
	postitplace "github.com/niflaot/pixels/networking/inbound/furniture/postit/place"
	postitsave "github.com/niflaot/pixels/networking/inbound/furniture/postit/save"
	postitset "github.com/niflaot/pixels/networking/inbound/furniture/postit/set"
	tonerapply "github.com/niflaot/pixels/networking/inbound/furniture/toner/apply"
	"go.uber.org/zap"
)

// Register registers every decorator packet adapter.
func Register(registry *netconn.HandlerRegistry, handler decorcmd.Handler, log *zap.Logger) {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	register := func(header uint16, decode func(codec.Packet) (decorcmd.Command, error)) {
		_ = registry.Register(header, func(connection netconn.Context, packet codec.Packet) error {
			payload, err := decode(packet)
			if err != nil {
				return err
			}
			payload.Handler = connection
			return dispatcher.Dispatch(context.Background(), command.Envelope[decorcmd.Command]{Command: payload, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
		})
	}
	register(postitplace.Header, decodePostItPlace)
	register(postitsave.Header, decodePostItSave)
	register(postitget.Header, decodePostItGet)
	register(postitset.Header, decodePostItSet)
	register(paintapply.Header, decodeSurface)
	register(dimmersettings.Header, decodeDimmerSettings)
	register(dimmersave.Header, decodeDimmerSave)
	register(dimmertoggle.Header, decodeDimmerToggle)
	register(mannequinlook.Header, decodeMannequinLook)
	register(mannequinname.Header, decodeMannequinName)
	register(tonerapply.Header, decodeToner)
}

// decodePostItPlace maps post-it placement.
func decodePostItPlace(packet codec.Packet) (decorcmd.Command, error) {
	value, err := postitplace.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindPostItPlace, ItemID: int64(value.ItemID), WallPosition: value.WallPosition}, err
}

// decodePostItSave maps initial post-it content.
func decodePostItSave(packet codec.Packet) (decorcmd.Command, error) {
	value, err := postitsave.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindPostItSave, ItemID: int64(value.ItemID), WallPosition: value.WallPosition, Color: value.Color, Text: value.Text}, err
}

// decodePostItGet maps a post-it data request.
func decodePostItGet(packet codec.Packet) (decorcmd.Command, error) {
	value, err := postitget.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindPostItGet, ItemID: int64(value.ItemID)}, err
}

// decodePostItSet maps a post-it edit.
func decodePostItSet(packet codec.Packet) (decorcmd.Command, error) {
	value, err := postitset.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindPostItSet, ItemID: int64(value.ItemID), Color: value.Color, Text: value.Text}, err
}

// decodeSurface maps a room surface consumable.
func decodeSurface(packet codec.Packet) (decorcmd.Command, error) {
	value, err := paintapply.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindSurface, ItemID: int64(value.ItemID)}, err
}

// decodeDimmerSettings maps a preset request.
func decodeDimmerSettings(packet codec.Packet) (decorcmd.Command, error) {
	_, err := dimmersettings.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindDimmerSettings}, err
}

// decodeDimmerSave maps a preset save.
func decodeDimmerSave(packet codec.Packet) (decorcmd.Command, error) {
	value, err := dimmersave.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindDimmerSave, PresetID: value.PresetID, Type: value.Type, Color: value.Color, First: value.Brightness, Apply: value.Apply}, err
}

// decodeDimmerToggle maps a mood-light toggle.
func decodeDimmerToggle(packet codec.Packet) (decorcmd.Command, error) {
	_, err := dimmertoggle.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindDimmerToggle}, err
}

// decodeMannequinLook maps an outfit save.
func decodeMannequinLook(packet codec.Packet) (decorcmd.Command, error) {
	value, err := mannequinlook.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindMannequinLook, ItemID: int64(value.ItemID)}, err
}

// decodeMannequinName maps an outfit name save.
func decodeMannequinName(packet codec.Packet) (decorcmd.Command, error) {
	value, err := mannequinname.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindMannequinName, ItemID: int64(value.ItemID), Text: value.Name}, err
}

// decodeToner maps a toner HSL update.
func decodeToner(packet codec.Packet) (decorcmd.Command, error) {
	value, err := tonerapply.Decode(packet)
	return decorcmd.Command{Kind: decorcmd.KindTonerApply, ItemID: int64(value.ItemID), First: value.Hue, Second: value.Saturation, Third: value.Lightness}, err
}
