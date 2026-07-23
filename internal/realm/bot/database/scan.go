package database

import botrecord "github.com/niflaot/pixels/internal/realm/bot/record"

const botColumns = `b.id,b.owner_player_id,p.username,b.room_id,b.behavior_type,b.name,b.motto,b.figure,b.gender,b.x,b.y,b.z,b.rotation,b.can_walk,b.dance_type,b.chat_auto,b.chat_random,b.chat_delay_seconds,b.bubble_style,b.effect_id,b.created_at,b.updated_at,b.version,coalesce(array_agg(l.line order by l.order_num) filter (where l.line is not null),'{}'::text[])`

const botFrom = ` from bots b join players p on p.id=b.owner_player_id left join bot_chat_lines l on l.bot_id=b.id `

const botGroup = ` group by b.id,p.username `

// rowScanner scans one PostgreSQL row.
type rowScanner interface {
	// Scan copies columns into destinations.
	Scan(...any) error
}

// rowsScanner scans a PostgreSQL result set.
type rowsScanner interface {
	// Next advances the result set.
	Next() bool
	// Scan copies current columns into destinations.
	Scan(...any) error
	// Err reports iteration failure.
	Err() error
	// Close releases result resources.
	Close()
}

// scanBot maps one joined bot row.
func scanBot(row rowScanner) (botrecord.Bot, error) {
	bot := botrecord.Bot{}
	err := row.Scan(&bot.ID, &bot.OwnerPlayerID, &bot.OwnerName, &bot.RoomID, &bot.BehaviorType, &bot.Name, &bot.Motto, &bot.Figure, &bot.Gender, &bot.X, &bot.Y, &bot.Z, &bot.Rotation, &bot.CanWalk, &bot.DanceType, &bot.ChatAuto, &bot.ChatRandom, &bot.ChatDelaySeconds, &bot.BubbleStyle, &bot.EffectID, &bot.CreatedAt, &bot.UpdatedAt, &bot.Version, &bot.ChatLines)
	return bot, err
}

// scanBots maps every joined bot row.
func scanBots(rows rowsScanner) ([]botrecord.Bot, error) {
	defer rows.Close()
	items := make([]botrecord.Bot, 0)
	for rows.Next() {
		item, err := scanBot(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
