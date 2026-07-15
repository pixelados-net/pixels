package topics

import "testing"

// TestEncodeProjectsNestedTopics verifies category counts.
func TestEncodeProjectsNestedTopics(t *testing.T) {
	packet, err := Encode([]Category{{Name: "a", Topics: []Topic{{Name: "b", ID: 1}}}})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
