--liquibase formatted sql
--changeset pixels:pixels-bot-0003-create-room-visits
create table room_visits (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    entered_at timestamptz not null default now()
);
create index room_visits_room_since_idx on room_visits (room_id, entered_at desc);
create index room_visits_player_idx on room_visits (player_id, entered_at desc);
--rollback drop table if exists room_visits;
