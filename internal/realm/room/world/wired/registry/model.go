// Package registry defines the immutable WIRED behavior manifest.
package registry

// Family classifies one WIRED furniture behavior.
type Family uint8

const (
	// FamilyTrigger starts a WIRED stack from a room event.
	FamilyTrigger Family = iota + 1
	// FamilyEffect changes room or durable state.
	FamilyEffect
	// FamilyCondition gates a stack.
	FamilyCondition
	// FamilyExtra modifies stack selection or represents a game pickup.
	FamilyExtra
	// FamilyHighscore projects durable game scores.
	FamilyHighscore
)

// SelectionPolicy describes whether a descriptor accepts furniture targets.
type SelectionPolicy uint8

const (
	// SelectionNone rejects selected furniture.
	SelectionNone SelectionPolicy = iota
	// SelectionOptional accepts zero or more selected furniture items.
	SelectionOptional
	// SelectionRequired requires at least one selected furniture item.
	SelectionRequired
)

// ActorPolicy describes which event actors a behavior accepts.
type ActorPolicy uint8

const (
	// ActorOptional accepts system events and all room-unit actors.
	ActorOptional ActorPolicy = iota
	// ActorPlayer requires a player actor.
	ActorPlayer
	// ActorUnit requires a player, bot, or pet actor.
	ActorUnit
	// ActorBot requires a bot actor.
	ActorBot
)

// Descriptor describes one canonical WIRED interaction.
type Descriptor struct {
	// Key stores the canonical furniture interaction type.
	Key string
	// Family stores the behavior family.
	Family Family
	// ClientCode stores Nitro's editor layout discriminator.
	ClientCode int32
	// Selection stores the furniture target policy.
	Selection SelectionPolicy
	// Actor stores the runtime actor policy.
	Actor ActorPolicy
	// Editor reports whether stock Nitro exposes a usable editor.
	Editor bool
	// Aliases stores accepted imported interaction names.
	Aliases []string
}

// ManifestSource identifies the audited upstream inventory revision.
const ManifestSource = "ArcturusMorningstar/Arcturus-Community master ItemManager + Nitro wired enums, audited 2026-07-15"
