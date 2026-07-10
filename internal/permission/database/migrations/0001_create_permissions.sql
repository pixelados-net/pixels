--liquibase formatted sql

--changeset pixels:pixels-permission-0001-create-permissions
create table permission_groups (
    id bigint generated always as identity primary key,
    name text not null,
    weight integer not null default 0,
    prefix text not null default '',
    prefix_color text not null default '',
    parent_group_id bigint null references permission_groups(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint permission_groups_name_length_chk check (char_length(name) between 1 and 40),
    constraint permission_groups_parent_not_self_chk check (parent_group_id is null or parent_group_id <> id),
    constraint permission_groups_version_positive_chk check (version > 0)
);

create unique index permission_groups_name_active_uidx on permission_groups (name) where deleted_at is null;
create index permission_groups_parent_idx on permission_groups (parent_group_id) where deleted_at is null;

create table permission_group_nodes (
    group_id bigint not null references permission_groups(id) on delete cascade,
    node text not null,
    allowed boolean not null default true,
    primary key (group_id, node),
    constraint permission_group_nodes_length_chk check (char_length(node) between 1 and 160)
);

create table permission_player_groups (
    player_id bigint not null references players(id) on delete cascade,
    group_id bigint not null references permission_groups(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (player_id, group_id)
);

create index permission_player_groups_group_idx on permission_player_groups (group_id, player_id);

create table permission_player_nodes (
    player_id bigint not null references players(id) on delete cascade,
    node text not null,
    allowed boolean not null default true,
    primary key (player_id, node),
    constraint permission_player_nodes_length_chk check (char_length(node) between 1 and 160)
);
--rollback drop table if exists permission_player_nodes;
--rollback drop table if exists permission_player_groups;
--rollback drop table if exists permission_group_nodes;
--rollback drop table if exists permission_groups;
