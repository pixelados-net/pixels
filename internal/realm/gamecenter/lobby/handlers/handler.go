package handlers

import (
	"context"
	"errors"
	"time"

	gamecenterlobby "github.com/niflaot/pixels/internal/realm/gamecenter/lobby"
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	infriends "github.com/niflaot/pixels/networking/inbound/gamecenter/leaderboard/friends"
	inweekly "github.com/niflaot/pixels/networking/inbound/gamecenter/leaderboard/weekly"
	indir "github.com/niflaot/pixels/networking/inbound/gamecenter/lobby/directory"
	ininit "github.com/niflaot/pixels/networking/inbound/gamecenter/lobby/init"
	inlist "github.com/niflaot/pixels/networking/inbound/gamecenter/lobby/list"
	injoin "github.com/niflaot/pixels/networking/inbound/gamecenter/queue/join"
	inleave "github.com/niflaot/pixels/networking/inbound/gamecenter/queue/leave"
	inwinner "github.com/niflaot/pixels/networking/inbound/gamecenter/reward/winners"
	inaccount "github.com/niflaot/pixels/networking/inbound/gamecenter/status/account"
	inget "github.com/niflaot/pixels/networking/inbound/gamecenter/status/get"
	outfriends "github.com/niflaot/pixels/networking/outbound/gamecenter/leaderboard/friends"
	outweekly "github.com/niflaot/pixels/networking/outbound/gamecenter/leaderboard/weekly"
	outlist "github.com/niflaot/pixels/networking/outbound/gamecenter/lobby/list"
	outload "github.com/niflaot/pixels/networking/outbound/gamecenter/lobby/load"
	outurl "github.com/niflaot/pixels/networking/outbound/gamecenter/lobby/loadurl"
	outjoined "github.com/niflaot/pixels/networking/outbound/gamecenter/queue/joined"
	outjoinfailed "github.com/niflaot/pixels/networking/outbound/gamecenter/queue/joinfailed"
	outleft "github.com/niflaot/pixels/networking/outbound/gamecenter/queue/left"
	outwinners "github.com/niflaot/pixels/networking/outbound/gamecenter/reward/winners"
	outaccount "github.com/niflaot/pixels/networking/outbound/gamecenter/status/account"
	outdirectory "github.com/niflaot/pixels/networking/outbound/gamecenter/status/directory"
	outstatus "github.com/niflaot/pixels/networking/outbound/gamecenter/status/game"
)

// handle dispatches one validated Game Center request.
func (handler Handler) handle(connection netconn.Context, packet codec.Packet) error {
	if handler.Lobby == nil {
		return nil
	}
	switch packet.Header {
	case inlist.Header:
		if err := inlist.Decode(packet); err != nil {
			return err
		}
		response, err := outlist.Encode(handler.Lobby.List())
		return send(connection, response, err)
	case ininit.Header:
		gameID, err := ininit.Decode(packet)
		if err != nil {
			return err
		}
		return handler.sendStatus(connection, gameID)
	case inget.Header:
		gameID, err := inget.Decode(packet)
		if err != nil {
			return err
		}
		return handler.sendStatus(connection, gameID)
	case inaccount.Header:
		gameID, err := inaccount.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outaccount.Encode(gameID, -1, 0)
		return send(connection, response, err)
	case indir.Header:
		if err := indir.Decode(packet); err != nil {
			return err
		}
		response, err := outdirectory.Encode(0, 0, 0, -1)
		return send(connection, response, err)
	case injoin.Header:
		gameID, err := injoin.Decode(packet)
		if err != nil {
			return err
		}
		return handler.launch(connection, gameID)
	case inleave.Header:
		gameID, err := inleave.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outleft.Encode(gameID)
		return send(connection, response, err)
	case inweekly.Header:
		request, err := inweekly.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outweekly.Encode(request.Year, request.Week, 0, request.Offset, minutesUntilWeek())
		return send(connection, response, err)
	case infriends.Header:
		request, err := infriends.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outfriends.Encode(request.Year, request.Week, 0, request.Offset, minutesUntilWeek())
		return send(connection, response, err)
	case inwinner.Header:
		gameID, err := inwinner.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outwinners.Encode(gameID, nil)
		return send(connection, response, err)
	default:
		return decodeNoop(packet)
	}
}

// sendStatus responds honestly for known and unknown games.
func (handler Handler) sendStatus(connection netconn.Context, gameID int32) error {
	status := int32(0)
	if _, err := handler.Lobby.FindLaunch(gameID); err != nil {
		status = 1
	}
	response, err := outstatus.Encode(gameID, status)
	return send(connection, response, err)
}

// launch sends queue admission followed by the configured external launcher.
func (handler Handler) launch(connection netconn.Context, gameID int32) error {
	game, err := handler.Lobby.FindLaunch(gameID)
	if errors.Is(err, gamecenterlobby.ErrUnavailable) {
		response, encodeErr := outjoinfailed.Encode(gameID, 1)
		return send(connection, response, encodeErr)
	}
	if err != nil {
		return err
	}
	response, encodeErr := outjoined.Encode(gameID)
	if err = send(connection, response, encodeErr); err != nil {
		return err
	}
	clientID := "pixels-" + time.Now().UTC().Format("20060102150405.000000000")
	if game.LaunchKind == gamecenterrecord.LaunchParameters {
		response, encodeErr = outload.Encode(outload.Data{GameTypeID: gameID, GameClientID: clientID, URL: game.LaunchURL, Quality: "high", ScaleMode: "showAll", FrameRate: 60, Parameters: map[string]string{}})
		err = send(connection, response, encodeErr)
	} else {
		response, encodeErr = outurl.Encode(gameID, clientID, game.LaunchURL)
		err = send(connection, response, encodeErr)
	}
	if err == nil {
		handler.Lobby.RecordLaunch()
	}
	return err
}

// send delivers one encoded response.
func send(connection netconn.Context, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), packet)
}

// minutesUntilWeek returns a bounded reset countdown.
func minutesUntilWeek() int32 {
	return minutesUntilWeekAt(time.Now().UTC())
}

// minutesUntilWeekAt returns minutes until the next Monday UTC boundary.
func minutesUntilWeekAt(now time.Time) int32 {
	now = now.UTC()
	days := (8 - int(now.Weekday())) % 7
	if days == 0 {
		days = 7
	}
	next := time.Date(now.Year(), now.Month(), now.Day()+days, 0, 0, 0, 0, time.UTC)
	return int32(next.Sub(now) / time.Minute)
}
