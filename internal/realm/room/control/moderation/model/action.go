// Package model contains persistent room moderation records.
package model

// Action identifies one room moderation operation.
type Action string

const (
	// ActionKick immediately removes a player.
	ActionKick Action = "kick"
	// ActionMute temporarily prevents room chat.
	ActionMute Action = "mute"
	// ActionUnmute ends an active mute.
	ActionUnmute Action = "unmute"
	// ActionBan temporarily prevents room entry.
	ActionBan Action = "ban"
	// ActionUnban ends an active ban.
	ActionUnban Action = "unban"
)

// Valid reports whether the action is supported.
func (action Action) Valid() bool {
	switch action {
	case ActionKick, ActionMute, ActionUnmute, ActionBan, ActionUnban:
		return true
	default:
		return false
	}
}
