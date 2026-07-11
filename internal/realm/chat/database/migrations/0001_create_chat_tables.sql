--liquibase formatted sql

--changeset pixels:chat-0001 splitStatements:false
create table chat_global_word_filters (
    word varchar(32) primary key,
    created_at timestamptz not null default now(),
    constraint chat_global_word_filters_word_chk check (btrim(word) <> '')
);

create table chat_bubble_unlocks (
    bubble_id integer primary key,
    min_weight integer not null default 0,
    constraint chat_bubble_unlocks_id_chk check (bubble_id >= 0),
    constraint chat_bubble_unlocks_weight_chk check (min_weight >= 0)
);

create table chat_messages (
    id bigint generated always as identity,
    room_id bigint not null,
    player_id bigint not null,
    target_player_id bigint,
    kind varchar(8) not null,
    message text not null,
    censored boolean not null default false,
    created_at timestamptz not null,
    constraint chat_messages_kind_chk check (kind in ('talk', 'shout', 'whisper')),
    primary key (id, created_at)
) partition by range (created_at);

create index chat_messages_room_created_idx on chat_messages (room_id, created_at desc, id desc);
create index chat_messages_player_created_idx on chat_messages (player_id, created_at desc, id desc);

--rollback drop table chat_messages; drop table chat_bubble_unlocks; drop table chat_global_word_filters;
