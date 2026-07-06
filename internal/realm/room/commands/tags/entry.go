package tags

import (
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	outtags "github.com/niflaot/pixels/networking/outbound/room/tags"
)

// tagEntries maps room tags to protocol entries.
func tagEntries(tags []roommodel.Tag) []outtags.Entry {
	entries := make([]outtags.Entry, 0, len(tags))
	for _, tag := range tags {
		entries = append(entries, outtags.Entry{Tag: tag.Value, Count: 1})
	}

	return entries
}
