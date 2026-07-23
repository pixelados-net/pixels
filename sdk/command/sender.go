// Package command contains plugin command sender and permission helpers.
package command

import (
	"context"

	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

const (
	// SenderKindPlayer identifies a connected player command source.
	SenderKindPlayer = "player"
	// SenderKindConsole identifies a trusted host command source.
	SenderKindConsole = "console"
)

// Sender is an identity capable of issuing and receiving command feedback.
type Sender interface {
	// Name returns a human-readable identity.
	Name() string
	// Kind identifies the sender family.
	Kind() string
	// HasPermission reports whether this sender holds one capability.
	HasPermission(string) bool
	// Reply sends feedback through the sender's originating channel.
	Reply(context.Context, string) error
}

// PlayerAccess contains the player operations needed by PlayerSender.
type PlayerAccess interface {
	// Message sends one system message to a connected player.
	Message(int64, string) error
	// HasPermission resolves one player permission node.
	HasPermission(int64, string) (bool, error)
}

// PlayerSender adapts one immutable connected player to Sender.
type PlayerSender struct {
	// player stores the immutable source snapshot.
	player sdkplayer.Player
	// access resolves feedback and permissions through the host.
	access PlayerAccess
}

// NewPlayerSender creates a player-backed command sender.
func NewPlayerSender(player sdkplayer.Player, access PlayerAccess) PlayerSender {
	return PlayerSender{player: player, access: access}
}

// Name returns the player username.
func (sender PlayerSender) Name() string { return sender.player.Username }

// Kind identifies a connected player sender.
func (PlayerSender) Kind() string { return SenderKindPlayer }

// HasPermission resolves one node through the real host permission system.
func (sender PlayerSender) HasPermission(node string) bool {
	allowed, err := sender.access.HasPermission(sender.player.ID, node)
	return err == nil && allowed
}

// Reply sends a system message to the player.
func (sender PlayerSender) Reply(ctx context.Context, message string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return sender.access.Message(sender.player.ID, message)
}

// ConsoleSender represents a trusted host-issued command source.
type ConsoleSender struct {
	// ReplyFunc optionally receives command feedback.
	ReplyFunc func(context.Context, string) error
}

// Name returns the console identity.
func (ConsoleSender) Name() string { return "console" }

// Kind identifies a console sender.
func (ConsoleSender) Kind() string { return SenderKindConsole }

// HasPermission grants every command capability to the host console.
func (ConsoleSender) HasPermission(string) bool { return true }

// Reply writes feedback when a console callback was configured.
func (sender ConsoleSender) Reply(ctx context.Context, message string) error {
	if sender.ReplyFunc == nil {
		return nil
	}

	return sender.ReplyFunc(ctx, message)
}
