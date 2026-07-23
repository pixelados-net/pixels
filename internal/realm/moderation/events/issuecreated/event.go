// Package issuecreated defines the moderation issue created event.
package issuecreated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a created issue.
const Name bus.Name = "moderation.issue.created"

// Payload describes queue-relevant issue fields.
type Payload struct {
	// IssueID identifies the issue.
	IssueID int64
	// ReporterID identifies the reporter.
	ReporterID int64
	// ReportedID identifies the target when present.
	ReportedID int64
	// TopicID identifies the selected topic.
	TopicID int64
}
