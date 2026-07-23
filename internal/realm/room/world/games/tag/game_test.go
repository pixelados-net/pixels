package tag

import "testing"

// TestGameLifecycle verifies join, transfer, and deterministic tag reassignment.
func TestGameLifecycle(t *testing.T) {
	game := New(IceTag)
	if !game.Join(20) || !game.Join(10) || game.Tagger() != 20 {
		t.Fatal("unexpected join state")
	}
	if game.Transfer(10, 20, true) {
		t.Fatal("non-tagger transferred")
	}
	if !game.Transfer(20, 10, true) || game.Tagger() != 10 {
		t.Fatal("tag transfer failed")
	}
	game.Leave(10)
	if game.Tagger() != 20 {
		t.Fatal("tag was not reassigned")
	}
}

// TestEffects verifies every documented variant effect.
func TestEffects(t *testing.T) {
	tests := []struct {
		variant Variant
		female  bool
		tagger  bool
		want    int32
	}{{IceTag, false, false, 38}, {IceTag, true, false, 39}, {IceTag, false, true, 45}, {IceTag, true, true, 46}, {Rollerskate, false, false, 55}, {Rollerskate, true, false, 56}, {Rollerskate, false, true, 57}, {Rollerskate, true, true, 58}, {Bunnyrun, false, false, 0}, {Bunnyrun, false, true, 68}}
	for _, test := range tests {
		if got := Effect(test.variant, test.female, test.tagger); got != test.want {
			t.Errorf("effect=%d want %d", got, test.want)
		}
	}
}
