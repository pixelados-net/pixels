--liquibase formatted sql

--changeset pixels:pixels-player-0001-create-players
create table players (
    id bigint generated always as identity primary key,
    username text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    last_login_at timestamptz null,
    last_logout_at timestamptz null,
    last_seen_at timestamptz null,
    constraint players_username_length_chk check (char_length(username) between 1 and 64),
    constraint players_version_positive_chk check (version > 0)
);

create unique index players_username_active_uidx
on players (lower(username))
where deleted_at is null;
--rollback drop table if exists players;
