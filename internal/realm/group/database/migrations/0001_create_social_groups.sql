--liquibase formatted sql

--changeset pixels:pixels-group-0001-create-social-groups
create table social_groups (
    id bigint generated always as identity primary key,
    name text not null unique,
    badge_code text not null default '',
    created_at timestamptz not null default now(),
    constraint social_groups_name_chk check (char_length(name) between 1 and 64)
);

create table social_group_members (
    group_id bigint not null references social_groups(id) on delete cascade,
    player_id bigint not null references players(id),
    joined_at timestamptz not null default now(),
    primary key (group_id, player_id)
);

create index social_group_members_player_idx on social_group_members(player_id);

--rollback drop index if exists social_group_members_player_idx; drop table if exists social_group_members; drop table if exists social_groups;
