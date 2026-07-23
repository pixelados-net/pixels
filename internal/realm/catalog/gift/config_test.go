package gift

import "testing"

// TestNewOptionsLoadsCompleteWrappingConfiguration verifies embedded Nitro choices.
func TestNewOptionsLoadsCompleteWrappingConfiguration(t *testing.T) {
	options := NewOptions()
	if options.Price != 2 || len(options.Wrappers) != 10 || len(options.Boxes) != 8 ||
		len(options.Ribbons) != 11 || len(options.DefaultGifts) != 7 {
		t.Fatalf("unexpected wrapping options %#v", options)
	}
}

// TestResolveValidatesAndMapsSelectionIndexes verifies untrusted client choices.
func TestResolveValidatesAndMapsSelectionIndexes(t *testing.T) {
	options := Options{Wrappers: []int32{3372}, Boxes: []int32{0, 8}, Ribbons: []int32{2, 10}, DefaultGifts: []int32{187}}
	box, ribbon, valid := options.Resolve(3372, 1, 1)
	if !valid || box != 8 || ribbon != 10 {
		t.Fatalf("unexpected wrapping selection box=%d ribbon=%d valid=%t", box, ribbon, valid)
	}
	if _, _, valid = options.Resolve(9999, 0, 0); valid {
		t.Fatal("expected unknown wrapping sprite rejection")
	}
	box, ribbon, valid = options.Resolve(187, 2, 0)
	if !valid || box != 2 || ribbon != 2 {
		t.Fatalf("unexpected default wrapping selection box=%d ribbon=%d valid=%t", box, ribbon, valid)
	}
	if _, _, valid = options.Resolve(187, 3, 0); valid {
		t.Fatal("expected out-of-range box rejection")
	}
}
