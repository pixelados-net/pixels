package unit

import "sort"

const (
	// StatusMove stores movement status key.
	StatusMove = "mv"

	// StatusSit stores sit status key.
	StatusSit = "sit"

	// StatusLay stores lay status key.
	StatusLay = "lay"

	// StatusFlatControl stores room controller level status.
	StatusFlatControl = "flatctrl"

	// StatusTrade stores the direct-trade activity status.
	StatusTrade = "trd"

	// StatusDance stores the persistent avatar dance id.
	StatusDance = "dance"

	// StatusSign stores a held sign only while its one-shot status packet is assembled.
	StatusSign = "sign"
)

// Status stores one unit status value.
type Status struct {
	// Key stores the status key.
	Key string

	// Value stores the status value.
	Value string
}

// statuses stores mutable unit statuses.
type statuses struct {
	// values stores status values by key.
	values map[string]string
}

// set stores a status value.
func (statuses *statuses) set(key string, value string) {
	if statuses.values == nil {
		statuses.values = make(map[string]string, 4)
	}
	statuses.values[key] = value
}

// clear removes a status value.
func (statuses *statuses) clear(key string) {
	if statuses.values == nil {
		return
	}
	delete(statuses.values, key)
}

// has reports whether a status key is present.
func (statuses statuses) has(key string) bool {
	_, found := statuses.values[key]

	return found
}

// snapshot returns statuses ordered by key.
func (statuses statuses) snapshot() []Status {
	if len(statuses.values) == 0 {
		return nil
	}
	result := make([]Status, 0, len(statuses.values))
	for key, value := range statuses.values {
		result = append(result, Status{Key: key, Value: value})
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left].Key < result[right].Key
	})

	return result
}
