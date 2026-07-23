--liquibase formatted sql

--changeset pixels:room-games-0015
create table room_game_scores (
    id bigint generated always as identity primary key,
    room_id bigint not null references rooms(id) on delete cascade,
    game_kind text not null check (game_kind in ('banzai','freeze','football','tag','wired')),
    started_at timestamptz not null,
    player_id bigint not null references players(id),
    team smallint not null check (team between 0 and 4),
    score bigint not null,
    team_score bigint not null,
    created_at timestamptz not null default now()
);

create index room_game_scores_room_created_idx on room_game_scores(room_id, created_at desc, id desc);
create index room_game_scores_player_idx on room_game_scores(player_id, created_at desc);

--rollback drop table room_game_scores;
