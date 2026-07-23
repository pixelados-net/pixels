package request

// Quiz creates one quiz definition.
type Quiz struct {
	Audit
	// Code identifies the quiz.
	Code string `json:"code"`
	// Kind selects safety or poll behavior.
	Kind string `json:"kind"`
}

// QuizQuestion creates or replaces one quiz question.
type QuizQuestion struct {
	Audit
	// QuestionRef identifies client localization content.
	QuestionRef int32 `json:"questionRef"`
	// CorrectAnswerID stores the expected answer.
	CorrectAnswerID int32 `json:"correctAnswerId"`
}

// Poll launches one bounded room word quiz.
type Poll struct {
	Audit
	// RoomID identifies the active room.
	RoomID int64 `json:"roomId"`
	// Question stores the visible word-quiz prompt.
	Question string `json:"question"`
	// DurationSeconds stores the bounded answer window in seconds.
	DurationSeconds int32 `json:"durationSeconds"`
}
