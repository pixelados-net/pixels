--liquibase formatted sql

--changeset pixels:pixels-group-0003-membership-badges
create table social_group_requests (
    group_id bigint not null references social_groups(id) on delete cascade,
    player_id bigint not null references players(id),
    requested_at timestamptz not null default now(),
    primary key(group_id,player_id)
);

create index social_group_requests_player_idx on social_group_requests(player_id,group_id);
create index social_group_requests_group_time_idx on social_group_requests(group_id,requested_at,player_id);

create table player_social_group_preferences (
    player_id bigint primary key references players(id) on delete cascade,
    favorite_group_id bigint null references social_groups(id) on delete set null,
    version bigint not null default 1,
    updated_at timestamptz not null default now(),
    constraint player_social_group_preferences_version_chk check(version > 0)
);

create index player_social_group_favorite_idx on player_social_group_preferences(favorite_group_id) where favorite_group_id is not null;

create table social_group_badge_elements (
    kind smallint not null,
    id integer not null,
    value_a text not null,
    value_b text not null default '',
    enabled boolean not null default true,
    order_num integer not null default 0,
    primary key(kind,id),
    constraint social_group_badge_elements_kind_chk check(kind between 0 and 1),
    constraint social_group_badge_elements_id_chk check(id between 1 and 999)
);

create table social_group_badge_colors (
    family smallint not null,
    id integer not null,
    hex text not null,
    enabled boolean not null default true,
    order_num integer not null default 0,
    primary key(family,id),
    constraint social_group_badge_colors_family_chk check(family between 0 and 2),
    constraint social_group_badge_colors_id_chk check(id between 1 and 999),
    constraint social_group_badge_colors_hex_chk check(hex ~ '^[0-9A-Fa-f]{6}$')
);

create table social_group_badge_parts (
    group_id bigint not null references social_groups(id) on delete cascade,
    ordinal smallint not null,
    kind smallint not null,
    element_id integer not null,
    color_family smallint not null default 0,
    color_id integer not null,
    position integer not null,
    primary key(group_id,ordinal),
    foreign key(kind,element_id) references social_group_badge_elements(kind,id),
    foreign key(color_family,color_id) references social_group_badge_colors(family,id),
    constraint social_group_badge_parts_ordinal_chk check(ordinal between 0 and 4),
    constraint social_group_badge_parts_kind_chk check(kind between 0 and 1),
    constraint social_group_badge_parts_color_family_chk check(color_family = kind),
    constraint social_group_badge_parts_position_chk check(position between 0 and 9),
    constraint social_group_badge_parts_family_chk check(color_id > 0)
);

--rollback drop table if exists social_group_badge_parts; drop table if exists social_group_badge_colors; drop table if exists social_group_badge_elements; drop index if exists player_social_group_favorite_idx; drop table if exists player_social_group_preferences; drop index if exists social_group_requests_group_time_idx; drop index if exists social_group_requests_player_idx; drop table if exists social_group_requests;
