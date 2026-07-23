// Package issueclosed defines the moderation issue closed event.
package issueclosed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a closed issue.
const Name bus.Name = "moderation.issue.closed"

// Payload describes one issue resolution.
type Payload struct {
	// IssueID identifies the issue.
	IssueID int64
	// Resolution stores the close code.
	Resolution int32
	// ModeratorID identifies the resolver.
	ModeratorID int64
}
