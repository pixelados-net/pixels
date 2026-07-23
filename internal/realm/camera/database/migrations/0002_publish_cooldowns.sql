--liquibase formatted sql

--changeset pixels:pixels-camera-0002-publish-cooldowns
create table camera_publish_cooldowns (
    player_id bigint primary key references players(id),
    last_published_at timestamptz not null
);
--rollback drop table if exists camera_publish_cooldowns;
