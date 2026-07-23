--liquibase formatted sql

--changeset pixels:pixels-room-0002-create-room-records
create table room_categories (
    id bigint generated always as identity primary key,
    caption text not null,
    caption_key text not null,
    visible boolean not null default true,
    automatic boolean not null default false,
    automatic_key text not null default '',
    global_key text not null default '',
    staff_only boolean not null default false,
    order_num integer not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint room_categories_caption_length_chk check (char_length(caption) between 1 and 64),
    constraint room_categories_caption_key_length_chk check (char_length(caption_key) between 1 and 64),
    constraint room_categories_version_positive_chk check (version > 0)
);

create unique index room_categories_caption_key_active_uidx
on room_categories (caption_key)
where deleted_at is null;

create table rooms (
    id bigint generated always as identity primary key,
    owner_player_id bigint not null references players(id),
    owner_name text not null,
    name text not null,
    description text not null default '',
    model_name text not null,
    password_hash text null,
    door_mode smallint not null default 0,
    max_users integer not null default 25,
    score integer not null default 0,
    category_id bigint null references room_categories(id),
    trade_mode smallint not null default 0,
    allow_walkthrough boolean not null default false,
    allow_pets boolean not null default true,
    allow_pets_eat boolean not null default false,
    hide_walls boolean not null default false,
    wall_thickness integer not null default 0,
    floor_thickness integer not null default 0,
    chat_mode smallint not null default 0,
    chat_weight smallint not null default 1,
    chat_speed smallint not null default 1,
    chat_distance smallint not null default 50,
    chat_protection smallint not null default 2,
    moderation_mute smallint not null default 0,
    moderation_kick smallint not null default 0,
    moderation_ban smallint not null default 0,
    staff_picked boolean not null default false,
    public_room boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint rooms_owner_name_length_chk check (char_length(owner_name) between 1 and 64),
    constraint rooms_name_length_chk check (char_length(name) between 3 and 25),
    constraint rooms_description_length_chk check (char_length(description) <= 128),
    constraint rooms_model_name_prefix_chk check (model_name like 'model_%'),
    constraint rooms_door_mode_chk check (door_mode between 0 and 3),
    constraint rooms_max_users_chk check (max_users between 1 and 100),
    constraint rooms_trade_mode_chk check (trade_mode between 0 and 2),
    constraint rooms_version_positive_chk check (version > 0)
);

create index rooms_owner_player_id_idx on rooms (owner_player_id) where deleted_at is null;
create index rooms_category_id_idx on rooms (category_id) where deleted_at is null;
create index rooms_score_idx on rooms (score desc) where deleted_at is null;

create table room_tags (
    room_id bigint not null references rooms(id) on delete cascade,
    tag text not null,
    created_at timestamptz not null default now(),
    primary key (room_id, tag),
    constraint room_tags_tag_length_chk check (char_length(tag) between 1 and 32)
);
--rollback drop table if exists room_tags;
--rollback drop table if exists rooms;
--rollback drop table if exists room_categories;
