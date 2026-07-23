package projection

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

// mannequinData stores durable mannequin appearance JSON.
type mannequinData struct {
	// Gender stores the outfit gender.
	Gender string `json:"gender"`
	// Figure stores mannequin clothing parts.
	Figure string `json:"figure"`
	// Name stores the visible outfit name.
	Name string `json:"name"`
}

// highscoreData stores durable bounded WIRED board JSON.
type highscoreData struct {
	// State stores the legacy board visual state.
	State string `json:"state"`
	// ScoreType stores Nitro's grouping mode.
	ScoreType int32 `json:"score_type"`
	// ClearType stores Nitro's UTC rollover mode.
	ClearType int32 `json:"clear_type"`
	// Entries stores ranked board rows.
	Entries []stuffdata.HighscoreEntry `json:"entries"`
}

// SpecializedObjectData maps durable state to Nitro object-data formats.
func SpecializedObjectData(interactionType string, extraData string) *stuffdata.Data {
	switch interactionType {
	case "mannequin":
		var data mannequinData
		if json.Unmarshal([]byte(extraData), &data) != nil {
			return nil
		}
		return stuffdata.Map([]stuffdata.Pair{{Key: "GENDER", Value: data.Gender}, {Key: "FIGURE", Value: data.Figure}, {Key: "OUTFIT_NAME", Value: data.Name}})
	case "background_toner":
		parts := strings.Split(extraData, ":")
		if len(parts) != 4 {
			return nil
		}
		values := make([]int32, 4)
		for index, part := range parts {
			value, err := strconv.Atoi(part)
			if err != nil || value < 0 || value > 255 || index == 0 && value > 1 {
				return nil
			}
			values[index] = int32(value)
		}
		return stuffdata.IntArray(values)
	case "wf_highscore":
		var data highscoreData
		if json.Unmarshal([]byte(extraData), &data) != nil || data.ScoreType < 0 || data.ScoreType > 2 || data.ClearType < 0 || data.ClearType > 3 || len(data.Entries) > 100 {
			return stuffdata.Board(stuffdata.Highscore{State: "0", ScoreType: 2, ClearType: 0})
		}
		for _, entry := range data.Entries {
			if len(entry.Users) == 0 || len(entry.Users) > 100 {
				return stuffdata.Board(stuffdata.Highscore{State: "0", ScoreType: data.ScoreType, ClearType: data.ClearType})
			}
		}
		return stuffdata.Board(stuffdata.Highscore{State: data.State, ScoreType: data.ScoreType, ClearType: data.ClearType, Entries: data.Entries})
	case "custom_variables":
		var values map[string]string
		if json.Unmarshal([]byte(extraData), &values) != nil || len(values) > 100 {
			return nil
		}
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		pairs := make([]stuffdata.Pair, 0, len(keys))
		for _, key := range keys {
			pairs = append(pairs, stuffdata.Pair{Key: key, Value: values[key]})
		}
		return stuffdata.Map(pairs)
	default:
		return nil
	}
}
