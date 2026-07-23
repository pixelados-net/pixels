--liquibase formatted sql

--changeset pixels:messenger-0002
create table messenger_ignored_players (
    player_id bigint not null references players(id) on delete cascade,
    ignored_player_id bigint not null references players(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (player_id, ignored_player_id),
    constraint messenger_ignored_players_not_self_chk check (player_id <> ignored_player_id)
);

create index messenger_ignored_players_target_idx
on messenger_ignored_players (ignored_player_id);

--rollback drop table if exists messenger_ignored_players;
