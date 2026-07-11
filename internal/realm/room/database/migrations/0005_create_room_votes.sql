--liquibase formatted sql

--changeset pixels:pixels-room-0005-create-room-votes
create table room_votes (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (room_id, player_id)
);

create index room_votes_room_created_idx
    on room_votes (room_id, created_at desc, player_id desc);

--rollback drop table if exists room_votes;
