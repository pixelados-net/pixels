package observability

import (
	"sync/atomic"
	"time"
)

// ObserveRoomLoad records one room-load duration.
func (metrics *Metrics) ObserveRoomLoad(duration time.Duration) {
	if metrics != nil {
		metrics.roomLoad.observe(duration)
	}
}

// ObserveInventoryList records one inventory-read duration.
func (metrics *Metrics) ObserveInventoryList(duration time.Duration) {
	if metrics != nil {
		metrics.inventoryList.observe(duration)
	}
}

// ObserveBehaviorDue records one due-cycle duration.
func (metrics *Metrics) ObserveBehaviorDue(duration time.Duration) {
	if metrics != nil {
		metrics.behaviorDue.observe(duration)
	}
}

// ObserveTransaction records one durable workflow duration.
func (metrics *Metrics) ObserveTransaction(duration time.Duration) {
	if metrics != nil {
		metrics.transaction.observe(duration)
	}
}

// Snapshot returns a consistent-enough lock-free telemetry view.
func (metrics *Metrics) Snapshot() Snapshot {
	if metrics == nil {
		return Snapshot{}
	}
	var snapshot Snapshot
	for kind := range snapshot.Operations {
		snapshot.Operations[kind] = loadResults(&metrics.operations[kind])
	}
	for kind := range snapshot.BehaviorDecisions {
		snapshot.BehaviorDecisions[kind] = metrics.decisions[kind].Load()
	}
	for commandID := range snapshot.Actions {
		snapshot.Actions[commandID] = loadResults(&metrics.actions[commandID])
	}
	for result := range snapshot.PathResults {
		snapshot.PathResults[result] = metrics.paths[result].Load()
	}
	for kind := range snapshot.Breeding {
		snapshot.Breeding[kind] = loadResults(&metrics.breeding[kind])
	}
	for kind := range snapshot.ProductUses {
		snapshot.ProductUses[kind] = loadResults(&metrics.products[kind])
	}
	snapshot.RoomCount = metrics.roomCount.Load()
	snapshot.StatFlush = loadResults(&metrics.statFlush)
	snapshot.RoomLoad = metrics.roomLoad.snapshot()
	snapshot.InventoryList = metrics.inventoryList.snapshot()
	snapshot.BehaviorDue = metrics.behaviorDue.snapshot()
	snapshot.Transaction = metrics.transaction.snapshot()
	return snapshot
}

// observe records one bounded duration without allocation.
func (histogram *histogram) observe(duration time.Duration) {
	if histogram == nil {
		return
	}
	thresholds := [...]time.Duration{50 * time.Microsecond, 100 * time.Microsecond, 250 * time.Microsecond, 500 * time.Microsecond, time.Millisecond, 5 * time.Millisecond, 25 * time.Millisecond}
	bucket := len(thresholds)
	for index, threshold := range thresholds {
		if duration <= threshold {
			bucket = index
			break
		}
	}
	histogram.buckets[bucket].Add(1)
	histogram.count.Add(1)
	if duration > 0 {
		histogram.nanoseconds.Add(uint64(duration))
	}
}

// snapshot loads one bounded histogram.
func (histogram *histogram) snapshot() HistogramSnapshot {
	var result HistogramSnapshot
	for index := range result.Buckets {
		result.Buckets[index] = histogram.buckets[index].Load()
	}
	result.Count = histogram.count.Load()
	result.Nanoseconds = histogram.nanoseconds.Load()
	return result
}

// loadResults loads one fixed result counter group.
func loadResults(values *[resultCount]atomic.Uint64) ResultCounters {
	return ResultCounters{Success: values[ResultSuccess].Load(), Rejected: values[ResultRejected].Load(), Failed: values[ResultFailed].Load()}
}
