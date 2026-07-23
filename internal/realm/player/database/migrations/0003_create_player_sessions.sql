--liquibase formatted sql

--changeset pixels:pixels-player-0003-create-player-sessions
create table player_sessions (
    id bigint generated always as identity primary key,
    player_id bigint not null,
    connection_kind text not null,
    started_at timestamptz not null default now(),
    authenticated_at timestamptz null,
    ended_at timestamptz null,
    disconnect_code text null,
    ip_hash text null,
    machine_id_hash text null,
    user_agent text null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint player_sessions_player_id_fkey foreign key (player_id) references players (id),
    constraint player_sessions_connection_kind_length_chk check (char_length(connection_kind) between 1 and 32),
    constraint player_sessions_disconnect_code_length_chk check (disconnect_code is null or char_length(disconnect_code) <= 64),
    constraint player_sessions_time_order_chk check (ended_at is null or ended_at >= started_at)
);

create index player_sessions_player_id_started_at_idx
on player_sessions (player_id, started_at desc);

create index player_sessions_active_idx
on player_sessions (player_id)
where ended_at is null;
--rollback drop table if exists player_sessions;
