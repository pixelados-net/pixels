package roller

import (
	"context"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"go.uber.org/zap"
)

// enqueuePersistence stores one bounded best-effort durable position update.
func (service *Service) enqueuePersistence(roomID int64, item worldfurniture.Item) {
	select {
	case service.persistence <- persistence{roomID: roomID, item: item}:
	default:
		service.logFailure(item.ID, roomID, "roller persistence queue full", nil)
	}
}

// runPersistence drains durable updates until shutdown and then flushes queued work.
func (service *Service) runPersistence() {
	defer close(service.done)
	for {
		select {
		case update := <-service.persistence:
			service.persist(update)
		case <-service.stop:
			for {
				select {
				case update := <-service.persistence:
					service.persist(update)
				default:
					return
				}
			}
		}
	}
}

// persist stores one final furniture position without affecting the completed roll.
func (service *Service) persist(update persistence) {
	item := update.item
	_, err := service.furniture.Move(context.Background(), furniturePlacement(
		item.ID, item.OwnerPlayerID, update.roomID, int(item.Point.X), int(item.Point.Y), item.Z.Units(), item.Rotation,
	))
	if err != nil {
		service.logFailure(item.ID, update.roomID, "persist rolled furniture", err)
	}
}

// logFailure records one best-effort persistence problem.
func (service *Service) logFailure(itemID int64, roomID int64, message string, err error) {
	if service.log == nil {
		return
	}
	fields := []zap.Field{zap.Int64("item_id", itemID), zap.Int64("room_id", roomID)}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	service.log.Warn(message, fields...)
}
