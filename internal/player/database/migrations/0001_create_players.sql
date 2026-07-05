--liquibase formatted sql

--changeset pixels:pixels-player-0001-create-players
create table players (
    id uuid primary key default gen_random_uuid(),
    username text not null,
    email text null,
    look text not null default '',
    motto text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    last_login_at timestamptz null,
    last_logout_at timestamptz null,
    constraint players_username_length_chk check (char_length(username) between 1 and 64),
    constraint players_email_length_chk check (email is null or char_length(email) <= 254),
    constraint players_version_positive_chk check (version > 0)
);

create unique index players_username_active_uidx
on players (lower(username))
where deleted_at is null;

create unique index players_email_active_uidx
on players (lower(email))
where email is not null and deleted_at is null;
--rollback drop table if exists players;
