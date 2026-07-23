--liquibase formatted sql

--changeset pixels:pixels-camera-0001-captures
create table camera_captures (
    id bigserial primary key,
    capture_uuid uuid not null unique,
    player_id bigint not null references players(id),
    room_id bigint not null references rooms(id),
    kind text not null check (kind in ('photo','thumbnail')),
    storage_key text not null,
    url text not null,
    created_at timestamptz not null default now(),
    consumed_at timestamptz null,
    version bigint not null default 1 check (version > 0)
);
create index camera_captures_pending_idx on camera_captures(player_id, created_at desc) where kind='photo' and consumed_at is null;
create unique index camera_captures_one_pending_idx on camera_captures(player_id) where kind='photo' and consumed_at is null;
--rollback drop table if exists camera_captures;
