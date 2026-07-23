--liquibase formatted sql

--changeset pixels:pixels-room-seed-test-0001-test-room-layouts context:test
insert into room_layouts (name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled)
values
    ('model_test', 9, E'xxx\rx0x\rxxx', 1, 1, 0, 2, 0, true)
on conflict do nothing;
--rollback delete from room_layouts where name = 'model_test';
