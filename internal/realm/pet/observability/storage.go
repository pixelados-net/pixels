package observability

import "sync/atomic"

// Metrics stores lock-free pet telemetry.
type Metrics struct {
	// operations stores durable workflow results by bounded kind.
	operations [operationCount][resultCount]atomic.Uint64
	// roomCount gauges loaded room generations containing the pet capability.
	roomCount atomic.Int64
	// decisions stores autonomous behavior decisions by bounded kind.
	decisions [decisionCount]atomic.Uint64
	// actions stores protocol command outcomes by command and result.
	actions [47][resultCount]atomic.Uint64
	// paths stores bounded world path outcomes.
	paths [pathResultCount]atomic.Uint64
	// statFlush stores deferred persistence outcomes.
	statFlush [resultCount]atomic.Uint64
	// breeding stores breeding outcomes by workflow kind.
	breeding [breedingOperationCount][resultCount]atomic.Uint64
	// products stores product outcomes by bounded kind.
	products [productKindCount][resultCount]atomic.Uint64
	// roomLoad stores room load duration observations.
	roomLoad histogram
	// inventoryList stores inventory list duration observations.
	inventoryList histogram
	// behaviorDue stores due-cycle duration observations.
	behaviorDue histogram
	// transaction stores durable transaction duration observations.
	transaction histogram
}

// histogram stores fixed duration buckets.
type histogram struct {
	// buckets stores bounded duration counts.
	buckets [8]atomic.Uint64
	// count stores all observations.
	count atomic.Uint64
	// nanoseconds stores cumulative duration.
	nanoseconds atomic.Uint64
}
