--liquibase formatted sql

--changeset pixels:pixels-room-0003-create-room-rights-moderation
alter table rooms
    add constraint rooms_moderation_mute_chk check (moderation_mute between 0 and 2),
    add constraint rooms_moderation_kick_chk check (moderation_kick between 0 and 2),
    add constraint rooms_moderation_ban_chk check (moderation_ban between 0 and 2);

create table room_rights (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    granted_by_player_id bigint not null references players(id),
    created_at timestamptz not null default now(),
    primary key (room_id, player_id)
);

create index room_rights_player_id_idx on room_rights (player_id, room_id);

create table room_mutes (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    ends_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (room_id, player_id)
);

create index room_mutes_room_active_idx on room_mutes (room_id, ends_at desc);
create index room_mutes_player_active_idx on room_mutes (player_id, ends_at desc);

create table room_bans (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    ends_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (room_id, player_id)
);

create index room_bans_room_active_idx on room_bans (room_id, ends_at desc);
create index room_bans_player_active_idx on room_bans (player_id, ends_at desc);

create table room_rights_audit (
    id bigint generated always as identity primary key,
    room_id bigint not null references rooms(id),
    player_id bigint not null references players(id),
    actor_kind text not null default 'player',
    actor_id bigint null references players(id),
    action text not null,
    created_at timestamptz not null default now(),
    constraint room_rights_audit_actor_kind_chk check (actor_kind in ('player', 'admin', 'system')),
    constraint room_rights_audit_action_chk check (action in ('granted', 'revoked', 'revoked_all', 'relinquished'))
);

create index room_rights_audit_room_idx on room_rights_audit (room_id, id desc);
create index room_rights_audit_player_idx on room_rights_audit (player_id, id desc);
create index room_rights_audit_actor_idx on room_rights_audit (actor_id, id desc);

create table room_moderation_actions (
    id bigint generated always as identity primary key,
    room_id bigint not null references rooms(id),
    target_player_id bigint not null references players(id),
    actor_kind text not null default 'player',
    actor_id bigint null references players(id),
    action_type text not null,
    duration_seconds bigint null,
    expires_at timestamptz null,
    created_at timestamptz not null default now(),
    constraint room_moderation_actions_actor_kind_chk check (actor_kind in ('player', 'admin', 'system')),
    constraint room_moderation_actions_action_chk check (action_type in ('kick', 'mute', 'unmute', 'ban', 'unban')),
    constraint room_moderation_actions_duration_chk check (duration_seconds is null or duration_seconds > 0)
);

create index room_moderation_actions_room_idx on room_moderation_actions (room_id, id desc);
create index room_moderation_actions_target_idx on room_moderation_actions (target_player_id, id desc);
create index room_moderation_actions_actor_idx on room_moderation_actions (actor_id, id desc);
create index room_moderation_actions_type_idx on room_moderation_actions (action_type, id desc);

--rollback drop table if exists room_moderation_actions;
--rollback drop table if exists room_rights_audit;
--rollback drop table if exists room_bans;
--rollback drop table if exists room_mutes;
--rollback drop table if exists room_rights;
--rollback alter table rooms drop constraint if exists rooms_moderation_ban_chk;
--rollback alter table rooms drop constraint if exists rooms_moderation_kick_chk;
--rollback alter table rooms drop constraint if exists rooms_moderation_mute_chk;
