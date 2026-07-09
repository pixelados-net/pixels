--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0002-default-furniture-items context:development
-- Coordinates target model_a's current heightmap (walkable x=4..11, y=1..13; x=3..11 on y=5),
-- refreshed by pixels-room-seed-development-0003/0004 after this changeset was first written.
-- Ownership is split between player 1 (demo, the room owner) and player 2 (alice) so multi-owner
-- move/pickup authorization (no_rights) can be exercised against real data.
-- The table sits away from the door (3,5) entry column (x=4..5,y=5) so it does not block walk-in.
insert into furniture_items (definition_id, owner_player_id, room_id, x, y, z, rotation)
values
    (1, 1, 1, 7, 10, 0, 0),
    (2, 2, 1, 7, 3, 0, 0),
    (3, 1, 1, 9, 3, 0, 0),
    (4, 2, 1, 7, 7, 0, 0),
    (5, 1, 1, 9, 7, 0, 0)
on conflict do nothing;
--rollback delete from furniture_items where room_id = 1 and definition_id in (1, 2, 3, 4, 5);
