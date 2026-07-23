--liquibase formatted sql
--changeset pixels:pixels-progression-0001-achievements
create table achievement_definitions (
    id bigint generated always as identity primary key,
    name text not null unique, category text not null, subcategory text not null default '', trigger_key text not null,
    visible boolean not null default true, enabled boolean not null default true,
    created_at timestamptz not null default now(), updated_at timestamptz not null default now(),
    version bigint not null default 1 check(version>0),
    constraint achievement_definition_name_chk check(char_length(name) between 1 and 70),
    constraint achievement_definition_trigger_chk check(char_length(trigger_key) between 1 and 120)
);
create index achievement_definitions_trigger_idx on achievement_definitions(trigger_key,id) where enabled;
create table achievement_levels (
    definition_id bigint not null references achievement_definitions(id) on delete cascade,
    level integer not null check(level>0), progress_needed bigint not null check(progress_needed>0),
    reward_currency_type integer not null default 0, reward_amount bigint not null default 0 check(reward_amount>=0),
    score_points integer not null default 10 check(score_points>=0), primary key(definition_id,level)
);
--rollback drop table if exists achievement_levels; drop table if exists achievement_definitions;
