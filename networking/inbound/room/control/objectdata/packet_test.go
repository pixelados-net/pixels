package objectdata

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies flattened key/value decoding and strict validation.
func TestDecode(t *testing.T) {
	payload, e := codec.AppendPayload(nil, PrefixDefinition, codec.Int32(7), codec.Int32(4))
	if e != nil {
		t.Fatal(e)
	}
	payload, e = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField, codec.StringField, codec.StringField}, codec.String("a"), codec.String("1"), codec.String("b"), codec.String("2"))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(codec.Packet{Header: Header, Payload: payload})
	if e != nil || v.ObjectID != 7 || v.Data["a"] != "1" || v.Data["b"] != "2" {
		t.Fatalf("value=%#v err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: payload}); e == nil {
		t.Fatal("expected header error")
	}
	bad, e := codec.NewPacket(Header, PrefixDefinition, codec.Int32(7), codec.Int32(3))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Decode(bad); e == nil {
		t.Fatal("expected odd-count error")
	}
}
