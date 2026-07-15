package projection

import (
	"encoding/json"
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
	default:
		return nil
	}
}
