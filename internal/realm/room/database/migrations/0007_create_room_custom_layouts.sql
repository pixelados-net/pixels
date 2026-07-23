--liquibase formatted sql

--changeset pixels:pixels-room-0007-create-room-custom-layouts
create table room_custom_layouts (
    room_id bigint primary key references rooms(id) on delete cascade,
    heightmap text not null,
    door_x smallint not null,
    door_y smallint not null,
    door_direction smallint not null,
    wall_thickness smallint not null default 0,
    floor_thickness smallint not null default 0,
    wall_height smallint not null default -1,
    updated_at timestamptz not null default now(),
    constraint room_custom_layouts_door_direction_chk check (door_direction between 0 and 7),
    constraint room_custom_layouts_wall_thickness_chk check (wall_thickness between -2 and 1),
    constraint room_custom_layouts_floor_thickness_chk check (floor_thickness between -2 and 1),
    constraint room_custom_layouts_wall_height_chk check (wall_height between -1 and 15),
    constraint room_custom_layouts_heightmap_length_chk check (char_length(heightmap) between 1 and 8191)
);

--rollback drop table if exists room_custom_layouts;
