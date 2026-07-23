package profile

import (
	"context"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outinfo "github.com/niflaot/pixels/networking/outbound/user/info"
	"go.uber.org/zap"
)

// sendInfo encodes and sends one live user snapshot.
func sendInfo(connection netconn.Context, snapshot playerlive.Snapshot) error {
	packet, err := outinfo.Encode(outinfo.Params{UserID: int32(snapshot.ID), Username: snapshot.Username, Figure: snapshot.Look,
		Gender: string(snapshot.Gender), Motto: snapshot.Motto, CanChangeName: snapshot.AllowNameChange,
		RespectsReceived: snapshot.RespectsReceived, RespectsRemaining: snapshot.RespectsRemaining,
		RespectsPetRemaining: snapshot.RespectsPetRemaining, LastAccessDate: snapshot.LastAccessDate, SafetyLocked: snapshot.SafetyLocked})
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), packet)
}

// mutationFailed logs a recoverable mutation error and keeps the client connected.
func (handler Handler) mutationFailed(connection netconn.Context, playerID int64, operation string, err error) error {
	if handler.Log != nil {
		handler.Log.Warn("player profile mutation failed", zap.Int64("player_id", playerID), zap.String("operation", operation), zap.Error(err))
	}
	return handler.alert(connection, "user.profile.update_failed")
}
