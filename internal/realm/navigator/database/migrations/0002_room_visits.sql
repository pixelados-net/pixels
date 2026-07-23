--liquibase formatted sql

--changeset pixels:pixels-navigator-0002-room-visits
create table navigator_room_visits (
    player_id bigint not null references players(id) on delete cascade,
    room_id bigint not null references rooms(id) on delete cascade,
    visit_count integer not null default 1 check (visit_count > 0),
    first_visited_at timestamptz not null default now(),
    last_visited_at timestamptz not null default now(),
    primary key (player_id, room_id)
);

create index navigator_room_visits_recent_idx on navigator_room_visits(player_id,last_visited_at desc);
create index navigator_room_visits_frequent_idx on navigator_room_visits(player_id,visit_count desc,last_visited_at desc);

--rollback drop index if exists navigator_room_visits_frequent_idx; drop index if exists navigator_room_visits_recent_idx; drop table if exists navigator_room_visits;
