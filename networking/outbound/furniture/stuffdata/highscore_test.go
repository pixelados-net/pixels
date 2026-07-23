package stuffdata

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendHighscoreMatchesNitroFormat verifies format-six field order.
func TestAppendHighscoreMatchesNitroFormat(t *testing.T) {
	payload, err := AppendHighscore(nil, Highscore{State: "1", ScoreType: 2, ClearType: 3, Entries: []HighscoreEntry{{Score: 50, Users: []string{"Alice", "Bob"}}}})
	if err != nil {
		t.Fatal(err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField}, payload)
	if err != nil || len(rest) != 0 {
		t.Fatalf("decode err=%v rest=%d", err, len(rest))
	}
	if values[0].Int32 != 6 || values[1].String != "1" || values[5].Int32 != 50 || values[7].String != "Alice" || values[8].String != "Bob" {
		t.Fatalf("unexpected values: %+v", values)
	}
}
