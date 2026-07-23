package quest

import (
	"context"
	"strconv"

	messengerrecord "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfriendnotification "github.com/niflaot/pixels/networking/outbound/messenger/friendnotification"
	outcancelled "github.com/niflaot/pixels/networking/outbound/progression/quest/cancelled"
	outcompleted "github.com/niflaot/pixels/networking/outbound/progression/quest/completed"
	outcurrent "github.com/niflaot/pixels/networking/outbound/progression/quest/current"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
	outlist "github.com/niflaot/pixels/networking/outbound/progression/quest/list"
	outroomforward "github.com/niflaot/pixels/networking/outbound/room/forward"
)

// LiveProjector publishes quest transitions to online players and friends.
type LiveProjector struct {
	// catalog resolves full campaign counts for filtered offers.
	catalog *progressionengine.Catalog
	// players resolves online player sessions.
	players *playerlive.Registry
	// connections sends packets to online sessions.
	connections *netconn.Registry
	// friends resolves quest notification audiences.
	friends messengerrecord.Store
}

// NewLiveProjector creates one live quest projector.
func NewLiveProjector(catalog *progressionengine.Catalog, players *playerlive.Registry, connections *netconn.Registry, friends messengerrecord.Store) *LiveProjector {
	return &LiveProjector{catalog: catalog, players: players, connections: connections, friends: friends}
}

// Accepted publishes one activation and optional cancellation.
func (projector *LiveProjector) Accepted(ctx context.Context, playerID int64, quest progressionrecord.QuestDefinition, previous int64) {
	if previous != 0 && previous != quest.ID {
		projector.cancel(ctx, playerID, false)
	}
	projector.current(ctx, playerID, quest, progressionrecord.PlayerQuestState{PlayerID: playerID, ActiveQuestID: quest.ID})
}

// Progressed publishes one current quest update.
func (projector *LiveProjector) Progressed(ctx context.Context, playerID int64, quest progressionrecord.QuestDefinition, state progressionrecord.PlayerQuestState) {
	projector.current(ctx, playerID, quest, state)
}

// Completed publishes one completed quest and friend notification.
func (projector *LiveProjector) Completed(ctx context.Context, playerID int64, quest progressionrecord.QuestDefinition) {
	packet, err := outcompleted.Encode(Data(quest, progressionrecord.PlayerQuestState{PlayerID: playerID, ActiveQuestID: quest.ID, Progress: quest.GoalAmount}, 1, 1), true)
	if err == nil {
		projector.send(ctx, playerID, packet)
	}
	projector.notifyFriends(ctx, playerID, quest.LocalizationCode)
}

// Listed publishes the refreshed quest catalog after one completion.
func (projector *LiveProjector) Listed(ctx context.Context, playerID int64, quests []progressionrecord.QuestDefinition, progress map[int64]progressionrecord.PlayerQuestState) {
	values := make([]questdata.Quest, 0, len(quests))
	for _, quest := range quests {
		total, completed := int32(0), int32(0)
		if projector.catalog != nil && projector.catalog.Current() != nil {
			total, completed = CampaignCounts(projector.catalog.Current().Catalog.Quests, progress, quest.CampaignCode)
		}
		values = append(values, Data(quest, progress[quest.ID], total, completed))
	}
	packet, err := outlist.Encode(values, true)
	if err == nil {
		projector.send(ctx, playerID, packet)
	}
}

// Cancelled publishes one quest cancellation.
func (projector *LiveProjector) Cancelled(ctx context.Context, playerID int64, expired bool) {
	projector.cancel(ctx, playerID, expired)
}

// RoomReward forwards one online player to the configured reward room.
func (projector *LiveProjector) RoomReward(ctx context.Context, playerID int64, roomID int64) {
	packet, err := outroomforward.Encode(int32(roomID))
	if err == nil {
		projector.send(ctx, playerID, packet)
	}
}

// current sends one active quest projection.
func (projector *LiveProjector) current(ctx context.Context, playerID int64, quest progressionrecord.QuestDefinition, state progressionrecord.PlayerQuestState) {
	packet, err := outcurrent.Encode(Data(quest, state, 1, 0))
	if err == nil {
		projector.send(ctx, playerID, packet)
	}
}

// cancel sends one cancellation response.
func (projector *LiveProjector) cancel(ctx context.Context, playerID int64, expired bool) {
	packet, err := outcancelled.Encode(expired)
	if err == nil {
		projector.send(ctx, playerID, packet)
	}
}

// notifyFriends sends one renderer-native quest completion notification.
func (projector *LiveProjector) notifyFriends(ctx context.Context, playerID int64, data string) {
	if projector.friends == nil {
		return
	}
	ids, err := projector.friends.ListFriendIDs(ctx, playerID)
	if err != nil {
		return
	}
	packet, err := outfriendnotification.Encode(strconv.FormatInt(playerID, 10), outfriendnotification.TypeQuestCompleted, data)
	if err != nil {
		return
	}
	for _, id := range ids {
		projector.send(ctx, id, packet)
	}
}

// send delivers one packet to an online player's connection.
func (projector *LiveProjector) send(ctx context.Context, playerID int64, packet codec.Packet) {
	if projector == nil || projector.players == nil || projector.connections == nil {
		return
	}
	player, found := projector.players.Find(playerID)
	if !found {
		return
	}
	peer := player.Peer()
	connection, found := projector.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if found {
		_ = connection.Send(ctx, packet)
	}
}
