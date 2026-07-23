// Package handlers adapts Nitro quest packets to durable quest behavior.
package handlers

import (
	"context"
	"time"

	progressionquest "github.com/niflaot/pixels/internal/realm/progression/quest"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	friendquest "github.com/niflaot/pixels/networking/inbound/messenger/friend/questcomplete"
	inaccept "github.com/niflaot/pixels/networking/inbound/progression/quest/accept"
	inactivate "github.com/niflaot/pixels/networking/inbound/progression/quest/activate"
	incancel "github.com/niflaot/pixels/networking/inbound/progression/quest/cancel"
	indaily "github.com/niflaot/pixels/networking/inbound/progression/quest/daily"
	inlist "github.com/niflaot/pixels/networking/inbound/progression/quest/list"
	inreject "github.com/niflaot/pixels/networking/inbound/progression/quest/reject"
	inseasonal "github.com/niflaot/pixels/networking/inbound/progression/quest/seasonal"
	intiming "github.com/niflaot/pixels/networking/inbound/progression/quest/timing"
	intracker "github.com/niflaot/pixels/networking/inbound/progression/quest/tracker"
	outcurrent "github.com/niflaot/pixels/networking/outbound/progression/quest/current"
	outdaily "github.com/niflaot/pixels/networking/outbound/progression/quest/daily"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
	outlist "github.com/niflaot/pixels/networking/outbound/progression/quest/list"
	outseasonal "github.com/niflaot/pixels/networking/outbound/progression/quest/seasonal"
	outtiming "github.com/niflaot/pixels/networking/outbound/progression/quest/timing"
)

// Handler owns every quest request adapter.
type Handler struct {
	// Service owns durable quest behavior.
	Service *progressionquest.Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs every real quest request.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	for _, header := range []uint16{friendquest.Header, inaccept.Header, inactivate.Header, incancel.Header, indaily.Header, inlist.Header, inreject.Header, inseasonal.Header, intiming.Header, intracker.Header} {
		_ = registry.Register(header, handler.handle)
	}
}

// handle decodes and dispatches one quest request.
func (handler Handler) handle(connection netconn.Context, packet codec.Packet) error {
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	ctx := context.Background()
	switch packet.Header {
	case friendquest.Header:
		if err := friendquest.Decode(packet); err != nil {
			return err
		}
		return handler.Service.ProgressTrigger(ctx, playerID, "friend.request.quest", "", 1)
	case inlist.Header:
		if err := inlist.Decode(packet); err != nil {
			return err
		}
		return handler.sendList(ctx, connection, playerID, false, true)
	case inseasonal.Header:
		if err := inseasonal.Decode(packet); err != nil {
			return err
		}
		return handler.sendSeasonal(ctx, connection, playerID)
	case inactivate.Header:
		questID, err := inactivate.Decode(packet)
		return handler.activate(ctx, playerID, int64(questID), err)
	case inaccept.Header:
		questID, err := inaccept.Decode(packet)
		return handler.activate(ctx, playerID, int64(questID), err)
	case incancel.Header:
		if err := incancel.Decode(packet); err != nil {
			return err
		}
		return handler.Service.Cancel(ctx, playerID, false)
	case inreject.Header:
		if err := inreject.Decode(packet); err != nil {
			return err
		}
		if err := handler.Service.RejectDaily(ctx, playerID, time.Now()); err != nil {
			return err
		}
		return handler.sendDaily(ctx, connection, playerID, false)
	case indaily.Header:
		request, err := indaily.Decode(packet)
		if err != nil {
			return err
		}
		return handler.sendDaily(ctx, connection, playerID, request.Easy)
	case intracker.Header:
		if err := intracker.Decode(packet); err != nil {
			return err
		}
		return handler.sendTracker(ctx, connection, playerID)
	case intiming.Header:
		if err := intiming.Decode(packet); err != nil {
			return err
		}
		code, until := handler.Service.Timing(time.Now())
		response, err := outtiming.Encode(code, until)
		if err != nil {
			return err
		}
		return connection.Send(ctx, response)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// activate validates decoding and activates one quest without disconnecting on expected rejection.
func (handler Handler) activate(ctx context.Context, playerID int64, questID int64, decodeErr error) error {
	if decodeErr != nil {
		return decodeErr
	}
	err := handler.Service.Activate(ctx, playerID, questID)
	if err == progressionrecord.ErrNotFound || err == progressionrecord.ErrUnavailable || err == progressionrecord.ErrConflict {
		return nil
	}
	return err
}

// sendList sends normal quests with campaign progress.
func (handler Handler) sendList(ctx context.Context, connection netconn.Context, playerID int64, seasonal bool, open bool) error {
	quests, progress, err := handler.Service.List(ctx, playerID, seasonal)
	if err != nil {
		return err
	}
	values := questValues(handler.Service, quests, progress)
	packet, err := outlist.Encode(values, open)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// sendSeasonal sends currently available seasonal quests.
func (handler Handler) sendSeasonal(ctx context.Context, connection netconn.Context, playerID int64) error {
	quests, progress, err := handler.Service.List(ctx, playerID, true)
	if err != nil {
		return err
	}
	packet, err := outseasonal.Encode(questValues(handler.Service, quests, progress))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// sendDaily sends one deterministic daily offer or an explicit empty response.
func (handler Handler) sendDaily(ctx context.Context, connection netconn.Context, playerID int64, easy bool) error {
	var value *questdata.Quest
	quest, found, err := handler.Service.Daily(ctx, playerID, time.Now(), easy)
	if err != nil {
		return err
	}
	if found {
		mapped := progressionquest.Data(quest, progressionrecord.PlayerQuestState{}, 1, 0)
		value = &mapped
	}
	easyCount, hardCount := handler.Service.DailyCounts(time.Now())
	packet, err := outdaily.Encode(value, easyCount, hardCount)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// sendTracker resends the active quest when one exists.
func (handler Handler) sendTracker(ctx context.Context, connection netconn.Context, playerID int64) error {
	quest, state, found, err := handler.Service.Active(ctx, playerID)
	if err != nil || !found {
		return err
	}
	packet, err := outcurrent.Encode(progressionquest.Data(quest, state, 1, 0))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// playerID resolves one authenticated connection binding.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}

// questValues maps a quest list with derived campaign counts.
func questValues(service *progressionquest.Service, quests []progressionrecord.QuestDefinition, progress map[int64]progressionrecord.PlayerQuestState) []questdata.Quest {
	values := make([]questdata.Quest, 0, len(quests))
	for _, quest := range quests {
		total, completed := service.CampaignCounts(progress, quest.CampaignCode)
		values = append(values, progressionquest.Data(quest, progress[quest.ID], total, completed))
	}
	return values
}
