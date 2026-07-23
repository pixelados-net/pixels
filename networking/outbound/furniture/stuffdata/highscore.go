package stuffdata

import "github.com/niflaot/pixels/networking/codec"

const (
	// highscoreFormat identifies Nitro highscore object data.
	highscoreFormat int32 = 6
)

// HighscoreEntry stores one board score and its participant names.
type HighscoreEntry struct {
	// Score stores the ranked score.
	Score int32
	// Users stores current participant usernames.
	Users []string
}

// Highscore stores one complete board projection.
type Highscore struct {
	// State stores the legacy furniture visual state.
	State string
	// ScoreType stores classic, per-team, or most-wins mode.
	ScoreType int32
	// ClearType stores all-time, daily, weekly, or monthly rollover mode.
	ClearType int32
	// Entries stores ranked board rows.
	Entries []HighscoreEntry
}

// AppendHighscore appends Nitro object-data format six.
func AppendHighscore(dst []byte, value Highscore) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(highscoreFormat), codec.String(value.State), codec.Int32(value.ScoreType), codec.Int32(value.ClearType), codec.Int32(int32(len(value.Entries))))
	if err != nil {
		return dst, err
	}
	for _, entry := range value.Entries {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(entry.Score), codec.Int32(int32(len(entry.Users))))
		if err != nil {
			return dst, err
		}
		for _, username := range entry.Users {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(username))
			if err != nil {
				return dst, err
			}
		}
	}
	return payload, nil
}
