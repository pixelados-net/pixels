package openapi

import progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"

// ProgressionPollRequest documents one live room word quiz launch.
type ProgressionPollRequest struct {
	progressionrequest.Poll
}

// ProgressionObjectResponse documents one small progression result.
type ProgressionObjectResponse map[string]any
