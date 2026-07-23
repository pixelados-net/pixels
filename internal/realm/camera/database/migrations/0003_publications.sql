--liquibase formatted sql

--changeset pixels:pixels-camera-0003-publications
create table camera_publications (
    id bigserial primary key,
    capture_id bigint not null unique references camera_captures(id),
    player_id bigint not null references players(id),
    room_id bigint not null references rooms(id),
    url text not null,
    created_at timestamptz not null default now(),
    removed_at timestamptz null,
    removed_reason text null
);
create index camera_publications_active_idx on camera_publications(created_at desc) where removed_at is null;
--rollback drop table if exists camera_publications;
