package database

import (
	"context"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// Topics lists topics in display order.
func (repository *Repository) Topics(ctx context.Context, includeDisabled bool) ([]moderationrecord.Topic, error) {
	statement := `select id,category,name_key,action,auto_reply_key,default_sanction_ladder,order_num,enabled from cfh_topics`
	if !includeDisabled {
		statement += ` where enabled=true`
	}
	statement += ` order by category,order_num,id`
	rows, err := repository.executor(ctx).Query(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]moderationrecord.Topic, 0)
	for rows.Next() {
		var value moderationrecord.Topic
		if err = rows.Scan(&value.ID, &value.Category, &value.NameKey, &value.Action, &value.AutoReplyKey, &value.DefaultSanctionLadder, &value.Order, &value.Enabled); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// Topic returns one topic by id, including disabled administrative records.
func (repository *Repository) Topic(ctx context.Context, id int64) (moderationrecord.Topic, bool, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select id,category,name_key,action,auto_reply_key,default_sanction_ladder,order_num,enabled from cfh_topics where id=$1`, id)
	if err != nil {
		return moderationrecord.Topic{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return moderationrecord.Topic{}, false, rows.Err()
	}
	var value moderationrecord.Topic
	err = rows.Scan(&value.ID, &value.Category, &value.NameKey, &value.Action, &value.AutoReplyKey, &value.DefaultSanctionLadder, &value.Order, &value.Enabled)
	return value, err == nil, err
}

// CreateTopic creates one call-for-help topic.
func (repository *Repository) CreateTopic(ctx context.Context, value moderationrecord.Topic) (moderationrecord.Topic, error) {
	err := repository.executor(ctx).QueryRow(ctx, `insert into cfh_topics(category,name_key,action,auto_reply_key,default_sanction_ladder,order_num,enabled) values($1,$2,$3,$4,$5,$6,$7) returning id`, value.Category, value.NameKey, value.Action, value.AutoReplyKey, value.DefaultSanctionLadder, value.Order, value.Enabled).Scan(&value.ID)
	return value, err
}

// UpdateTopic replaces one call-for-help topic.
func (repository *Repository) UpdateTopic(ctx context.Context, value moderationrecord.Topic) (moderationrecord.Topic, bool, error) {
	result, err := repository.executor(ctx).Exec(ctx, `update cfh_topics set category=$2,name_key=$3,action=$4,auto_reply_key=$5,default_sanction_ladder=$6,order_num=$7,enabled=$8 where id=$1`, value.ID, value.Category, value.NameKey, value.Action, value.AutoReplyKey, value.DefaultSanctionLadder, value.Order, value.Enabled)
	if err != nil {
		return moderationrecord.Topic{}, false, err
	}
	return value, result.RowsAffected() == 1, nil
}

// Presets lists moderator response presets.
func (repository *Repository) Presets(ctx context.Context, includeDisabled bool) ([]moderationrecord.Preset, error) {
	statement := `select id,category,message_key,enabled,order_num from moderation_presets`
	if !includeDisabled {
		statement += ` where enabled=true`
	}
	statement += ` order by category,order_num,id`
	rows, err := repository.executor(ctx).Query(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]moderationrecord.Preset, 0)
	for rows.Next() {
		var value moderationrecord.Preset
		if err = rows.Scan(&value.ID, &value.Category, &value.MessageKey, &value.Enabled, &value.Order); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// CreatePreset creates one moderator response preset.
func (repository *Repository) CreatePreset(ctx context.Context, value moderationrecord.Preset) (moderationrecord.Preset, error) {
	err := repository.executor(ctx).QueryRow(ctx, `insert into moderation_presets(category,message_key,enabled,order_num) values($1,$2,$3,$4) returning id`, value.Category, value.MessageKey, value.Enabled, value.Order).Scan(&value.ID)
	return value, err
}

// UpdatePreset replaces one moderator response preset.
func (repository *Repository) UpdatePreset(ctx context.Context, value moderationrecord.Preset) (moderationrecord.Preset, bool, error) {
	result, err := repository.executor(ctx).Exec(ctx, `update moderation_presets set category=$2,message_key=$3,enabled=$4,order_num=$5 where id=$1`, value.ID, value.Category, value.MessageKey, value.Enabled, value.Order)
	if err != nil {
		return moderationrecord.Preset{}, false, err
	}
	return value, result.RowsAffected() == 1, nil
}
