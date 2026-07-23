package database

import (
	"context"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// Save replaces mutable settings and ordered chat lines.
func (repository *Repository) Save(ctx context.Context, bot botrecord.Bot) (botrecord.Bot, bool, error) {
	var saved botrecord.Bot
	var found bool
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executorFor(txCtx).Exec(txCtx, `update bots set name=$2,motto=$3,figure=$4,gender=$5,can_walk=$6,dance_type=$7,chat_auto=$8,chat_random=$9,chat_delay_seconds=$10,bubble_style=$11,effect_id=$12,updated_at=now(),version=version+1 where id=$1 and version=$13`, bot.ID, bot.Name, bot.Motto, bot.Figure, bot.Gender, bot.CanWalk, bot.DanceType, bot.ChatAuto, bot.ChatRandom, bot.ChatDelaySeconds, bot.BubbleStyle, bot.EffectID, bot.Version)
		if err != nil || command.RowsAffected() == 0 {
			return err
		}
		if _, err = repository.executorFor(txCtx).Exec(txCtx, `delete from bot_chat_lines where bot_id=$1`, bot.ID); err != nil {
			return err
		}
		for index, line := range bot.ChatLines {
			if _, err = repository.executorFor(txCtx).Exec(txCtx, `insert into bot_chat_lines(bot_id,order_num,line) values($1,$2,$3)`, bot.ID, index, line); err != nil {
				return err
			}
		}
		saved, found, err = repository.Find(txCtx, bot.ID)
		return err
	})
	return saved, found, err
}

// CloneRoom copies placed bots and their chat lines with one set-based statement.
func (repository *Repository) CloneRoom(ctx context.Context, sourceRoomID int64, targetRoomID int64, targetOwnerID int64) (int, error) {
	const query = `with source as (
select b.*,nextval(pg_get_serial_sequence('bots','id')) as new_id from bots b where room_id=$1 order by id
), inserted as (
insert into bots(id,owner_player_id,room_id,behavior_type,name,motto,figure,gender,x,y,z,rotation,can_walk,dance_type,chat_auto,chat_random,chat_delay_seconds,bubble_style,effect_id)
overriding system value select new_id,$3,$2,behavior_type,name,motto,figure,gender,x,y,z,rotation,can_walk,dance_type,chat_auto,chat_random,chat_delay_seconds,bubble_style,effect_id from source returning id
), lines as (
insert into bot_chat_lines(bot_id,order_num,line) select source.new_id,l.order_num,l.line from source join bot_chat_lines l on l.bot_id=source.id returning 1
) select count(*) from inserted`
	var count int
	err := repository.executorFor(ctx).QueryRow(ctx, query, sourceRoomID, targetRoomID, targetOwnerID).Scan(&count)
	return count, err
}
