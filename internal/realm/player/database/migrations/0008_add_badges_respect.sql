--liquibase formatted sql

--changeset pixels:pixels-player-0008-add-badges-respect
create table player_badges (
    player_id bigint not null references players(id),
    code text not null,
    equipped boolean not null default false,
    granted_at timestamptz not null default now(),
    source text not null,
    primary key (player_id, code),
    constraint player_badges_code_chk check (char_length(code) between 1 and 64),
    constraint player_badges_source_chk check (char_length(source) between 1 and 32)
);

create index player_badges_equipped_idx on player_badges(player_id) where equipped;

create table player_respect_totals (
    player_id bigint primary key references players(id),
    received bigint not null default 0 check (received >= 0),
    updated_at timestamptz not null default now()
);

create table player_respect_ledger (
    source_key text primary key,
    player_id bigint not null references players(id),
    amount integer not null check (amount > 0),
    source text not null,
    created_at timestamptz not null default now()
);

--rollback drop table if exists player_respect_ledger; drop table if exists player_respect_totals; drop index if exists player_badges_equipped_idx; drop table if exists player_badges;
