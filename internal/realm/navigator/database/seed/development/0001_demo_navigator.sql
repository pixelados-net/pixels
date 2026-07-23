--liquibase formatted sql

--changeset pixels:pixels-navigator-seed-development-0001-demo-navigator context:development
insert into room_favorites (player_id, room_id)
values
    (1, 1),
    (1, 2),
    (2, 1)
on conflict do nothing;

insert into navigator_preferences (player_id)
values
    (1),
    (2),
    (3),
    (4)
on conflict do nothing;

insert into navigator_lifted_rooms (room_id, area_id, image, caption, order_num)
values
    (1, 0, 'navigator/lobby.png', 'Pixels Lobby', 1)
on conflict do nothing;
--rollback delete from navigator_lifted_rooms where room_id = 1;
--rollback delete from navigator_preferences where player_id in (1, 2, 3, 4);
--rollback delete from room_favorites where player_id in (1, 2);
