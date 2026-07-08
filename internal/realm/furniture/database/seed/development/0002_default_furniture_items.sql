--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0002-default-furniture-items context:development
insert into furniture_items (definition_id, owner_player_id, room_id, x, y, z, rotation)
values
    (1, 1, 1, 4, 4, 0, 0),
    (2, 1, 1, 3, 3, 0, 0),
    (3, 1, 1, 6, 3, 0, 0),
    (4, 1, 1, 2, 6, 0, 0),
    (5, 1, 1, 5, 6, 0, 0)
on conflict do nothing;
--rollback delete from furniture_items where room_id = 1 and definition_id in (1, 2, 3, 4, 5);
