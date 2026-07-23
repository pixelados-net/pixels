--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0004-refresh-default-room-doors context:development
insert into room_layouts (name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled)
values
    ('model_a', 105, E'xxxxxxxxxxxx\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxx000000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxx00000000\rxxxxxxxxxxxx\rxxxxxxxxxxxx', 3, 5, 0, 2, 0, true),
    ('model_b', 95, E'xxxxxxxxxxxx\rxxxxx0000000\rxxxxx0000000\rxxxxx0000000\rxxxxx0000000\r000000000000\rx00000000000\rx00000000000\rx00000000000\rx00000000000\rx00000000000\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx', 0, 5, 0, 2, 0, true),
    ('model_c', 37, E'xxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxx000000x\rxxxxx000000x\rxxxx0000000x\rxxxxx000000x\rxxxxx000000x\rxxxxx000000x\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx\rxxxxxxxxxxxx', 4, 7, 0, 2, 0, true)
on conflict (name) where deleted_at is null do update
set tile_size = excluded.tile_size,
    heightmap = excluded.heightmap,
    door_x = excluded.door_x,
    door_y = excluded.door_y,
    door_z = excluded.door_z,
    door_direction = excluded.door_direction,
    club_level = excluded.club_level,
    enabled = excluded.enabled,
    updated_at = now(),
    version = room_layouts.version + 1;
--rollback select 1;
