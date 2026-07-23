--liquibase formatted sql
--changeset pixels:pixels-progression-0003-talents
create table talent_track_levels (
    track text not null, level integer not null check(level>0), requirements jsonb not null default '[]'::jsonb,
    reward_items bigint[] not null default '{}', reward_perks text[] not null default '{}', reward_badges text[] not null default '{}',
    primary key(track,level), constraint talent_track_chk check(char_length(track) between 1 and 40)
);
create table player_talent_levels (
    player_id bigint not null references players(id) on delete cascade, track text not null,
    level integer not null default 0 check(level>=0), updated_at timestamptz not null default now(), primary key(player_id,track)
);
--rollback drop table if exists player_talent_levels; drop table if exists talent_track_levels;
