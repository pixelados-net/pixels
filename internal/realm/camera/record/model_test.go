package record

import (
	"testing"
)

// TestPhotoDataJSON verifies Nitro's compact external-image contract.
func TestPhotoDataJSON(t *testing.T) {
	data := PhotoData{Timestamp: 10, UUID: "abc", RoomID: 40, URL: "https://storage/photo.png", Caption: "", Name: "demo", OwnerName: "demo", OwnerID: 12}
	encoded, err := data.JSON()
	if err != nil {
		t.Fatalf("encode photo data: %v", err)
	}
	want := `{"t":10,"u":"abc","s":40,"w":"https://storage/photo.png","m":"","n":"demo","o":"demo","oi":12}`
	if encoded != want {
		t.Fatalf("unexpected photo data\nwant %s\n got %s", want, encoded)
	}
}

// TestPhotoCompanionKey verifies Nitro's deterministic wall-image convention.
func TestPhotoCompanionKey(t *testing.T) {
	key, ok := PhotoCompanionKey("photos/7/abc.png")
	if !ok || key != "photos/7/abc_small.png" {
		t.Fatalf("unexpected companion key=%q ok=%t", key, ok)
	}
	if key, ok = PhotoCompanionKey("rooms/7/thumbnail.jpg"); ok || key != "" {
		t.Fatalf("unexpected invalid companion key=%q ok=%t", key, ok)
	}
}

// BenchmarkPhotoDataJSON measures the only repeated camera JSON construction.
func BenchmarkPhotoDataJSON(b *testing.B) {
	data := PhotoData{Timestamp: 10, UUID: "67445470-f8a2-4671-a049-3249c360f14c", RoomID: 40, URL: "https://storage.example/photos/12/67445470-f8a2-4671-a049-3249c360f14c.png", Name: "demo", OwnerName: "demo", OwnerID: 12}
	b.ReportAllocs()
	for range b.N {
		if _, err := data.JSON(); err != nil {
			b.Fatal(err)
		}
	}
}
