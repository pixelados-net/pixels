package permission

// HolderKind identifies one permission grant owner family.
type HolderKind string

const (
	// HolderPlayer marks an individual player's direct grants.
	HolderPlayer HolderKind = "player"
	// HolderGroup marks a shared permission group.
	HolderGroup HolderKind = "group"
)

// Permissible identifies an entity that can hold permission grants.
type Permissible interface {
	// HolderID identifies the underlying player or group.
	HolderID() int64
	// HolderKind reports the holder family.
	HolderKind() HolderKind
}
