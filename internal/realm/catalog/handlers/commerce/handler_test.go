package commerce

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ingiftpurchase "github.com/niflaot/pixels/networking/inbound/catalog/gift/purchase"
)

// TestDecodeGiftPreservesProductExtraData verifies trophy text reaches catalog commerce.
func TestDecodeGiftPreservesProductExtraData(t *testing.T) {
	t.Parallel()
	packet, err := codec.NewPacket(ingiftpurchase.Header, ingiftpurchase.Definition,
		codec.Int32(2), codec.Int32(3), codec.String("trophy text"), codec.String("alice"), codec.String("gift"),
		codec.Int32(3372), codec.Int32(8), codec.Int32(10), codec.Bool(true))
	if err != nil {
		t.Fatal(err)
	}
	command, err := decode(netconn.Context{}, packet)
	if err != nil {
		t.Fatal(err)
	}
	if command.ExtraData != "trophy text" || command.ReceiverName != "alice" || command.OfferID != 3 {
		t.Fatalf("unexpected gift command %#v", command)
	}
}
