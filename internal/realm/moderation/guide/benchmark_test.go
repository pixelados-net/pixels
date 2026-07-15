package guide

import "testing"

// BenchmarkDutyCount measures the small guide-pool lookup.
func BenchmarkDutyCount(b *testing.B) {
	manager := New(nil)
	for id := int64(1); id <= 10; id++ {
		manager.SetDuty(id, true, true, true)
	}
	b.ReportAllocs()
	for range b.N {
		manager.DutyCount()
	}
}
