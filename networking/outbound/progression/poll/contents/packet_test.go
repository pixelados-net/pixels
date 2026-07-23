package contents

import "testing"

// TestEncode verifies questions, choices, children, and the NPS tail.
func TestEncode(t *testing.T) {
	packet, err := Encode(Data{ID: 9, StartMessage: "start", EndMessage: "end", NPS: true, Questions: []Question{{ID: 1, Type: 1, Text: "Pick", Choices: []Choice{{Value: "a", Text: "A", Type: 0}}, Children: []Question{{ID: 2, Text: "Why?"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header != Header || len(packet.Payload) < 30 || packet.Payload[len(packet.Payload)-1] != 1 {
		t.Fatalf("unexpected packet: %+v", packet)
	}
}
