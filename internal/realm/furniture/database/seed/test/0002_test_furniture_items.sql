--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-test-0002-test-furniture-items context:test
insert into furniture_items (definition_id, owner_player_id, room_id, x, y, z, rotation)
values
    (2, 1, null, null, null, null, 0)
on conflict do nothing;
--rollback delete from furniture_items where owner_player_id = 1 and room_id is null and definition_id = 2;
