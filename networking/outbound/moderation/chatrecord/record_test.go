package chatrecord

import "testing"

// TestAppendEncodesEmptyRecord verifies base wire shape.
func TestAppendEncodesEmptyRecord(t *testing.T) {
	if payload, err := Append(nil, Record{}); err != nil || len(payload) == 0 {
		t.Fatal(err)
	}
}
