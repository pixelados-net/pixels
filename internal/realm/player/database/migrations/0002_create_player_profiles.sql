--liquibase formatted sql

--changeset pixels:pixels-player-0002-create-player-profiles
create table player_profiles (
    player_id bigint primary key,
    look text not null default '',
    gender text not null default 'M',
    motto text not null default '',
    home_room_id bigint null,
    allow_name_change boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    constraint player_profiles_player_id_fkey foreign key (player_id) references players (id),
    constraint player_profiles_gender_chk check (gender in ('M', 'F')),
    constraint player_profiles_version_positive_chk check (version > 0)
);
--rollback drop table if exists player_profiles;
