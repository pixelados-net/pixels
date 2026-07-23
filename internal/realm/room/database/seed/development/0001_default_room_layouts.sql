--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0001-default-room-layouts context:development
insert into room_layouts (name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled)
values
    ('model_a', 104, E'xxxxxxxxxx\rx00000000x\rx00000000x\rx00000000x\rx00000000x\rx00000000x\rx00000000x\rx00000000x\rx00000000x\rxxxxxxxxxx', 1, 1, 0, 2, 0, true),
    ('model_b', 94, E'xxxxxxxxx\rx0000000x\rx0000000x\rx0000000x\rx0000000x\rx0000000x\rx0000000x\rx0000000x\rxxxxxxxxx', 1, 1, 0, 2, 0, true),
    ('model_c', 36, E'xxxxxx\rx0000x\rx0000x\rx0000x\rx0000x\rxxxxxx', 1, 1, 0, 2, 0, true)
on conflict do nothing;
--rollback delete from room_layouts where name in ('model_a', 'model_b', 'model_c');
