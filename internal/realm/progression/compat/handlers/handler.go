// Package handlers owns explicit unsupported progression compatibility flows.
package handlers

import (
	"context"
	"sort"
	"strings"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrandom "github.com/niflaot/pixels/networking/inbound/navigator/competition/random"
	inroom "github.com/niflaot/pixels/networking/inbound/navigator/competition/room"
	insearch "github.com/niflaot/pixels/networking/inbound/navigator/competition/search"
	insubmittable "github.com/niflaot/pixels/networking/inbound/navigator/competition/submittable"
	incompetitioninit "github.com/niflaot/pixels/networking/inbound/progression/competition/init"
	inpartof "github.com/niflaot/pixels/networking/inbound/progression/competition/partof"
	insubmit "github.com/niflaot/pixels/networking/inbound/progression/competition/submit"
	ingamelist "github.com/niflaot/pixels/networking/inbound/progression/game/list"
	ingameuser "github.com/niflaot/pixels/networking/inbound/progression/game/user"
	inresolutionopen "github.com/niflaot/pixels/networking/inbound/progression/resolution/open"
	inresolutionreset "github.com/niflaot/pixels/networking/inbound/progression/resolution/reset"
	outstatus "github.com/niflaot/pixels/networking/outbound/camera/competitionstatus"
	outentry "github.com/niflaot/pixels/networking/outbound/progression/competition/entry"
	outpartof "github.com/niflaot/pixels/networking/outbound/progression/competition/partof"
	outrooms "github.com/niflaot/pixels/networking/outbound/progression/competition/rooms"
	outgamelist "github.com/niflaot/pixels/networking/outbound/progression/game/list"
	outgameuser "github.com/niflaot/pixels/networking/outbound/progression/game/user"
	outresolutionlist "github.com/niflaot/pixels/networking/outbound/progression/resolution/list"
)

// Handler decodes compatibility features and serves real game achievements.
type Handler struct {
	// Catalog owns immutable achievement definitions.
	Catalog *progressionengine.Catalog
}

// Register installs every progression compatibility adapter.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	for _, header := range []uint16{inresolutionopen.Header, inresolutionreset.Header, ingamelist.Header, ingameuser.Header, incompetitioninit.Header, inpartof.Header, insubmit.Header, inroom.Header, insearch.Header, inrandom.Header, insubmittable.Header} {
		_ = registry.Register(header, handler.handle)
	}
}

// handle routes one compatibility request without inventing unavailable data.
func (handler Handler) handle(connection netconn.Context, packet codec.Packet) error {
	ctx := context.Background()
	switch packet.Header {
	case inresolutionopen.Header:
		request, err := inresolutionopen.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outresolutionlist.Encode(request.ItemID, nil, 0)
		return send(ctx, connection, response, err)
	case inresolutionreset.Header:
		_, err := inresolutionreset.Decode(packet)
		return err
	case ingamelist.Header:
		if err := ingamelist.Decode(packet); err != nil {
			return err
		}
		response, err := outgamelist.Encode(handler.gameAchievements())
		return send(ctx, connection, response, err)
	case ingameuser.Header:
		if _, err := ingameuser.Decode(packet); err != nil {
			return err
		}
		response, err := outgameuser.Encode()
		return send(ctx, connection, response, err)
	case incompetitioninit.Header:
		if err := incompetitioninit.Decode(packet); err != nil {
			return err
		}
		response, err := outstatus.Encode(false, "")
		return send(ctx, connection, response, err)
	case inpartof.Header:
		if _, err := inpartof.Decode(packet); err != nil {
			return err
		}
		response, err := outpartof.Encode(false, 0)
		return send(ctx, connection, response, err)
	case insubmit.Header:
		request, err := insubmit.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outentry.Encode(0, request.GoalCode, outentry.PrerequisitesNotMet, nil, nil)
		return send(ctx, connection, response, err)
	case insearch.Header:
		request, err := insearch.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outrooms.Encode(request.GoalID, request.PageIndex, 0)
		return send(ctx, connection, response, err)
	case inroom.Header:
		request, err := inroom.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outrooms.Encode(0, request.Index, 0)
		return send(ctx, connection, response, err)
	case inrandom.Header:
		if _, err := inrandom.Decode(packet); err != nil {
			return err
		}
		response, err := outrooms.Encode(0, 0, 0)
		return send(ctx, connection, response, err)
	case insubmittable.Header:
		if err := insubmittable.Decode(packet); err != nil {
			return err
		}
		response, err := outrooms.Encode(0, 0, 0)
		return send(ctx, connection, response, err)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// gameAchievements groups enabled game definitions for Nitro's lobby parser.
func (handler Handler) gameAchievements() []outgamelist.Group {
	if handler.Catalog == nil || handler.Catalog.Current() == nil {
		return nil
	}
	groups := make(map[int32][]outgamelist.Achievement)
	for _, definition := range handler.Catalog.Current().Catalog.Achievements {
		gameTypeID := gameType(definition.Category, definition.Subcategory)
		if !definition.Enabled || !definition.Visible || gameTypeID == 0 {
			continue
		}
		groups[gameTypeID] = append(groups[gameTypeID], outgamelist.Achievement{ID: int32(definition.ID), Name: definition.Name, Levels: int32(len(definition.Levels))})
	}
	ids := make([]int, 0, len(groups))
	for id := range groups {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	result := make([]outgamelist.Group, 0, len(ids))
	for _, id := range ids {
		achievements := groups[int32(id)]
		sort.Slice(achievements, func(left int, right int) bool { return achievements[left].ID < achievements[right].ID })
		result = append(result, outgamelist.Group{GameTypeID: int32(id), Achievements: achievements})
	}
	return result
}

// gameType maps progression subcategories to stable Game Center ids.
func gameType(category string, subcategory string) int32 {
	if !strings.EqualFold(strings.TrimSpace(category), "games") {
		return 0
	}
	switch strings.ToLower(strings.TrimSpace(subcategory)) {
	case "banzai":
		return 1
	case "freeze":
		return 2
	case "football":
		return 3
	case "tag":
		return 4
	default:
		return 5
	}
}

// send delivers an encoded compatibility response.
func send(ctx context.Context, connection netconn.Context, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}
