--liquibase formatted sql

--changeset pixels:pixels-group-0002-extend-social-groups
alter table social_groups
    add column owner_player_id bigint null references players(id),
    add column description text not null default '',
    add column home_room_id bigint null,
    add column state smallint not null default 0,
    add column can_members_decorate boolean not null default false,
    add column color_a integer not null default 1,
    add column color_b integer not null default 1,
    add column forum_enabled boolean not null default false,
    add column forum_read_policy smallint not null default 0,
    add column forum_post_message_policy smallint not null default 1,
    add column forum_post_thread_policy smallint not null default 1,
    add column forum_moderate_policy smallint not null default 2,
    add column member_count integer not null default 0,
    add column pending_count integer not null default 0,
    add column thread_count integer not null default 0,
    add column post_count integer not null default 0,
    add column updated_at timestamptz not null default now(),
    add column deactivated_at timestamptz null,
    add column version bigint not null default 1,
    add constraint social_groups_description_chk check (char_length(description) <= 254),
    add constraint social_groups_state_chk check (state between 0 and 2),
    add constraint social_groups_color_a_chk check (color_a > 0),
    add constraint social_groups_color_b_chk check (color_b > 0),
    add constraint social_groups_forum_read_chk check (forum_read_policy between 0 and 3),
    add constraint social_groups_forum_post_message_chk check (forum_post_message_policy between 0 and 3),
    add constraint social_groups_forum_post_thread_chk check (forum_post_thread_policy between 0 and 3),
    add constraint social_groups_forum_moderate_chk check (forum_moderate_policy between 0 and 3),
    add constraint social_groups_counts_chk check (member_count >= 0 and pending_count >= 0 and thread_count >= 0 and post_count >= 0),
    add constraint social_groups_version_chk check (version > 0);

alter table social_group_members
    add column role smallint not null default 2,
    add column updated_at timestamptz not null default now(),
    add column version bigint not null default 1,
    add constraint social_group_members_role_chk check (role between 0 and 2),
    add constraint social_group_members_version_chk check (version > 0);

update social_groups groups
set owner_player_id = owners.player_id
from (
    select distinct on (group_id) group_id,player_id
    from social_group_members
    order by group_id,joined_at,player_id
) owners
where groups.id=owners.group_id and groups.owner_player_id is null;

update social_group_members members
set role=0
from social_groups groups
where groups.id=members.group_id and groups.owner_player_id=members.player_id;

update social_groups groups
set member_count=(select count(*) from social_group_members members where members.group_id=groups.id);

create unique index social_groups_owner_active_uidx on social_groups(id,owner_player_id) where deactivated_at is null;
create index social_groups_owner_idx on social_groups(owner_player_id) where deactivated_at is null;
create index social_groups_home_room_idx on social_groups(home_room_id) where deactivated_at is null;
create index social_group_members_player_group_idx on social_group_members(player_id,group_id);
create index social_group_members_group_role_idx on social_group_members(group_id,role,player_id);

--rollback drop index if exists social_group_members_group_role_idx; drop index if exists social_group_members_player_group_idx; drop index if exists social_groups_home_room_idx; drop index if exists social_groups_owner_idx; drop index if exists social_groups_owner_active_uidx; alter table social_group_members drop column version,drop column updated_at,drop column role; alter table social_groups drop column version,drop column deactivated_at,drop column updated_at,drop column post_count,drop column thread_count,drop column pending_count,drop column member_count,drop column forum_moderate_policy,drop column forum_post_thread_policy,drop column forum_post_message_policy,drop column forum_read_policy,drop column forum_enabled,drop column color_b,drop column color_a,drop column can_members_decorate,drop column state,drop column home_room_id,drop column description,drop column owner_player_id;
