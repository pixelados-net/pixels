package unit

// SetStatus stores a unit status value.
func (unit *Unit) SetStatus(key string, value string) {
	unit.statuses.set(key, value)
}

// ClearStatus removes a unit status value.
func (unit *Unit) ClearStatus(key string) {
	unit.statuses.clear(key)
}

// Statuses returns current statuses ordered by key.
func (unit *Unit) Statuses() []Status {
	return unit.statuses.snapshot()
}
