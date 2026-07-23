package staff

import (
	"context"
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inuserinfo "github.com/niflaot/pixels/networking/inbound/moderation/staff/userinfo"
	invisits "github.com/niflaot/pixels/networking/inbound/moderation/staff/userrooms"
	outuserinfo "github.com/niflaot/pixels/networking/outbound/moderation/staff/userinfo"
	outvisits "github.com/niflaot/pixels/networking/outbound/moderation/staff/visits"
)

// userInfo sends player identity and punishment aggregates.
func (handler Handler) userInfo(connection netconn.Context, packet codec.Packet) error {
	payload, err := inuserinfo.Decode(packet)
	if err != nil {
		return err
	}
	if err = handler.canRead(context.Background(), connection); err != nil {
		return err
	}
	record, found, err := handler.PlayerRecords.FindByID(context.Background(), int64(payload.PlayerID))
	if err != nil || !found {
		return err
	}
	punishments, err := handler.Sanctions.History(context.Background(), int64(payload.PlayerID), 500)
	if err != nil {
		return err
	}
	params := outuserinfo.Params{PlayerID: payload.PlayerID, Username: record.Player.Username, Look: record.Profile.Look, Online: false}
	if _, ok := handler.Players.Find(int64(payload.PlayerID)); ok {
		params.Online = true
	}
	for _, value := range punishments {
		switch string(value.Kind) {
		case "warn":
			params.WarningCount++
		case "ban":
			params.BanCount++
		case "trade_lock":
			params.TradeLockCount++
		}
		if params.LastSanction == "" {
			params.LastSanction = value.IssuedAt.Format(time.RFC3339)
			params.SanctionAgeHours = int32(time.Since(value.IssuedAt).Hours())
		}
	}
	response, err := outuserinfo.Encode(params)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// visits sends recent room entries.
func (handler Handler) visits(connection netconn.Context, packet codec.Packet) error {
	payload, err := invisits.Decode(packet)
	if err != nil {
		return err
	}
	if err = handler.canRead(context.Background(), connection); err != nil {
		return err
	}
	record, found, err := handler.PlayerRecords.FindByID(context.Background(), int64(payload.PlayerID))
	if err != nil || !found {
		return err
	}
	items, err := handler.Moderation.Store().Visits(context.Background(), int64(payload.PlayerID), 100)
	if err != nil {
		return err
	}
	values := make([]outvisits.Visit, len(items))
	for index, item := range items {
		values[index] = outvisits.Visit{RoomID: int32(item.RoomID), RoomName: item.RoomName, Hour: int32(item.EnteredAt.Hour()), Minute: int32(item.EnteredAt.Minute())}
	}
	response, err := outvisits.Encode(payload.PlayerID, record.Player.Username, values)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}
