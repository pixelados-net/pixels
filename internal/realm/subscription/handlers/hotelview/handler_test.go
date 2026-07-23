package hotelview

import (
	"testing"

	hotelviewcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/hotelview"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbonus "github.com/niflaot/pixels/networking/inbound/subscription/bonusrare/request"
	incampaign "github.com/niflaot/pixels/networking/inbound/subscription/campaign/start"
	incountdown "github.com/niflaot/pixels/networking/inbound/subscription/countdown"
)

// TestDecode maps every hotel-view packet and rejects unknown headers.
func TestDecode(t *testing.T) {
	connection := netconn.Context{ConnectionID: "subscription-test"}
	countdown, err := codec.NewPacket(incountdown.Header, incountdown.Definition, codec.String("2030-04-05 12:30"))
	if err != nil {
		t.Fatal(err)
	}
	campaign, err := codec.NewPacket(incampaign.Header, incampaign.Definition, codec.String("legacy"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		packet codec.Packet
		action hotelviewcmd.Action
		value  string
	}{{name: "bonus", packet: codec.Packet{Header: inbonus.Header}, action: hotelviewcmd.BonusRare},
		{name: "countdown", packet: countdown, action: hotelviewcmd.Countdown, value: "2030-04-05 12:30"},
		{name: "campaign noop", packet: campaign, action: hotelviewcmd.StartCampaign, value: "legacy"}} {
		t.Run(test.name, func(t *testing.T) {
			input, decodeErr := decode(connection, test.packet)
			if decodeErr != nil || input.Action != test.action || input.Value != test.value {
				t.Fatalf("input=%#v err=%v", input, decodeErr)
			}
		})
	}
	if _, err = decode(connection, codec.Packet{Header: 65535}); err == nil {
		t.Fatal("expected unexpected header")
	}
}

// TestRegisterHandlesNil verifies optional compatibility wiring is safe.
func TestRegisterHandlesNil(t *testing.T) {
	Register(nil, nil)
	registry := netconn.NewHandlerRegistry()
	Register(registry, func(netconn.Context, codec.Packet) error { return nil })
	if registry.Len() != 3 {
		t.Fatalf("handlers=%d", registry.Len())
	}
}
