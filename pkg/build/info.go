// Package build exposes process build metadata shared by binaries and tests.
package build

const (
	// Name is the registered project name.
	Name = "pixels"

	// Version is the registered project version.
	Version = "0.1.0"
)

// CommitHash is the build commit hash set by linker flags.
var CommitHash = "dev"

// Info describes the current emulator build.
type Info struct {
	// Name stores the project name.
	Name string
	// Version stores the project version and short commit.
	Version string
	// Commit stores the short commit hash.
	Commit string
}

// DefaultInfo returns the default build metadata for local development.
func DefaultInfo() Info {
	return NewInfo(Name, Version, CommitHash)
}

// NewInfo creates build metadata from project and source control values.
func NewInfo(name string, version string, commitHash string) Info {
	commit := ShortCommit(commitHash)

	return Info{
		Name:    name,
		Version: version + "-" + commit,
		Commit:  commit,
	}
}

// ShortCommit returns the first eight characters of a commit hash.
func ShortCommit(commitHash string) string {
	if len(commitHash) <= 8 {
		return commitHash
	}

	return commitHash[:8]
}
