package entry

import (
	"context"
	"strconv"
)

// authorizePassword validates password lockout and hashing state.
func (service *Service) authorizePassword(ctx context.Context, request Request) (Result, error) {
	locked, err := service.locked(ctx, request.Room.ID, request.PlayerID)
	if err != nil {
		return Result{}, err
	}
	if locked {
		return Result{}, ErrEntryLocked
	}
	if passwordMatches(request.Room.PasswordHash, request.Password) {
		if service.redis == nil {
			return Result{}, nil
		}

		return Result{}, service.redis.Delete(ctx, attemptKey(request.Room.ID, request.PlayerID))
	}

	return service.failedPassword(ctx, request.Room.ID, request.PlayerID)
}

// locked reports whether password entry is frozen.
func (service *Service) locked(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	if service.redis == nil {
		return false, nil
	}
	_, found, err := service.redis.Find(ctx, lockoutKey(roomID, playerID))

	return found, err
}

// failedPassword records a failed attempt and applies lockout when needed.
func (service *Service) failedPassword(ctx context.Context, roomID int64, playerID int64) (Result, error) {
	if service.redis == nil {
		return Result{}, ErrWrongPassword
	}
	count, err := service.redis.Increment(ctx, attemptKey(roomID, playerID), service.config.AttemptWindow)
	if err != nil {
		return Result{}, err
	}
	if count < service.config.MaxPasswordAttempts {
		return Result{}, ErrWrongPassword
	}
	created, err := service.redis.SetIfAbsent(ctx, lockoutKey(roomID, playerID), []byte{'1'}, service.config.LockoutDuration())
	if err != nil {
		return Result{}, err
	}
	if !created {
		return Result{}, ErrEntryLocked
	}
	if err := service.redis.Delete(ctx, attemptKey(roomID, playerID)); err != nil {
		return Result{}, err
	}

	return Result{Alert: service.lockoutAlert()}, ErrEntryLocked
}

// lockoutAlert creates a localized lockout message.
func (service *Service) lockoutAlert() string {
	return service.lockoutMessage
}

// entryKey creates a compact Redis entry key.
func entryKey(prefix string, roomID int64, playerID int64) string {
	var storage [96]byte
	key := append(storage[:0], prefix...)
	key = strconv.AppendInt(key, roomID, 10)
	key = append(key, ':')
	key = strconv.AppendInt(key, playerID, 10)

	return string(key)
}

// attemptKey returns one password-attempt key.
func attemptKey(roomID int64, playerID int64) string {
	return entryKey("room:entry:attempts:", roomID, playerID)
}

// lockoutKey returns one password-lockout key.
func lockoutKey(roomID int64, playerID int64) string {
	return entryKey("room:entry:lockout:", roomID, playerID)
}
