package settings

import (
	"context"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Pending contains coalesced client-owned settings fields.
type Pending struct {
	// PlayerID identifies the settings owner.
	PlayerID int64
	// Volume replaces all volume fields when present.
	Volume [3]int32
	// HasVolume reports whether Volume replaces persisted values.
	HasVolume bool
	// OldChat replaces legacy chat selection when present.
	OldChat bool
	// HasOldChat reports whether OldChat replaces persisted selection.
	HasOldChat bool
	// CameraFollowBlocked replaces camera-follow privacy when present.
	CameraFollowBlocked bool
	// HasCameraFollowBlocked reports whether CameraFollowBlocked replaces persisted privacy.
	HasCameraFollowBlocked bool
}

// Writer coalesces repeated client settings packets by player.
type Writer struct {
	// service persists final settings values.
	service *Service
	// log records asynchronous persistence failures.
	log *zap.Logger
	// interval controls periodic flush cadence.
	interval time.Duration
	// limit bounds distinct pending players.
	limit int
	// mutex protects pending replacements.
	mutex sync.Mutex
	// pending stores the latest fields per player.
	pending map[int64]Pending
}

// NewWriter creates a bounded settings writer.
func NewWriter(service *Service, log *zap.Logger, config Config) *Writer {
	if config.FlushInterval <= 0 {
		config.FlushInterval = 250 * time.Millisecond
	}
	if config.PendingLimit <= 0 {
		config.PendingLimit = 4096
	}
	return &Writer{service: service, log: log, interval: config.FlushInterval, limit: config.PendingLimit, pending: make(map[int64]Pending)}
}

// RegisterWriter owns the coalesced settings writer lifecycle.
func RegisterWriter(lifecycle fx.Lifecycle, writer *Writer) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		go writer.run(ctx, done)
		return nil
	}, OnStop: func(stopCtx context.Context) error {
		cancel()
		select {
		case <-done:
			return nil
		case <-stopCtx.Done():
			return stopCtx.Err()
		}
	}})
}

// EnqueueVolume replaces one player's pending volume fields.
func (writer *Writer) EnqueueVolume(playerID int64, system int32, furniture int32, trax int32) bool {
	return writer.enqueue(playerID, [3]int32{system, furniture, trax}, true, false, false, false, false)
}

// EnqueueOldChat replaces one player's pending legacy chat field.
func (writer *Writer) EnqueueOldChat(playerID int64, oldChat bool) bool {
	return writer.enqueue(playerID, [3]int32{}, false, oldChat, true, false, false)
}

// EnqueueCameraFollowBlocked replaces one player's pending camera privacy field.
func (writer *Writer) EnqueueCameraFollowBlocked(playerID int64, blocked bool) bool {
	return writer.enqueue(playerID, [3]int32{}, false, false, false, blocked, true)
}

// enqueue merges one pending field replacement.
func (writer *Writer) enqueue(playerID int64, volume [3]int32, hasVolume bool, oldChat bool, hasOldChat bool, cameraFollowBlocked bool, hasCameraFollowBlocked bool) bool {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	pending, found := writer.pending[playerID]
	if !found && len(writer.pending) >= writer.limit {
		return false
	}
	pending.PlayerID = playerID
	if hasVolume {
		pending.Volume = volume
		pending.HasVolume = true
	}
	if hasOldChat {
		pending.OldChat = oldChat
		pending.HasOldChat = true
	}
	if hasCameraFollowBlocked {
		pending.CameraFollowBlocked = cameraFollowBlocked
		pending.HasCameraFollowBlocked = true
	}
	writer.pending[playerID] = pending
	return true
}

// run flushes final values at a bounded cadence and on shutdown.
func (writer *Writer) run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(writer.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			writer.flush(context.Background())
			return
		case <-ticker.C:
			writer.flush(ctx)
		}
	}
}

// flush swaps pending state before performing cold-path writes.
func (writer *Writer) flush(ctx context.Context) {
	writer.mutex.Lock()
	pending := writer.pending
	writer.pending = make(map[int64]Pending, len(pending))
	writer.mutex.Unlock()
	for playerID, value := range pending {
		if value.HasVolume {
			_, writerErr := writer.service.SetVolume(ctx, playerID, value.Volume[0], value.Volume[1], value.Volume[2])
			writer.report(playerID, writerErr)
		}
		if value.HasOldChat {
			_, writerErr := writer.service.SetOldChat(ctx, playerID, value.OldChat)
			writer.report(playerID, writerErr)
		}
		if value.HasCameraFollowBlocked {
			_, writerErr := writer.service.SetCameraFollowBlocked(ctx, playerID, value.CameraFollowBlocked)
			writer.report(playerID, writerErr)
		}
	}
}

// report records one asynchronous settings persistence failure.
func (writer *Writer) report(playerID int64, err error) {
	if err != nil && writer.log != nil {
		writer.log.Warn("save player settings", zap.Int64("player_id", playerID), zap.Error(err))
	}
}
