package model

// Version tracks optimistic locking state.
type Version struct {
	// Version is the optimistic locking version.
	Version int64
}

// Next returns the next optimistic locking version.
func (version Version) Next() int64 {
	return version.Version + 1
}
