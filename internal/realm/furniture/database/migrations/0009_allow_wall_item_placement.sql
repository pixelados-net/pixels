--liquibase formatted sql

--changeset pixels:pixels-furniture-0009-allow-wall-item-placement
alter table furniture_items
drop constraint furniture_items_room_placement_chk;

update furniture_items
set x = null,
    y = null,
    z = null,
    rotation = 0
where room_id is not null
  and wall_position is not null;

alter table furniture_items
add constraint furniture_items_room_placement_chk check (
    (room_id is null and x is null and y is null and z is null and wall_position is null)
    or
    (
        room_id is not null
        and (
            (x is not null and y is not null and z is not null and wall_position is null)
            or
            (x is null and y is null and z is null and wall_position is not null)
        )
    )
);

--rollback alter table furniture_items drop constraint furniture_items_room_placement_chk; update furniture_items set x = 0, y = 0, z = 0 where room_id is not null and wall_position is not null; alter table furniture_items add constraint furniture_items_room_placement_chk check ((room_id is null and x is null and y is null and z is null) or (room_id is not null and x is not null and y is not null and z is not null));
