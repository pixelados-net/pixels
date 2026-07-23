package openapi

// GroupForumPageRequest stores bounded forum pagination.
type GroupForumPageRequest struct {
	GroupReadRequest
	// Start stores the zero-based offset.
	Start int `query:"start,omitempty" minimum:"0"`
	// Amount bounds returned records.
	Amount int `query:"amount,omitempty" minimum:"1" maximum:"50" default:"50"`
}

// GroupForumThreadReadRequest identifies one thread page.
type GroupForumThreadReadRequest struct {
	GroupForumPageRequest
	// ThreadID identifies the thread.
	ThreadID int64 `path:"threadId" required:"true" minimum:"1"`
}

// GroupForumSettingsRequest stores four access policies and entitlement.
type GroupForumSettingsRequest struct {
	GroupVersionRequest
	// Enabled controls the forum entitlement.
	Enabled bool `json:"enabled" required:"true"`
	// Read controls reads.
	Read int16 `json:"read" required:"true" minimum:"0" maximum:"3"`
	// PostMessage controls replies.
	PostMessage int16 `json:"postMessage" required:"true" minimum:"0" maximum:"3"`
	// PostThread controls thread creation.
	PostThread int16 `json:"postThread" required:"true" minimum:"0" maximum:"3"`
	// Moderate controls group moderation.
	Moderate int16 `json:"moderate" required:"true" minimum:"0" maximum:"3"`
}

// GroupForumThreadRequest stores one optimistic thread moderation patch.
type GroupForumThreadRequest struct {
	GroupVersionRequest
	// ThreadID identifies the thread.
	ThreadID int64 `path:"threadId" required:"true" minimum:"1"`
	// Pinned optionally changes ordering.
	Pinned *bool `json:"pinned,omitempty"`
	// Locked optionally changes reply access.
	Locked *bool `json:"locked,omitempty"`
	// State optionally changes retained visibility.
	State *int16 `json:"state,omitempty" enum:"0,1,10,20"`
}

// GroupForumPostRequest stores one optimistic post moderation patch.
type GroupForumPostRequest struct {
	GroupVersionRequest
	// PostID identifies the retained post.
	PostID int64 `path:"postId" required:"true" minimum:"1"`
	// State changes retained visibility.
	State int16 `json:"state" required:"true" enum:"0,1,10,20"`
}

// GroupForumSettingsResponse documents forum policy output.
type GroupForumSettingsResponse struct {
	Enabled bool `json:"enabled"`
}

// GroupForumThreadsResponse documents one thread page.
type GroupForumThreadsResponse struct {
	Total int32 `json:"total"`
}

// GroupForumThreadResponse documents one thread mutation.
type GroupForumThreadResponse struct {
	ID int64 `json:"id"`
}

// GroupForumPostResponse documents one post mutation.
type GroupForumPostResponse struct {
	ID int64 `json:"id"`
}
