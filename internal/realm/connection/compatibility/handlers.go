// Package compatibility serves explicit empty snapshots for optional client systems.
package compatibility

import (
	"context"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inconcurrent "github.com/niflaot/pixels/networking/inbound/economy/community/concurrentprogress"
	inreward "github.com/niflaot/pixels/networking/inbound/economy/community/concurrentreward"
	inearned "github.com/niflaot/pixels/networking/inbound/economy/community/earnedprizes"
	inhall "github.com/niflaot/pixels/networking/inbound/economy/community/halloffame"
	inprogress "github.com/niflaot/pixels/networking/inbound/economy/community/progress"
	inredeem "github.com/niflaot/pixels/networking/inbound/economy/community/redeemprize"
	invote "github.com/niflaot/pixels/networking/inbound/economy/community/vote"
	inaliases "github.com/niflaot/pixels/networking/inbound/furniture/aliases"
	insongs "github.com/niflaot/pixels/networking/inbound/furniture/songinfo"
	ingetinterstitial "github.com/niflaot/pixels/networking/inbound/notification/legacy/getinterstitial"
	interstitialshown "github.com/niflaot/pixels/networking/inbound/notification/legacy/interstitialshown"
	infaqcategory "github.com/niflaot/pixels/networking/inbound/other/faq/getcategory"
	infaqtext "github.com/niflaot/pixels/networking/inbound/other/faq/gettext"
	infaqsearch "github.com/niflaot/pixels/networking/inbound/other/faq/search"
	inphonereset "github.com/niflaot/pixels/networking/inbound/other/phone/resetstate"
	inphonestatus "github.com/niflaot/pixels/networking/inbound/other/phone/setverificationstatus"
	inphonenumber "github.com/niflaot/pixels/networking/inbound/other/phone/trynumber"
	inphonecode "github.com/niflaot/pixels/networking/inbound/other/phone/verifycode"
	inpromos "github.com/niflaot/pixels/networking/inbound/session/promoarticles"
	outconcurrent "github.com/niflaot/pixels/networking/outbound/economy/community/concurrentprogress"
	outearned "github.com/niflaot/pixels/networking/outbound/economy/community/earnedprizes"
	outhall "github.com/niflaot/pixels/networking/outbound/economy/community/halloffame"
	outprogress "github.com/niflaot/pixels/networking/outbound/economy/community/progress"
	outvote "github.com/niflaot/pixels/networking/outbound/economy/community/voteevent"
	outaliases "github.com/niflaot/pixels/networking/outbound/furniture/aliases"
	outsongs "github.com/niflaot/pixels/networking/outbound/furniture/songinfo"
	outpromos "github.com/niflaot/pixels/networking/outbound/session/promoarticles"
)

// Register installs optional-system compatibility handlers.
func Register(registry *netconn.HandlerRegistry) {
	if registry == nil {
		return
	}
	_ = registry.Register(inpromos.Header, promoArticles)
	_ = registry.Register(inaliases.Header, furnitureAliases)
	_ = registry.Register(inhall.Header, communityHallOfFame)
	_ = registry.Register(inredeem.Header, communityGoals)
	_ = registry.Register(inprogress.Header, communityGoals)
	_ = registry.Register(inconcurrent.Header, communityGoals)
	_ = registry.Register(inearned.Header, communityGoals)
	_ = registry.Register(invote.Header, communityGoals)
	_ = registry.Register(inreward.Header, communityGoals)
	_ = registry.Register(insongs.Header, songInfo)
	_ = registry.Register(interstitialshown.Header, retiredLanding)
	_ = registry.Register(ingetinterstitial.Header, retiredLanding)
	for _, header := range []uint16{inphonereset.Header, inphonestatus.Header, inphonenumber.Header, inphonecode.Header, infaqcategory.Header, infaqtext.Header, infaqsearch.Header} {
		_ = registry.Register(header, retiredSupport)
	}
}

// retiredLanding validates abandoned interstitial telemetry without inventing an ad system.
// Nitro React ships no interstitial widget or message-event listener, so these
// composers cannot produce visible behavior in the bundled client.
func retiredLanding(_ netconn.Context, packet codec.Packet) error {
	switch packet.Header {
	case interstitialshown.Header:
		return interstitialshown.Decode(packet)
	case ingetinterstitial.Header:
		return ingetinterstitial.Decode(packet)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// retiredSupport validates abandoned phone and FAQ requests without side effects.
// Pixels has no telecom provider, Nitro React has no phone composer caller, and
// its FAQ entry is permanently disabled; accepting these strictly is safer than
// fabricating verification or support content.
func retiredSupport(_ netconn.Context, packet codec.Packet) error {
	switch packet.Header {
	case inphonereset.Header:
		return inphonereset.Decode(packet)
	case inphonestatus.Header:
		return inphonestatus.Decode(packet)
	case inphonenumber.Header:
		return inphonenumber.Decode(packet)
	case inphonecode.Header:
		return inphonecode.Decode(packet)
	case infaqcategory.Header:
		return infaqcategory.Decode(packet)
	case infaqtext.Header:
		return infaqtext.Decode(packet)
	case infaqsearch.Header:
		return infaqsearch.Decode(packet)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// promoArticles returns an explicit empty landing-page feed.
func promoArticles(connection netconn.Context, packet codec.Packet) error {
	if err := inpromos.Decode(packet); err != nil {
		return err
	}
	response, err := outpromos.Encode()
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// furnitureAliases returns an explicit empty alias map.
func furnitureAliases(connection netconn.Context, packet codec.Packet) error {
	if err := inaliases.Decode(packet); err != nil {
		return err
	}
	response, err := outaliases.Encode()
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// communityHallOfFame returns an explicit empty goal leaderboard.
func communityHallOfFame(connection netconn.Context, packet codec.Packet) error {
	goalCode, err := inhall.Decode(packet)
	if err != nil {
		return err
	}
	response, err := outhall.Encode(goalCode)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// communityGoals returns neutral snapshots for client-inaccessible landing systems.
// The bundled Nitro has no widget for these requests; Comet's concurrent-user
// implementation is intentionally not modeled without that client surface.
// See plan/finals/ECONOMY.md Part 0.3 for the packet-by-packet evidence.
func communityGoals(connection netconn.Context, packet codec.Packet) error {
	var response codec.Packet
	var err error
	switch packet.Header {
	case inredeem.Header:
		_, err = inredeem.Decode(packet)
		if err == nil {
			response, err = outearned.Encode()
		}
	case inprogress.Header:
		err = inprogress.Decode(packet)
		if err == nil {
			response, err = outprogress.Encode()
		}
	case inconcurrent.Header:
		err = inconcurrent.Decode(packet)
		if err == nil {
			response, err = outconcurrent.Encode(0, 0, 0)
		}
	case inearned.Header:
		err = inearned.Decode(packet)
		if err == nil {
			response, err = outearned.Encode()
		}
	case invote.Header:
		_, err = invote.Decode(packet)
		if err == nil {
			response, err = outvote.Encode(false)
		}
	case inreward.Header:
		err = inreward.Decode(packet)
		if err == nil {
			response, err = outconcurrent.Encode(0, 0, 0)
		}
	default:
		err = codec.ErrUnexpectedHeader
	}
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// songInfo validates requested IDs and returns an explicit empty song list.
func songInfo(connection netconn.Context, packet codec.Packet) error {
	if _, err := insongs.Decode(packet); err != nil {
		return err
	}
	response, err := outsongs.Encode()
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}
