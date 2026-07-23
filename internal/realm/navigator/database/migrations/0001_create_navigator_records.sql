--liquibase formatted sql

--changeset pixels:pixels-navigator-0001-create-navigator-records
create table room_favorites (
    player_id bigint not null references players(id) on delete cascade,
    room_id bigint not null references rooms(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (player_id, room_id)
);

create table navigator_saved_searches (
    id bigint generated always as identity primary key,
    player_id bigint not null references players(id) on delete cascade,
    code text not null,
    filter text not null default '',
    localization text not null default '',
    created_at timestamptz not null default now(),
    constraint navigator_saved_searches_code_length_chk check (char_length(code) between 1 and 64),
    constraint navigator_saved_searches_filter_length_chk check (char_length(filter) <= 128)
);

create index navigator_saved_searches_player_id_idx
on navigator_saved_searches (player_id, created_at desc);

create table navigator_preferences (
    player_id bigint primary key references players(id) on delete cascade,
    window_x integer not null default 68,
    window_y integer not null default 42,
    window_width integer not null default 425,
    window_height integer not null default 592,
    left_panel_hidden boolean not null default false,
    results_mode smallint not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint navigator_preferences_size_chk check (window_width > 0 and window_height > 0)
);

create table navigator_category_preferences (
    player_id bigint not null references players(id) on delete cascade,
    code text not null,
    collapsed boolean not null default false,
    list_mode smallint not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (player_id, code),
    constraint navigator_category_preferences_code_length_chk check (char_length(code) between 1 and 64)
);

create table navigator_lifted_rooms (
    id bigint generated always as identity primary key,
    room_id bigint not null references rooms(id) on delete cascade,
    area_id integer not null default 0,
    image text not null default '',
    caption text not null default '',
    order_num integer not null default 0,
    starts_at timestamptz null,
    ends_at timestamptz null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint navigator_lifted_rooms_version_positive_chk check (version > 0)
);

create index navigator_lifted_rooms_active_idx
on navigator_lifted_rooms (order_num, id)
where deleted_at is null;
--rollback drop table if exists navigator_lifted_rooms;
--rollback drop table if exists navigator_category_preferences;
--rollback drop table if exists navigator_preferences;
--rollback drop table if exists navigator_saved_searches;
--rollback drop table if exists room_favorites;
