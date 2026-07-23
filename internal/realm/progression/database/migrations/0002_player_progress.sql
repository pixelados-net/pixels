--liquibase formatted sql
--changeset pixels:pixels-progression-0002-player-progress
create table player_achievement_progress (
    player_id bigint not null references players(id) on delete cascade,
    definition_id bigint not null references achievement_definitions(id) on delete cascade,
    progress bigint not null default 0 check(progress>=0), level integer not null default 0 check(level>=0),
    last_daily_at date null, updated_at timestamptz not null default now(), primary key(player_id,definition_id)
);
create index player_achievement_progress_player_idx on player_achievement_progress(player_id,definition_id);
alter table players add column achievement_score integer not null default 0 check(achievement_score>=0);
--rollback alter table players drop column if exists achievement_score; drop table if exists player_achievement_progress;
