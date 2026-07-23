package record

import "time"

// ThreadState identifies visible and moderated thread states.
type ThreadState int16

const (
	// ThreadOpen permits replies.
	ThreadOpen ThreadState = 0
	// ThreadClosed prevents ordinary replies.
	ThreadClosed ThreadState = 1
	// ThreadHiddenStaff hides a thread by staff action.
	ThreadHiddenStaff ThreadState = 10
	// ThreadHiddenAdmin hides a thread by group administration.
	ThreadHiddenAdmin ThreadState = 20
)

// Valid reports whether state is a supported moderation state.
func (state ThreadState) Valid() bool {
	return state == ThreadOpen || state == ThreadClosed || state == ThreadHiddenStaff || state == ThreadHiddenAdmin
}

// PostState identifies visible and moderated post states.
type PostState = ThreadState

// Thread stores one group-forum discussion.
type Thread struct {
	// ID identifies the thread.
	ID int64 `json:"id"`
	// GroupID identifies the owning group.
	GroupID int64 `json:"groupId"`
	// AuthorID identifies the original author.
	AuthorID int64 `json:"authorId"`
	// AuthorName stores an immutable author snapshot.
	AuthorName string `json:"authorName"`
	// Subject stores the plain-text heading.
	Subject string `json:"subject"`
	// State stores moderation visibility.
	State ThreadState `json:"state"`
	// Pinned reports prioritized ordering.
	Pinned bool `json:"pinned"`
	// Locked prevents ordinary replies.
	Locked bool `json:"locked"`
	// PostCount stores total retained messages.
	PostCount int32 `json:"postCount"`
	// UnreadCount stores viewer-specific unread messages.
	UnreadCount int32 `json:"unreadCount"`
	// LastPostID identifies the latest post.
	LastPostID int64 `json:"lastPostId"`
	// LastAuthorID identifies the latest author.
	LastAuthorID int64 `json:"lastAuthorId"`
	// LastAuthorName stores latest author snapshot.
	LastAuthorName string `json:"lastAuthorName"`
	// LastPostedAt stores latest activity.
	LastPostedAt time.Time `json:"lastPostedAt"`
	// ModeratorID identifies the latest moderator.
	ModeratorID *int64 `json:"moderatorId,omitempty"`
	// ModeratorName stores moderator snapshot.
	ModeratorName string `json:"moderatorName"`
	// ModerationReason stores the administrative reason.
	ModerationReason string `json:"moderationReason"`
	// ModeratedAt stores the latest moderation time.
	ModeratedAt *time.Time `json:"moderatedAt,omitempty"`
	// CreatedAt stores creation time.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Version stores the optimistic concurrency token.
	Version int64 `json:"version"`
}

// Post stores one forum message and author snapshot.
type Post struct {
	// ID identifies the post.
	ID int64 `json:"id"`
	// GroupID identifies the owning group.
	GroupID int64 `json:"groupId"`
	// ThreadID identifies the parent thread.
	ThreadID int64 `json:"threadId"`
	// Ordinal stores the stable zero-based message index.
	Ordinal int32 `json:"ordinal"`
	// AuthorID identifies the author.
	AuthorID int64 `json:"authorId"`
	// AuthorName stores the immutable name snapshot.
	AuthorName string `json:"authorName"`
	// AuthorFigure stores the immutable figure snapshot.
	AuthorFigure string `json:"authorFigure"`
	// Body stores filtered plain text.
	Body string `json:"body"`
	// State stores moderation visibility.
	State PostState `json:"state"`
	// ModeratorID identifies the latest moderator.
	ModeratorID *int64 `json:"moderatorId,omitempty"`
	// ModeratorName stores moderator snapshot.
	ModeratorName string `json:"moderatorName"`
	// ModerationReason stores the administrative reason.
	ModerationReason string `json:"moderationReason"`
	// ModeratedAt stores the moderation time.
	ModeratedAt *time.Time `json:"moderatedAt,omitempty"`
	// AuthorPostCount stores the author's group-forum total at read time.
	AuthorPostCount int32 `json:"authorPostCount"`
	// CreatedAt stores creation time.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Version stores the optimistic concurrency token.
	Version int64 `json:"version"`
}

// ReadMarker stores one monotonic forum read position.
type ReadMarker struct {
	// GroupID identifies the forum.
	GroupID int64
	// PlayerID identifies the reader.
	PlayerID int64
	// LastMessageID stores the greatest observed post identifier.
	LastMessageID int64
	// Flag stores Nitro marker metadata.
	Flag int32
	// UpdatedAt stores mutation time.
	UpdatedAt time.Time
}

// ForumSummary combines a group forum with viewer-specific unread data.
type ForumSummary struct {
	// Group stores group identity and counters.
	Group Group `json:"group"`
	// UnreadMessages stores authorized unread count.
	UnreadMessages int32 `json:"unreadMessages"`
	// LastPost stores latest visible post when present.
	LastPost *Post `json:"lastPost,omitempty"`
	// LeaderboardScore stores bounded popularity score.
	LeaderboardScore int32 `json:"leaderboardScore"`
}
