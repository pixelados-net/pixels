package chat

import "github.com/niflaot/pixels/internal/permission"

var (
	// BubbleAny allows selecting any protocol-supported chat bubble.
	BubbleAny = permission.RegisterNode("chat.bubble.any", "")
	// FloodImmune exempts a player from room flood control.
	FloodImmune = permission.RegisterNode("chat.flood.immune", "")
	// LengthUnlimited exempts a player from the message length limit.
	LengthUnlimited = permission.RegisterNode("chat.length.unlimited", "")
	// FilterImmune exempts a player from global and room word filters.
	FilterImmune = permission.RegisterNode("chat.filter.immune", "")
	// WhisperObserveAny allows passive observation of room whispers.
	WhisperObserveAny = permission.RegisterNode("chat.whisper.observe.any", "")
)
