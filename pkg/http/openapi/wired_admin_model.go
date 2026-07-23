package openapi

// WiredRoomRequest identifies one room WIRED graph.
type WiredRoomRequest struct {
	RoomIDRequest
}

// WiredItemRequest identifies one WIRED furniture item in a room.
type WiredItemRequest struct {
	RoomIDRequest
	// ItemID identifies the WIRED furniture item.
	ItemID int64 `path:"itemId" required:"true" minimum:"1"`
}

// WiredGameRequest identifies one game lifecycle action.
type WiredGameRequest struct {
	RoomIDRequest
	// Action selects start, end, or reset.
	Action string `path:"action" required:"true" enum:"start,end,reset"`
}

// WiredVisibilityRequest contains an optimistic room-level display mutation.
type WiredVisibilityRequest struct {
	RoomIDRequest
	// ExpectedVersion prevents overwriting another room settings mutation.
	ExpectedVersion int64 `json:"expectedVersion" required:"true" minimum:"0"`
	// HideBoxes controls whether Nitro receives WIRED configuration boxes on entry.
	HideBoxes bool `json:"hideBoxes" required:"true"`
}

// WiredVisibilityResponse reports the committed room-level display setting.
type WiredVisibilityResponse struct {
	// RoomID identifies the updated room.
	RoomID int64 `json:"roomId" required:"true"`
	// HideBoxes reports the committed visibility setting.
	HideBoxes bool `json:"hideBoxes" required:"true"`
	// Version stores the new optimistic room version.
	Version int64 `json:"version" required:"true"`
}

// WiredConfigUpdateRequest contains an optimistic complete configuration.
type WiredConfigUpdateRequest struct {
	WiredItemRequest
	// ExpectedVersion stores zero for first save or the current version.
	ExpectedVersion int64 `json:"expectedVersion" required:"true" minimum:"0"`
	// IntParams stores behavior-specific integer settings.
	IntParams []int32 `json:"intParams" required:"true" maxItems:"32"`
	// StringParam stores bounded behavior-specific text.
	StringParam string `json:"stringParam" required:"true" maxLength:"2048"`
	// SelectionMode stores zero none, one ID, two ID/type, or three context.
	SelectionMode int32 `json:"selectionMode" required:"true" minimum:"0" maximum:"3"`
	// DelayPulses stores 500 millisecond action delay units.
	DelayPulses int32 `json:"delayPulses" required:"true" minimum:"0" maximum:"7200"`
	// TargetIDs stores ordered same-room furniture identifiers.
	TargetIDs []int64 `json:"targetIds" required:"true" maxItems:"20"`
}

// WiredRewardModel describes one normalized reward option.
type WiredRewardModel struct {
	// Kind identifies the delivery capability.
	Kind string `json:"kind" required:"true" enum:"furniture,badge,credits,currency,respect,catalog_offer"`
	// Reference stores the capability identifier.
	Reference string `json:"reference" required:"true" maxLength:"128"`
	// Amount stores the positive delivered quantity.
	Amount int64 `json:"amount" required:"true" minimum:"1"`
	// Weight stores positive integer selection weight.
	Weight int32 `json:"weight" required:"true" minimum:"1"`
	// Stock stores optional remaining global stock.
	Stock *int64 `json:"stock,omitempty" minimum:"0"`
}

// WiredRewardsRequest replaces all normalized rewards atomically.
type WiredRewardsRequest struct {
	WiredItemRequest
	// Items stores ordered reward choices.
	Items []WiredRewardModel `json:"items" required:"true" maxItems:"100"`
}

// WiredTargetResponse describes one selected furniture reference.
type WiredTargetResponse struct {
	// ItemID identifies selected furniture.
	ItemID int64 `json:"ItemID" required:"true"`
	// SpriteID identifies its definition for type matching.
	SpriteID int32 `json:"SpriteID" required:"true"`
}

// WiredConfigModel describes one persisted node.
type WiredConfigModel struct {
	// ItemID identifies the WIRED furniture item.
	ItemID int64 `json:"ItemID" required:"true"`
	// RoomID identifies the room.
	RoomID int64 `json:"RoomID" required:"true"`
	// Interaction stores the canonical behavior key.
	Interaction string `json:"Interaction" required:"true"`
	// SpriteID stores the client furniture sprite.
	SpriteID int32 `json:"SpriteID" required:"true"`
	// X stores stack tile x.
	X int `json:"X" required:"true"`
	// Y stores stack tile y.
	Y int `json:"Y" required:"true"`
	// IntParams stores behavior settings.
	IntParams []int32 `json:"IntParams" required:"true"`
	// StringParam stores behavior text.
	StringParam string `json:"StringParam" required:"true"`
	// SelectionMode stores the target policy.
	SelectionMode int32 `json:"SelectionMode" required:"true"`
	// DelayPulses stores action delay units.
	DelayPulses int32 `json:"DelayPulses" required:"true"`
	// Version stores the optimistic revision.
	Version int64 `json:"Version" required:"true"`
	// Targets stores ordered selected furniture.
	Targets []WiredTargetResponse `json:"Targets" required:"true"`
}

// WiredDescriptorModel describes one registry entry.
type WiredDescriptorModel struct {
	// Key stores the canonical interaction.
	Key string `json:"Key" required:"true"`
	// Family stores trigger, effect, condition, extra, or board numeric family.
	Family uint8 `json:"Family" required:"true"`
	// ClientCode stores Nitro's editor discriminator.
	ClientCode int32 `json:"ClientCode" required:"true"`
	// Selection stores the descriptor target policy.
	Selection uint8 `json:"Selection" required:"true"`
	// Actor stores the accepted actor family.
	Actor uint8 `json:"Actor" required:"true"`
	// Editor reports stock Nitro editor support.
	Editor bool `json:"Editor" required:"true"`
	// Aliases stores accepted imported interaction names.
	Aliases []string `json:"Aliases" required:"true"`
}

// WiredConfigResponse contains one node and its descriptor.
type WiredConfigResponse struct {
	// Config stores durable settings.
	Config WiredConfigModel `json:"config" required:"true"`
	// Descriptor stores runtime metadata.
	Descriptor WiredDescriptorModel `json:"descriptor" required:"true"`
}

// WiredRoomResponse contains a complete room graph listing.
type WiredRoomResponse struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId" required:"true"`
	// Loaded reports active compilation.
	Loaded bool `json:"loaded" required:"true"`
	// Items stores configured nodes.
	Items []WiredConfigResponse `json:"items" required:"true"`
}

// WiredActionResponse reports a completed mutation.
type WiredActionResponse struct {
	// Success reports successful completion.
	Success bool `json:"success" required:"true"`
}

// WiredTraceResponse contains a sanitized execution trace.
type WiredTraceResponse struct {
	// ID identifies the event.
	ID uint64 `json:"id" required:"true"`
	// Kind stores the trigger kind.
	Kind uint8 `json:"kind" required:"true"`
	// Stacks stores visited stack count.
	Stacks int `json:"stacks" required:"true"`
	// Effects stores attempted effect count.
	Effects int `json:"effects" required:"true"`
	// BudgetExhausted reports a safety stop.
	BudgetExhausted bool `json:"budgetExhausted" required:"true"`
	// StartedAt stores start time.
	StartedAt string `json:"startedAt" required:"true" format:"date-time"`
	// DurationNanoseconds stores execution duration.
	DurationNanoseconds int64 `json:"durationNanoseconds" required:"true"`
}

// WiredMetricsResponse contains low-cardinality WIRED runtime counters.
type WiredMetricsResponse struct {
	// Events stores processed events indexed by trigger kind.
	Events [18]uint64 `json:"events" required:"true"`
	// StackResults stores passed, failed, and errored stack evaluations.
	StackResults [3]uint64 `json:"stackResults" required:"true"`
	// EffectResults stores effect results indexed by status.
	EffectResults [4]uint64 `json:"effectResults" required:"true"`
	// BudgetExhausted stores budget-stopped traces.
	BudgetExhausted uint64 `json:"budgetExhausted" required:"true"`
	// CompileFailures stores failed generation compilations.
	CompileFailures uint64 `json:"compileFailures" required:"true"`
	// DelayedTasks stores outstanding delayed effects.
	DelayedTasks int64 `json:"delayedTasks" required:"true"`
	// CompileCount stores successful generation compilations.
	CompileCount uint64 `json:"compileCount" required:"true"`
	// CompileNanoseconds stores cumulative compilation duration.
	CompileNanoseconds uint64 `json:"compileNanoseconds" required:"true"`
	// TraceCount stores completed traces.
	TraceCount uint64 `json:"traceCount" required:"true"`
	// TraceNanoseconds stores cumulative trace duration.
	TraceNanoseconds uint64 `json:"traceNanoseconds" required:"true"`
}

// WiredRegistryResponse contains the immutable runtime manifest.
type WiredRegistryResponse struct {
	// Source identifies the upstream audit.
	Source string `json:"source" required:"true"`
	// Total stores descriptor count.
	Total int `json:"total" required:"true"`
	// Items stores stable descriptors.
	Items []WiredDescriptorModel `json:"items" required:"true"`
}
