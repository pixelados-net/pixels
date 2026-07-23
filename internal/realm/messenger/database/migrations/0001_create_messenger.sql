--liquibase formatted sql

--changeset pixels:messenger-0001
create table messenger_friendships (
    player_id bigint not null references players(id) on delete cascade,
    friend_player_id bigint not null references players(id) on delete cascade,
    relation smallint not null default 0,
    category_id bigint null,
    created_at timestamptz not null default now(),
    primary key (player_id, friend_player_id),
    constraint messenger_friendships_not_self_chk check (player_id <> friend_player_id),
    constraint messenger_friendships_relation_chk check (relation between 0 and 3)
);

create index messenger_friendships_friend_player_id_idx
on messenger_friendships (friend_player_id);

create table messenger_friend_requests (
    from_player_id bigint not null references players(id) on delete cascade,
    to_player_id bigint not null references players(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (from_player_id, to_player_id),
    constraint messenger_friend_requests_not_self_chk check (from_player_id <> to_player_id)
);

create index messenger_friend_requests_to_player_id_idx
on messenger_friend_requests (to_player_id, created_at desc);

alter table player_profiles
    add column block_friend_requests boolean not null default false,
    add column block_room_invites boolean not null default false,
    add column block_following boolean not null default false;

create index players_username_prefix_idx
on players (lower(username) text_pattern_ops)
where deleted_at is null;

create table messenger_private_messages (
    id bigint generated always as identity primary key,
    from_player_id bigint not null references players(id) on delete cascade,
    to_player_id bigint not null references players(id) on delete cascade,
    message text not null,
    created_at timestamptz not null default now(),
    constraint messenger_private_messages_not_self_chk check (from_player_id <> to_player_id),
    constraint messenger_private_messages_length_chk check (char_length(message) between 1 and 255)
);

create index messenger_private_messages_players_created_idx
on messenger_private_messages (from_player_id, to_player_id, created_at desc);

--rollback drop table if exists messenger_private_messages; drop index if exists players_username_prefix_idx; alter table player_profiles drop column if exists block_following, drop column if exists block_room_invites, drop column if exists block_friend_requests; drop table if exists messenger_friend_requests; drop table if exists messenger_friendships;
