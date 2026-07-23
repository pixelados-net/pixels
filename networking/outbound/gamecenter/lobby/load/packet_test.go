package load

import (
	"bytes"
	"testing"
)

// TestEncode verifies deterministic parameter ordering and the launch header.
func TestEncode(t *testing.T) {
	data := Data{GameTypeID: 3, GameClientID: "client", URL: "https://game", Quality: "high", ScaleMode: "showAll", FrameRate: 60, Parameters: map[string]string{"z": "2", "a": "1"}}
	first, err := Encode(data)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Encode(data)
	if err != nil {
		t.Fatal(err)
	}
	if first.Header != Header || !bytes.Equal(first.Payload, second.Payload) {
		t.Fatal("launch encoding is not deterministic")
	}
}
