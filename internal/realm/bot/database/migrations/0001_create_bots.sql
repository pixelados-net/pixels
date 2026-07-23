--liquibase formatted sql
--changeset pixels:pixels-bot-0001-create-bots
create table bots (
    id bigint generated always as identity primary key,
    owner_player_id bigint not null references players(id),
    room_id bigint null references rooms(id) on delete cascade,
    behavior_type text not null default 'generic',
    name text not null,
    motto text not null default '',
    figure text not null,
    gender text not null,
    x integer null,
    y integer null,
    z double precision null,
    rotation smallint null,
    can_walk boolean not null default true,
    dance_type smallint not null default 0,
    chat_auto boolean not null default false,
    chat_random boolean not null default false,
    chat_delay_seconds integer not null default 10,
    bubble_style integer not null default 0,
    effect_id integer null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    constraint bots_behavior_chk check (behavior_type <> ''),
    constraint bots_name_length_chk check (char_length(name) between 1 and 15),
    constraint bots_gender_chk check (gender in ('M','F')),
    constraint bots_chat_delay_chk check (chat_delay_seconds between 7 and 604800),
    constraint bots_placement_chk check (
        (room_id is null and x is null and y is null and z is null and rotation is null) or
        (room_id is not null and x is not null and y is not null and z is not null and rotation between 0 and 7)
    )
);
create index bots_owner_inventory_idx on bots (owner_player_id, id) where room_id is null;
create index bots_room_idx on bots (room_id, id) where room_id is not null;

create table bot_chat_lines (
    bot_id bigint not null references bots(id) on delete cascade,
    order_num integer not null check (order_num >= 0),
    line text not null check (line <> ''),
    primary key (bot_id, order_num)
);
--rollback drop table if exists bot_chat_lines; drop table if exists bots;
