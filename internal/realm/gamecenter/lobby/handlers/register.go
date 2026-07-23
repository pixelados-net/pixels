// Package handlers adapts the Nitro Game Center protocol.
package handlers

import (
	gamecenterlobby "github.com/niflaot/pixels/internal/realm/gamecenter/lobby"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaccept "github.com/niflaot/pixels/networking/inbound/gamecenter/invite/accept"
	infriends "github.com/niflaot/pixels/networking/inbound/gamecenter/leaderboard/friends"
	inweekly "github.com/niflaot/pixels/networking/inbound/gamecenter/leaderboard/weekly"
	indir "github.com/niflaot/pixels/networking/inbound/gamecenter/lobby/directory"
	ininit "github.com/niflaot/pixels/networking/inbound/gamecenter/lobby/init"
	inlist "github.com/niflaot/pixels/networking/inbound/gamecenter/lobby/list"
	injoin "github.com/niflaot/pixels/networking/inbound/gamecenter/queue/join"
	inleave "github.com/niflaot/pixels/networking/inbound/gamecenter/queue/leave"
	inwinner "github.com/niflaot/pixels/networking/inbound/gamecenter/reward/winners"
	inchat "github.com/niflaot/pixels/networking/inbound/gamecenter/session/chat"
	inexit "github.com/niflaot/pixels/networking/inbound/gamecenter/session/exit"
	infull "github.com/niflaot/pixels/networking/inbound/gamecenter/session/fullstatus"
	inagain "github.com/niflaot/pixels/networking/inbound/gamecenter/session/playagain"
	inunloaded "github.com/niflaot/pixels/networking/inbound/gamecenter/session/unloaded"
	inready "github.com/niflaot/pixels/networking/inbound/gamecenter/stage/ready"
	inaccount "github.com/niflaot/pixels/networking/inbound/gamecenter/status/account"
	inget "github.com/niflaot/pixels/networking/inbound/gamecenter/status/get"
	"sort"
)

// Handler owns Game Center client requests.
type Handler struct {
	// Lobby serves game registrations and launch decisions.
	Lobby *gamecenterlobby.Service
}

// Register installs all 18 new Game Center inbound adapters.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	headers := []uint16{inaccount.Header, inlist.Header, inwinner.Header, infriends.Header, inexit.Header, injoin.Header, infull.Header, inleave.Header, inready.Header, inchat.Header, inweekly.Header, ininit.Header, inget.Header, inagain.Header, inunloaded.Header, indir.Header, inaccept.Header}
	sort.Slice(headers, func(left int, right int) bool { return headers[left] < headers[right] })
	for _, header := range headers {
		_ = registry.Register(header, handler.handle)
	}
}
