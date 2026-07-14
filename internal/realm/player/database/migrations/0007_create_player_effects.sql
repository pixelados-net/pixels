--liquibase formatted sql

--changeset pixels:pixels-player-0007-create-player-effects
alter table players add column active_effect_id integer null;

create table player_effects (
    player_id bigint not null references players(id) on delete cascade,
    effect_id integer not null,
    duration_seconds integer not null,
    activated_at timestamptz null,
    remaining_charges integer not null default 1,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (player_id, effect_id),
    constraint player_effects_effect_positive_chk check (effect_id > 0),
    constraint player_effects_duration_nonnegative_chk check (duration_seconds >= 0),
    constraint player_effects_charges_chk check (remaining_charges between 1 and 99)
);

create index player_effects_expiry_idx on player_effects (activated_at) where activated_at is not null and duration_seconds > 0;

--rollback drop table if exists player_effects;
--rollback alter table players drop column if exists active_effect_id;
