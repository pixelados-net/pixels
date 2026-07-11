--liquibase formatted sql

--changeset pixels:pixels-player-0004-add-player-club
alter table players
    add column club_level smallint not null default 0,
    add column club_expires_at timestamptz null,
    add constraint players_club_level_chk check (club_level between 0 and 2),
    add constraint players_club_expiration_chk check (club_level = 0 or club_expires_at is not null);
--rollback alter table players drop constraint if exists players_club_expiration_chk, drop constraint if exists players_club_level_chk, drop column if exists club_expires_at, drop column if exists club_level;
