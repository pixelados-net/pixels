--liquibase formatted sql
--changeset pixels:pixels-progression-0007-daily-rejections
create table player_daily_quest_rejections (
    player_id bigint not null references players(id) on delete cascade,
    offered_on date not null,
    rejected_at timestamptz not null default now(),
    primary key(player_id,offered_on)
);
--rollback drop table if exists player_daily_quest_rejections;
