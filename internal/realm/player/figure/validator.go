// Package figure validates bounded Nitro avatar figure syntax.
package figure

const (
	// MaxLength bounds the wire value before parsing.
	MaxLength = 512
	// MaxParts bounds figure components without allocating.
	MaxParts = 32
	// MaxColors bounds palette identifiers on one figure component.
	MaxColors = 8
)

// part contains one allocation-free parsed figure component.
type part struct {
	// kind stores the normalized two-byte type.
	kind [2]byte
	// setID stores the selected figure set.
	setID int32
	// colors stores bounded palette identifiers.
	colors [MaxColors]int32
	// colorCount stores the number of palette identifiers.
	colorCount int
}

// Valid reports whether a figure is a bounded sequence of unique typed sets.
func Valid(value string) bool {
	_, _, valid := parse(value)
	return valid
}

// parse returns allocation-free bounded figure components.
func parse(value string) ([MaxParts]part, int, bool) {
	var parts [MaxParts]part
	if len(value) == 0 || len(value) > MaxLength {
		return parts, 0, false
	}
	var types [MaxParts][2]byte
	partCount := 0
	index := 0
	for index < len(value) {
		if partCount == MaxParts || index+4 > len(value) || !letter(value[index]) || !letter(value[index+1]) || value[index+2] != '-' {
			return parts, 0, false
		}
		kind := [2]byte{lower(value[index]), lower(value[index+1])}
		for previous := 0; previous < partCount; previous++ {
			if types[previous] == kind {
				return parts, 0, false
			}
		}
		types[partCount] = kind
		parts[partCount].kind = kind
		partCount++
		index += 3
		var numberValue int32
		var valid bool
		index, numberValue, valid = number(value, index)
		if !valid {
			return parts, 0, false
		}
		parts[partCount-1].setID = numberValue
		for index < len(value) && value[index] == '-' {
			if parts[partCount-1].colorCount == MaxColors {
				return parts, 0, false
			}
			index++
			index, numberValue, valid = number(value, index)
			if !valid {
				return parts, 0, false
			}
			current := &parts[partCount-1]
			current.colors[current.colorCount] = numberValue
			current.colorCount++
		}
		if index == len(value) {
			break
		}
		if value[index] != '.' {
			return parts, 0, false
		}
		index++
		if index == len(value) {
			return parts, 0, false
		}
	}
	return parts, partCount, partCount > 0
}

// number consumes one positive bounded decimal integer.
func number(value string, index int) (int, int32, bool) {
	start := index
	number := uint64(0)
	for index < len(value) && value[index] >= '0' && value[index] <= '9' {
		number = number*10 + uint64(value[index]-'0')
		if number > 2147483647 {
			return index, 0, false
		}
		index++
	}
	return index, int32(number), index > start && number > 0
}

// letter reports whether one byte is an ASCII letter.
func letter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

// lower normalizes one validated ASCII letter.
func lower(value byte) byte {
	if value >= 'A' && value <= 'Z' {
		return value + ('a' - 'A')
	}
	return value
}
