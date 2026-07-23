package sanctiontradelock

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeRejectsWrongHeader verifies header validation.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	packet := codec.Packet{Header: Header + 1}
	if _, err := Decode(packet); err != codec.ErrUnexpectedHeader {
		t.Fatalf("err=%v", err)
	}
}

// TestDecodeWithoutOptionalIssue verifies Nitro may omit the issue identifier.
func TestDecodeWithoutOptionalIssue(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(2), codec.String("reason"), codec.Int32(4), codec.Int32(3))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Minutes != 4 || payload.TopicID != 3 || payload.IssueID != 0 {
		t.Fatalf("payload=%+v err=%v", payload, err)
	}
}

// TestDecodeWithOptionalIssue verifies Nitro's issue-linked packet shape.
func TestDecodeWithOptionalIssue(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(2), codec.String("reason"), codec.Int32(10080), codec.Int32(3), codec.Int32(9))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.IssueID != 9 {
		t.Fatalf("payload=%+v err=%v", payload, err)
	}
}
