// Package observability owns lock-free low-cardinality pet counters and timings.
package observability

const (
	// ResultSuccess identifies a completed operation.
	ResultSuccess Result = iota
	// ResultRejected identifies an expected domain rejection.
	ResultRejected
	// ResultFailed identifies an infrastructure or unexpected failure.
	ResultFailed
	// resultCount stores the bounded result array length.
	resultCount
)

const (
	// OperationGrant identifies a durable pet grant.
	OperationGrant Operation = iota
	// OperationPlace identifies room placement.
	OperationPlace
	// OperationPickup identifies room pickup.
	OperationPickup
	// OperationRespect identifies pet respect.
	OperationRespect
	// OperationPackage identifies package opening.
	OperationPackage
	// OperationAdmin identifies protected administration.
	OperationAdmin
	// operationCount stores the bounded workflow array length.
	operationCount
)

const (
	// DecisionNeed identifies an autonomous need decision.
	DecisionNeed Decision = iota
	// DecisionFollow identifies an autonomous follow decision.
	DecisionFollow
	// DecisionVocal identifies an autonomous speech decision.
	DecisionVocal
	// DecisionWander identifies an autonomous walk decision.
	DecisionWander
	// decisionCount stores the bounded behavior array length.
	decisionCount
)

const (
	// PathMoved identifies an accepted path intent.
	PathMoved PathResult = iota
	// PathBlocked identifies a rejected path intent.
	PathBlocked
	// PathCancelled identifies an invalidated path intent.
	PathCancelled
	// pathResultCount stores the bounded path result array length.
	pathResultCount
)

const (
	// BreedingStart identifies session start or owner confirmation.
	BreedingStart BreedingOperation = iota
	// BreedingConfirm identifies offspring confirmation.
	BreedingConfirm
	// BreedingCancel identifies reservation cancellation.
	BreedingCancel
	// breedingOperationCount stores the bounded breeding array length.
	breedingOperationCount
)

const (
	// ProductFood identifies food products.
	ProductFood ProductKind = iota
	// ProductDrink identifies drink products.
	ProductDrink
	// ProductToy identifies toy products.
	ProductToy
	// ProductNest identifies nest products.
	ProductNest
	// ProductSaddle identifies saddle products.
	ProductSaddle
	// ProductSupplement identifies plant supplements.
	ProductSupplement
	// ProductUnknown identifies unsupported product rules.
	ProductUnknown
	// productKindCount stores the bounded product array length.
	productKindCount
)

// Result classifies one bounded operation result.
type Result uint8

// Operation classifies durable pet workflows.
type Operation uint8

// Decision classifies autonomous room behavior.
type Decision uint8

// PathResult classifies world path intents.
type PathResult uint8

// BreedingOperation classifies breeding workflows.
type BreedingOperation uint8

// ProductKind classifies furniture-backed pet products.
type ProductKind uint8

// Classify returns success, expected rejection, or unexpected failure.
func Classify(err error, expected bool) Result {
	if err == nil {
		return ResultSuccess
	}
	if expected {
		return ResultRejected
	}
	return ResultFailed
}

// ResultCounters stores one counter per Result value.
type ResultCounters struct {
	// Success stores completed calls.
	Success uint64 `json:"success"`
	// Rejected stores expected domain rejections.
	Rejected uint64 `json:"rejected"`
	// Failed stores unexpected failures.
	Failed uint64 `json:"failed"`
}

// HistogramSnapshot stores bounded cumulative duration observations.
type HistogramSnapshot struct {
	// Buckets stores counts for <=50us, <=100us, <=250us, <=500us, <=1ms, <=5ms, <=25ms, and slower.
	Buckets [8]uint64 `json:"buckets"`
	// Count stores the total number of observations.
	Count uint64 `json:"count"`
	// Nanoseconds stores cumulative duration.
	Nanoseconds uint64 `json:"nanoseconds"`
}

// Snapshot contains process-wide low-cardinality pet telemetry.
type Snapshot struct {
	// Operations stores grant, place, pickup, respect, package, and admin counters.
	Operations [operationCount]ResultCounters `json:"operations"`
	// RoomCount stores currently loaded pet room generations.
	RoomCount int64 `json:"roomCount"`
	// BehaviorDecisions stores need, follow, vocal, and wander decisions.
	BehaviorDecisions [decisionCount]uint64 `json:"behaviorDecisions"`
	// Actions stores result counters indexed by protocol command ID.
	Actions [47]ResultCounters `json:"actions"`
	// PathResults stores moved, blocked, and cancelled path intents.
	PathResults [pathResultCount]uint64 `json:"pathResults"`
	// StatFlush stores persistence result counters.
	StatFlush ResultCounters `json:"statFlush"`
	// Breeding stores start, confirm, and cancel counters.
	Breeding [breedingOperationCount]ResultCounters `json:"breeding"`
	// ProductUses stores product results by bounded kind.
	ProductUses [productKindCount]ResultCounters `json:"productUses"`
	// RoomLoad stores room-load duration observations.
	RoomLoad HistogramSnapshot `json:"roomLoad"`
	// InventoryList stores inventory-read duration observations.
	InventoryList HistogramSnapshot `json:"inventoryList"`
	// BehaviorDue stores due-cycle duration observations.
	BehaviorDue HistogramSnapshot `json:"behaviorDue"`
	// Transaction stores durable workflow duration observations.
	Transaction HistogramSnapshot `json:"transaction"`
}

// New creates empty process-wide pet telemetry.
func New() *Metrics { return &Metrics{} }

// RecordOperation increments one durable workflow result.
func (metrics *Metrics) RecordOperation(kind Operation, result Result) {
	if metrics != nil && kind < operationCount && result < resultCount {
		metrics.operations[kind][result].Add(1)
	}
}

// RecordDecision increments one autonomous behavior decision.
func (metrics *Metrics) RecordDecision(kind Decision) {
	if metrics != nil && kind < decisionCount {
		metrics.decisions[kind].Add(1)
	}
}

// RecordAction increments one bounded protocol command result.
func (metrics *Metrics) RecordAction(commandID int32, result Result) {
	if metrics != nil && commandID >= 0 && commandID < 47 && result < resultCount {
		metrics.actions[commandID][result].Add(1)
	}
}

// RecordPath increments one world path result.
func (metrics *Metrics) RecordPath(result PathResult) {
	if metrics != nil && result < pathResultCount {
		metrics.paths[result].Add(1)
	}
}

// RecordPathError classifies one accepted or blocked world path intent.
func (metrics *Metrics) RecordPathError(err error) {
	if err == nil {
		metrics.RecordPath(PathMoved)
		return
	}
	metrics.RecordPath(PathBlocked)
}

// RecordStatFlush increments one persistence result.
func (metrics *Metrics) RecordStatFlush(result Result) {
	if metrics != nil && result < resultCount {
		metrics.statFlush[result].Add(1)
	}
}

// RecordBreeding increments one breeding workflow result.
func (metrics *Metrics) RecordBreeding(kind BreedingOperation, result Result) {
	if metrics != nil && kind < breedingOperationCount && result < resultCount {
		metrics.breeding[kind][result].Add(1)
	}
}

// RecordProduct increments one typed product result.
func (metrics *Metrics) RecordProduct(kind ProductKind, result Result) {
	if metrics != nil && kind < productKindCount && result < resultCount {
		metrics.products[kind][result].Add(1)
	}
}

// AddRoom changes the loaded pet-room generation gauge.
func (metrics *Metrics) AddRoom(delta int64) {
	if metrics != nil {
		metrics.roomCount.Add(delta)
	}
}
