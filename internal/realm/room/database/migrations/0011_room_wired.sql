--liquibase formatted sql

--changeset pixels:pixels-room-0011-room-wired-settings
create table room_social_groups (
    room_id bigint primary key references rooms(id) on delete cascade,
    group_id bigint not null references social_groups(id) on delete cascade
);

create table room_wired_settings (
    item_id bigint primary key references furniture_items(id) on delete cascade,
    int_params jsonb not null default '[]'::jsonb,
    string_param text not null default '',
    selection_mode smallint not null default 0,
    delay_pulses integer not null default 0,
    version bigint not null default 1,
    updated_at timestamptz not null default now(),
    constraint room_wired_settings_delay_chk check (delay_pulses between 0 and 86400),
    constraint room_wired_settings_selection_chk check (selection_mode between 0 and 3),
    constraint room_wired_settings_version_chk check (version > 0)
);

create table room_wired_selected_items (
    wired_item_id bigint not null references room_wired_settings(item_id) on delete cascade,
    selected_item_id bigint not null references furniture_items(id) on delete cascade,
    ordinal integer not null check (ordinal >= 0),
    snapshot_state text null,
    snapshot_x smallint null,
    snapshot_y smallint null,
    snapshot_z numeric(6,2) null,
    snapshot_rotation smallint null check (snapshot_rotation is null or snapshot_rotation in (0,2,4,6)),
    primary key (wired_item_id, selected_item_id),
    unique (wired_item_id, ordinal)
);

create index room_wired_selected_reverse_idx on room_wired_selected_items (selected_item_id);

--changeset pixels:pixels-room-0012-room-wired-rewards
create table room_wired_rewards (
    id bigint generated always as identity primary key,
    wired_item_id bigint not null references room_wired_settings(item_id) on delete cascade,
    ordinal integer not null check (ordinal >= 0),
    kind text not null,
    reference text not null,
    amount bigint not null default 1,
    weight integer not null,
    stock integer null,
    unique (wired_item_id, ordinal),
    constraint room_wired_reward_kind_chk check (kind in ('furniture','badge','credits','currency','respect','catalog_offer')),
    constraint room_wired_reward_weight_chk check (weight > 0),
    constraint room_wired_reward_amount_chk check (amount > 0),
    constraint room_wired_reward_stock_chk check (stock is null or stock >= 0)
);

create table room_wired_reward_claims (
    wired_item_id bigint not null references room_wired_settings(item_id) on delete cascade,
    player_id bigint not null references players(id),
    reward_id bigint null references room_wired_rewards(id),
    claimed_at timestamptz not null default now(),
    period_key text not null,
    primary key (wired_item_id, player_id, period_key)
);

create index room_wired_reward_claims_history_idx on room_wired_reward_claims (player_id, claimed_at desc);

--changeset pixels:pixels-room-0013-room-wired-highscores
create table room_wired_highscore_entries (
    id bigint generated always as identity primary key,
    board_item_id bigint not null references furniture_items(id) on delete cascade,
    period_kind text not null,
    period_start date null,
    participant_key text not null,
    score bigint not null,
    wins bigint not null default 0,
    updated_at timestamptz not null default now(),
    unique nulls not distinct (board_item_id, period_kind, period_start, participant_key),
    constraint room_wired_highscore_period_chk check (period_kind in ('alltime','daily','weekly','monthly'))
);

create table room_wired_highscore_participants (
    entry_id bigint not null references room_wired_highscore_entries(id) on delete cascade,
    player_id bigint not null references players(id),
    ordinal integer not null check (ordinal >= 0),
    primary key (entry_id, player_id),
    unique (entry_id, ordinal)
);

create index room_wired_highscore_rank_idx on room_wired_highscore_entries (board_item_id, period_kind, period_start, score desc, updated_at asc);

--rollback drop table if exists room_wired_highscore_participants; drop table if exists room_wired_highscore_entries; drop table if exists room_wired_reward_claims; drop table if exists room_wired_rewards; drop index if exists room_wired_selected_reverse_idx; drop table if exists room_wired_selected_items; drop table if exists room_wired_settings; drop table if exists room_social_groups;
