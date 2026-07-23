package engine

import (
	"context"
	"fmt"
	"strconv"

	messengerrecord "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfriendnotification "github.com/niflaot/pixels/networking/outbound/messenger/friendnotification"
	outbadgeadd "github.com/niflaot/pixels/networking/outbound/progression/achievement/badgeadd"
	achievementdata "github.com/niflaot/pixels/networking/outbound/progression/achievement/data"
	outlevelup "github.com/niflaot/pixels/networking/outbound/progression/achievement/levelup"
	outprogress "github.com/niflaot/pixels/networking/outbound/progression/achievement/progress"
	outscore "github.com/niflaot/pixels/networking/outbound/progression/achievement/score"
	"go.uber.org/zap"
)

// AchievementData maps durable progress to Nitro's target-level achievement record.
func AchievementData(definition progressionrecord.AchievementDefinition, progress progressionrecord.PlayerAchievement) achievementdata.Achievement {
	maximum := int32(len(definition.Levels))
	target := progress.Level + 1
	if target < 1 {
		target = 1
	}
	if target > maximum {
		target = maximum
	}
	value := achievementdata.Achievement{ID: int32(definition.ID), Level: target, BadgeCode: fmt.Sprintf("ACH_%s%d", definition.Name, target), CurrentPoints: Clamp(progress.Progress), FinalLevel: progress.Level >= maximum, Category: definition.Category, Subcategory: definition.Subcategory, LevelCount: maximum}
	if target <= 0 || int(target) > len(definition.Levels) {
		return value
	}
	level := definition.Levels[target-1]
	value.ScoreLimit = Clamp(level.ProgressNeeded)
	value.LevelRewardPoints = Clamp(level.RewardAmount)
	value.RewardPointType = level.RewardCurrencyType
	if target > 1 {
		value.ScoreAtStart = Clamp(definition.Levels[target-2].ProgressNeeded)
	}
	value.FinalLevel = value.FinalLevel && progress.Progress >= level.ProgressNeeded
	return value
}

// Clamp safely projects one durable counter into Nitro's signed 32-bit wire field.
func Clamp(value int64) int32 {
	if value > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	if value < 0 {
		return 0
	}
	return int32(value)
}

// LiveProjector publishes committed progression into online sessions and rooms.
type LiveProjector struct {
	// players resolves online player snapshots.
	players *playerlive.Registry
	// rooms resolves active room projections.
	rooms *roomlive.Registry
	// connections sends direct and room packets.
	connections *netconn.Registry
	// badges resolves durable badge identifiers.
	badges *playerachievement.Service
	// friends resolves level-up notification audiences.
	friends messengerrecord.Store
	// log records best-effort projection failures.
	log *zap.Logger
}

// NewLiveProjector creates a post-commit progression projector.
func NewLiveProjector(players *playerlive.Registry, rooms *roomlive.Registry, connections *netconn.Registry, badges *playerachievement.Service, friends messengerrecord.Store, log *zap.Logger) *LiveProjector {
	return &LiveProjector{players: players, rooms: rooms, connections: connections, badges: badges, friends: friends, log: log}
}

// Project publishes one committed achievement transition.
func (projector *LiveProjector) Project(ctx context.Context, transition Transition) {
	if projector == nil || projector.players == nil || projector.connections == nil {
		return
	}
	player, online := projector.players.Find(transition.PlayerID)
	if !online {
		return
	}
	if len(transition.Mutation.Crossed) == 0 {
		return
	}
	data := AchievementData(transition.Definition, transition.Mutation.After)
	if packet, err := outprogress.Encode(data); err == nil {
		projector.send(ctx, player, packet)
	}
	player.SetAchievementScore(transition.Score)
	projector.projectRoom(ctx, transition.PlayerID, transition.Score)
	projector.projectLevels(ctx, player, transition)
	if packet, err := outscore.Encode(transition.Score); err == nil {
		projector.send(ctx, player, packet)
	}
}

// projectLevels publishes every level crossed by one atomic mutation.
func (projector *LiveProjector) projectLevels(ctx context.Context, player *playerlive.Player, transition Transition) {
	badges, err := projector.badges.List(ctx, transition.PlayerID)
	if err != nil {
		projector.logError("progression badge projection failed", err)
	}
	for _, level := range transition.Mutation.Crossed {
		code := fmt.Sprintf("ACH_%s%d", transition.Definition.Name, level.Level)
		badgeID := badgeIdentifier(badges, code)
		removed := ""
		if level.Level > 1 {
			removed = fmt.Sprintf("ACH_%s%d", transition.Definition.Name, level.Level-1)
		}
		packet, encodeErr := outlevelup.Encode(outlevelup.Data{Type: int32(transition.Definition.ID), Level: level.Level, BadgeID: 144, BadgeCode: code, Points: int32(level.RewardAmount), RewardPoints: level.RewardCurrencyType, BonusPoints: 10, AchievementID: 21, RemovedBadgeCode: removed, Category: transition.Definition.Category, ShowDialog: true})
		if encodeErr == nil {
			projector.send(ctx, player, packet)
		}
		if badgeID > 0 {
			if packet, encodeErr = outbadgeadd.Encode(badgeID, code); encodeErr == nil {
				projector.send(ctx, player, packet)
			}
		}
		projector.notifyFriends(ctx, transition.PlayerID, code)
	}
}

// projectRoom refreshes the public score without a database read.
func (projector *LiveProjector) projectRoom(ctx context.Context, playerID int64, score int32) {
	if projector.rooms == nil {
		return
	}
	active, found := projector.rooms.FindByPlayer(playerID)
	if found && active.UpdateOccupantAchievementScore(playerID, score) {
		_ = roombroadcast.RoomSpawn(ctx, projector.connections, active, playerID, 0)
	}
}

// notifyFriends sends the renderer's toolbar achievement notification to online friends.
func (projector *LiveProjector) notifyFriends(ctx context.Context, playerID int64, code string) {
	if projector.friends == nil {
		return
	}
	friendIDs, err := projector.friends.ListFriendIDs(ctx, playerID)
	if err != nil {
		projector.logError("progression friend lookup failed", err)
		return
	}
	packet, err := outfriendnotification.Encode(strconv.FormatInt(playerID, 10), outfriendnotification.TypeAchievementCompleted, code)
	if err != nil {
		return
	}
	for _, friendID := range friendIDs {
		if friend, found := projector.players.Find(friendID); found {
			projector.send(ctx, friend, packet)
		}
	}
}

// send delivers one packet to an online player's bound connection.
func (projector *LiveProjector) send(ctx context.Context, player *playerlive.Player, packet codec.Packet) {
	peer := player.Peer()
	connection, found := projector.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if found {
		_ = connection.Send(ctx, packet)
	}
}

// logError records one best-effort projection failure.
func (projector *LiveProjector) logError(message string, err error) {
	if projector.log != nil {
		projector.log.Warn(message, zap.Error(err))
	}
}

// badgeIdentifier resolves one durable badge id for the incremental inventory packet.
func badgeIdentifier(badges []playerachievement.Badge, code string) int32 {
	for _, badge := range badges {
		if badge.Code == code && badge.ID > 0 && badge.ID <= int64(^uint32(0)>>1) {
			return int32(badge.ID)
		}
	}
	return 0
}
