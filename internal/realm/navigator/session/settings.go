package session

import (
	"context"

	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navrecord "github.com/niflaot/pixels/internal/realm/navigator/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	insettings "github.com/niflaot/pixels/networking/inbound/navigator/session/settings"
)

// SettingsHandler persists Navigator window settings.
type SettingsHandler struct {
	// Navigator coordinates preference persistence.
	Navigator navservice.Manager
	// Writer coalesces repeated resize packets.
	Writer *PreferenceWriter
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings resolves authenticated sessions.
	Bindings *binding.Registry
	// PositionLimit bounds absolute multi-monitor coordinates.
	PositionLimit int
	// MinimumWidth bounds saved Navigator width.
	MinimumWidth int
	// MaximumWidth bounds saved Navigator width.
	MaximumWidth int
	// MinimumHeight bounds saved Navigator height.
	MinimumHeight int
	// MaximumHeight bounds saved Navigator height.
	MaximumHeight int
}

// RegisterSettings installs the Navigator settings adapter.
func RegisterSettings(registry *netconn.HandlerRegistry, handler SettingsHandler) {
	if registry == nil {
		return
	}
	_ = registry.Register(insettings.Header, handler.save)
}

// save validates and persists one complete preference replacement.
func (handler SettingsHandler) save(connection netconn.Context, packet codec.Packet) error {
	payload, err := insettings.Decode(packet)
	if err != nil {
		return err
	}
	player, _, err := Player(connection, handler.Bindings, handler.Players)
	if err != nil || handler.Navigator == nil {
		return err
	}
	if !handler.valid(payload) {
		return navservice.ErrInvalidPreference
	}
	preference := navrecord.Preference{
		PlayerID: player.ID(), WindowX: int(payload.WindowX), WindowY: int(payload.WindowY),
		WindowWidth: int(payload.WindowWidth), WindowHeight: int(payload.WindowHeight),
		LeftPanelHidden: payload.LeftPanelHidden, ResultsMode: int16(payload.ResultsMode),
	}
	if handler.Writer != nil && handler.Writer.Enqueue(preference) {
		return nil
	}
	_, err = handler.Navigator.SavePreference(context.Background(), preference)
	return err
}

// valid reports whether one complete preference fits configured UI bounds.
func (handler SettingsHandler) valid(payload insettings.Payload) bool {
	positionLimit := handler.PositionLimit
	if positionLimit <= 0 {
		positionLimit = 32768
	}
	minimumWidth, maximumWidth := handler.MinimumWidth, handler.MaximumWidth
	minimumHeight, maximumHeight := handler.MinimumHeight, handler.MaximumHeight
	if minimumWidth <= 0 || maximumWidth < minimumWidth {
		minimumWidth, maximumWidth = 320, 4096
	}
	if minimumHeight <= 0 || maximumHeight < minimumHeight {
		minimumHeight, maximumHeight = 240, 2160
	}
	return int(payload.WindowX) >= -positionLimit && int(payload.WindowX) <= positionLimit && int(payload.WindowY) >= -positionLimit && int(payload.WindowY) <= positionLimit && int(payload.WindowWidth) >= minimumWidth && int(payload.WindowWidth) <= maximumWidth && int(payload.WindowHeight) >= minimumHeight && int(payload.WindowHeight) <= maximumHeight && payload.ResultsMode >= 0 && payload.ResultsMode <= 2
}
