package guardian

import moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"

// votingComplete reports whether every remaining accepted guardian voted.
func votingComplete(ticket *Ticket) bool {
	for _, reviewer := range ticket.Reviewers {
		if reviewer.Accepted && reviewer.Vote == nil {
			return false
		}
	}
	return true
}

// anonymize replaces speaker identities with deterministic aliases.
func anonymize(entries []moderationrecord.ChatEntry) []moderationrecord.ChatEntry {
	aliases := make(map[int64]string)
	values := make([]moderationrecord.ChatEntry, len(entries))
	for index, entry := range entries {
		values[index] = entry
		if entry.PlayerID != nil {
			alias := aliases[*entry.PlayerID]
			if alias == "" {
				alias = "User" + string(rune('A'+len(aliases)))
				aliases[*entry.PlayerID] = alias
			}
			values[index].PatternID = alias
			values[index].PlayerID = nil
		}
	}
	return values
}

// clone detaches mutable ticket storage.
func clone(ticket *Ticket) Ticket {
	value := *ticket
	value.Chatlog = append([]moderationrecord.ChatEntry(nil), ticket.Chatlog...)
	value.Reviewers = make(map[int64]*Reviewer, len(ticket.Reviewers))
	for id, reviewer := range ticket.Reviewers {
		copy := *reviewer
		if reviewer.Vote != nil {
			vote := *reviewer.Vote
			copy.Vote = &vote
		}
		value.Reviewers[id] = &copy
	}
	return value
}
