--liquibase formatted sql

--changeset pixels:pixels-room-0001-create-room-layouts
create table room_layouts (
    id bigint generated always as identity primary key,
    name text not null,
    tile_size integer not null,
    heightmap text not null,
    door_x integer not null,
    door_y integer not null,
    door_z integer not null default 0,
    door_direction integer not null,
    club_level integer not null default 0,
    enabled boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null,
    version bigint not null default 1,
    constraint room_layouts_name_length_chk check (char_length(name) between 1 and 64),
    constraint room_layouts_name_prefix_chk check (name like 'model_%'),
    constraint room_layouts_tile_size_positive_chk check (tile_size > 0),
    constraint room_layouts_door_x_positive_chk check (door_x >= 0),
    constraint room_layouts_door_y_positive_chk check (door_y >= 0),
    constraint room_layouts_door_direction_positive_chk check (door_direction >= 0),
    constraint room_layouts_club_level_positive_chk check (club_level >= 0),
    constraint room_layouts_version_positive_chk check (version > 0)
);

create unique index room_layouts_name_active_uidx
on room_layouts (name)
where deleted_at is null;
--rollback drop table if exists room_layouts;
