--liquibase formatted sql
--changeset pixels:pixels-progression-0004-quests
create table quest_campaigns (
    code text primary key, seasonal boolean not null default false, starts_at timestamptz null, ends_at timestamptz null,
    timing_code text not null default '', enabled boolean not null default true,
    constraint quest_campaign_code_chk check(char_length(code) between 1 and 70),
    constraint quest_campaign_window_chk check(starts_at is null or ends_at is null or starts_at<ends_at)
);
create table quest_definitions (
    id bigint generated always as identity primary key, campaign_code text not null references quest_campaigns(code),
    series_number integer not null check(series_number>0), name text not null, localization_code text not null,
    trigger_key text not null, goal_amount bigint not null check(goal_amount>0), goal_data text not null default '',
    reward_kind text not null, reward_currency_type integer not null default 0, reward_amount bigint not null default 0,
    reward_badge text not null default '', reward_definition_id bigint null references furniture_definitions(id),
    reward_room_id bigint null references rooms(id), daily boolean not null default false, easy boolean not null default true,
    sort_order integer not null default 0, enabled boolean not null default true, version bigint not null default 1 check(version>0),
    unique(campaign_code,series_number), constraint quest_reward_kind_chk check(reward_kind in('currency','badge','item','room'))
);
create index quest_definitions_trigger_idx on quest_definitions(trigger_key,id) where enabled;
create table player_quest_state (
    player_id bigint primary key references players(id) on delete cascade,
    active_quest_id bigint null references quest_definitions(id), accepted_at timestamptz null
);
create table player_quest_progress (
    player_id bigint not null references players(id) on delete cascade,
    quest_id bigint not null references quest_definitions(id) on delete cascade,
    progress bigint not null default 0 check(progress>=0), completed_at timestamptz null, updated_at timestamptz not null default now(),
    primary key(player_id,quest_id)
);
--rollback drop table if exists player_quest_progress; drop table if exists player_quest_state; drop table if exists quest_definitions; drop table if exists quest_campaigns;
